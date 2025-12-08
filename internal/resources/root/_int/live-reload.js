// Implement a simple live-reload trigger.
(function () {
  var socket;
  var reconnectAttempts = 0;
  var maxReconnectAttempts = 5;

  function connectWebSocket() {
    socket = new WebSocket('ws://' + window.location.host + '/_int/live-reload.ws');

    socket.onopen = function () {
      // Treat a reconnect as an implicit reload.
      if (reconnectAttempts > 0) {
        location.reload()
      }
      reconnectAttempts = 0;  // Reset reconnection attempts on success
    };

    socket.onclose = function (event) {
      // console.info("live-reload WebSocket connection closed.");
      if (reconnectAttempts < maxReconnectAttempts) {
        // Sanjay backoff our reconnects.
        const timeout = Math.trunc(1000 * Math.pow(1.414, reconnectAttempts++));
        console.info(`live-reload WebSocket reconnecting (Attempt ${reconnectAttempts} in ${timeout}ms)...`);
        setTimeout(connectWebSocket, timeout);
      } else {
        showReloadRequiredMessage();
      }
    };

    socket.onerror = function (error) {
      // Somehow WebSockets already log to the console, so no need to double log.
      socket.close();  // Close the connection on error - this may be redundant.
    };

    socket.onmessage = function (event) {
      if (event.data === "reload") {
        location.reload();
      }
    };
  }

  // Display the overlay when the server goes away
  function showReloadRequiredMessage() {
    var overlay = document.createElement('div');
    overlay.style.position = 'fixed';
    overlay.style.top = '0';
    overlay.style.left = '0';
    overlay.style.width = '100vw';
    overlay.style.height = '100vh';
    overlay.style.background = "rgba(0,0,0,0.25)";
    overlay.style.backdropFilter = "blur(5px)";
    overlay.style.color = '#fff';
    overlay.style.display = 'flex';
    overlay.style.justifyContent = 'center';
    overlay.style.alignItems = 'center';
    overlay.innerHTML = `<span style="border-radius: .5em; padding: 0.5em; background: #66666667; font-size: 2em; font-weight: bold;">
    Reload required. The server went away.
    </span>`
    document.body.appendChild(overlay);
  }

  // Start the WebSocket connection when the page is loaded, this ensures
  // that it reconnects when the browswer back button is used.
  window.onload = function () {
    connectWebSocket();
  };

})();