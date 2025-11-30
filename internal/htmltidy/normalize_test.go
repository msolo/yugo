package htmltidy

import (
	"slices"
	"strings"
	"testing"

	"github.com/ianbruene/go-difflib/difflib"
)

func testNormalizer(t *testing.T, in string, expected string) {
	out, err := NormalizeHTML(in)
	if err != nil {
		t.Fatalf("NormalizeHTML error: %s", err)
	}
	if out != expected {
		diff := strDiff(expected, out)
		t.Fatalf("output diff:\n--- expected ---\n%s\n--- output ---\n%s\n\n%s", expected, out, diff)
	}

	// Check that the output is stable.
	in = out
	out, err = NormalizeHTML(in)
	if err != nil {
		t.Fatalf("re-NormalizeHTML error: %s", err)
	}
	if out != expected {
		diff := strDiff(expected, out)
		t.Fatalf("unstable output diff:\n--- expected ---\n%s\n--- output ---\n%s\n\n%s", expected, out, diff)
	}
}

func strDiff(expected, out string) string {
	// Pick a character to help spot whitespace issues in black and white.
	const visualSpace = "⋅" // "∘" "∙"
	const visualTab = "→   "

	expected = strings.ReplaceAll(expected, " ", visualSpace)
	out = strings.ReplaceAll(out, " ", visualSpace)
	expected = strings.ReplaceAll(expected, "\t", visualTab)
	out = strings.ReplaceAll(out, "\t", visualTab)

	diff, err := difflib.GetUnifiedDiffString(difflib.LineDiffParams{
		A:        slices.Collect(strings.Lines(expected)),
		FromFile: "expected",
		B:        slices.Collect(strings.Lines(out)),
		ToFile:   "out",
	})
	if err != nil {
		panic(err)
	}
	return diff
}

func TestDoctypePreserved(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestInlineFormatting(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<p>
   <em>This   is</em>
   some   <i>in</i>line text
</p>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <p>
            <em>This is</em> some <i>in</i>line text
        </p>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestInlineFormatting2(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<p>
   <em>This   is</em>some   <i>in</i>line text
</p>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <p>
            <em>This is</em>some <i>in</i>line text
        </p>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestBlockFormatting(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
    <body>
            <p>
                <strong>Import File</strong> allows you to import several different special files that you have saved on your device.            </p>

    </body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <p>
            <strong>Import File</strong> allows you to import several different special files that you have saved on your device.
        </p>
    </body>
</html>
`

	testNormalizer(t, in, expected)
}

func TestBlockFormatting2(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
    <body>
<p>For instance:
<code>blinds up</code></p>
    </body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <p>
            For instance: <code>blinds up</code>
        </p>
    </body>
</html>
`

	testNormalizer(t, in, expected)
}

func TestInlineTextNormalization(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<p>
   This   is
   some   <i>in</i>line text
</p>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <p>
            This is some <i>in</i>line text
        </p>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestVoidElements(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<img src="x.png">
<br><br><br>
<hr>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <img src="x.png">
        <br>
        <br>
        <br>
        <hr>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestWhitespacePreserved(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<pre>
    line1
    line2
</pre>
<code>   a   b   c   </code>
<textarea>
   hello
     world
</textarea>
<div>
<script>
    if (true) {
        console.log("hi");
    }
</script>
</div>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <pre>
    line1
    line2
</pre>
        <code> a b c </code>
        <textarea>
   hello
     world
</textarea>
        <div>
            <script>
    if (true) {
        console.log("hi");
    }
</script>
        </div>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestBrowserHoist(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <div>
            <p>
                <div>
                </div>
                2. This <a href="#"><b>should</b></a>  be inline and normal, but because of html parsing, this gets hoisted.
            </p>
            <p>
                3. This <a href="#"><b>should</b></a>  be inline and normal.  As
								should this.
            </p>
        </div>
    </body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <div>
            <p>
            </p>
            <div>
            </div>
            2. This <a href="#"><b>should</b></a> be inline and normal, but because of html parsing, this gets hoisted.
            <p>
            </p>
            <p>
                3. This <a href="#"><b>should</b></a> be inline and normal. As should this.
            </p>
        </div>
    </body>
</html>
`

	testNormalizer(t, in, expected)
}

func TestStylePreserved(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
<style>
    body { color: red; }
</style>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <style>
    body { color: red; }
</style>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestHeuristics(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
<head></head>
<body>
        <header>
<a href="/index.html"><img src="/logo.svg" style="height: 4rem;"></a>
            <div class="spacer">
            </div>
            <nav>
                <ul>
                    <li>
<a href="/" class="">Home</a>                    </li>
</nav>
</ul>
</header>
</body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <header>
            <a href="/index.html"><img src="/logo.svg" style="height: 4rem;"></a>
            <div class="spacer">
            </div>
            <nav>
                <ul>
                    <li>
                        <a href="/" class="">Home</a>
                    </li>
                </ul>
            </nav>
        </header>
    </body>
</html>
`

	testNormalizer(t, in, expected)
}

func TestInlinePre(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <pre>    <i>line1</i><b>bold</b> is a menace
    line1 <b>  bold  </b> is a menace
    line2
</pre>
    </body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
    </head>
    <body>
        <pre>
    <i>line1</i><b>bold</b> is a menace
    line1 <b>  bold  </b> is a menace
    line2
</pre>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}

func TestNormalizeHTML(t *testing.T) {
	in := `<!DOCTYPE html>
<html>
    <head>
        <style>
            p {
                border: 1px solid black;
            }
    </style>
		</head>
    <body>
        <p>
            simple text
        </p>
        <pre>
    line1
    line2
</pre>
        <!-- This second pre gets normalized to the first. -->
        <pre>    line1
    line2
</pre>
        <pre>    <i>line1</i><b>bold</b> is a menace
    line1 <b>  bold  </b> is a menace
    line2
</pre>

        <pre><code>   expect   triple   spaces   </code></pre>
				<!-- in textarea, pre the first CR/LF is ignored, but a second is relevent and must be preserved. -->
        <textarea>

   hello
     world
</textarea>
        <textarea>
   hello
     world
</textarea>
        <hr>
        <div>
            <script>
    if (true) {
        console.log("hi");
    }
</script>
            <p>
                1. This <a href="#"><b>should</b></a> be inline.
            </p>
            <p>
                <div>
                </div>
                2. This <a href="#"><b>should</b></a>  be inline and normal, but because of html parsing, this gets hoisted.
            </p>
            <p>
                3. This <a href="#"><b>should</b></a>  be inline and normal.  As
								should this.
            </p>
        </div>
    </body>
</html>
`
	expected := `<!DOCTYPE html>
<html>
    <head>
        <style>
            p {
                border: 1px solid black;
            }
    </style>
    </head>
    <body>
        <p>
            simple text
        </p>
        <pre>
    line1
    line2
</pre>
        <!-- This second pre gets normalized to the first. -->
        <pre>
    line1
    line2
</pre>
        <pre>
    <i>line1</i><b>bold</b> is a menace
    line1 <b>  bold  </b> is a menace
    line2
</pre>
        <pre><code>   expect   triple   spaces   </code></pre>
        <!-- in textarea, pre the first CR/LF is ignored, but a second is relevent and must be preserved. -->
        <textarea>
   hello
     world
</textarea>
        <textarea>
   hello
     world
</textarea>
        <hr>
        <div>
            <script>
    if (true) {
        console.log("hi");
    }
</script>
            <p>
                1. This <a href="#"><b>should</b></a> be inline.
            </p>
            <p>
            </p>
            <div>
            </div>
            2. This <a href="#"><b>should</b></a> be inline and normal, but because of html parsing, this gets hoisted.
            <p>
            </p>
            <p>
                3. This <a href="#"><b>should</b></a> be inline and normal. As should this.
            </p>
        </div>
    </body>
</html>
`
	testNormalizer(t, in, expected)
}
