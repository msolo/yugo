package serve

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/msolo/yugo/internal/build"
	"golang.org/x/net/websocket"
)

func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		h.ServeHTTP(w, r)
	})
}

func Run(opts *build.Options) {
	builder := func() {
		build.Run(opts)
	}

	var clients sync.Map
	// WebSocket handler for live-reload
	lrHandler := websocket.Handler(func(ws *websocket.Conn) {
		clients.Store(ws, ws)
		defer func() {
			clients.Delete(ws)
			_ = ws.Close()
		}()

		// Keep reading messages until disconnect
		for {
			var dummy string
			if err := websocket.Message.Receive(ws, &dummy); err != nil {
				return
			}
		}
	})

	http.Handle("/", noCache(http.FileServer(http.Dir(opts.OutDir()))))
	http.Handle("/_int/live-reload.ws", noCache(lrHandler))

	addr := fmt.Sprintf("%s:%d", opts.Host(), opts.Port())
	go func() {
		fmt.Println("Serving at http://" + addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal(err)
		}
	}()

	builder()

	// Rebuild on changes
	debounce := 750 * time.Millisecond
	watcher, err := NewWatcher([]string{opts.SiteDir()}, debounce)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = watcher.Close()
	}()

	watcher.Start()

	for {
		<-watcher.Events()
		fmt.Println("🔄 Change detected — rebuilding...")
		builder()

		if opts.LiveReload() {
			// Send a reload message to all WS clients
			clients.Range(func(_, v any) bool {
				ws := v.(*websocket.Conn)
				// Most errors are likely to be about the client going away
				// which is fine to ignore given the target use case.
				_ = websocket.Message.Send(ws, "reload")
				return true
			})
		}
	}
}
