package main

import (
	"flag"
	"fmt"
	"git.sr.ht/~justinsantoro/gemtext"
	"git.sr.ht/~justinsantoro/gemtext/ast"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const GMI_EXT string = ".gmi"

// "/en/posts/xxxxx.gmi" -> "/en/posts/xxxxx.html"
func urlReplace(link *ast.Link) {
	if strings.HasPrefix(link.Url, "/") && strings.HasSuffix(link.Url, GMI_EXT) {
		link.Url = strings.TrimSuffix(link.Url, GMI_EXT) + ".html"
	}
}

func es(text string) string {
	return template.HTMLEscapeString(text)
}

func highlight(writer io.Writer, source, lang, style string) error {
	l := lexers.Get(lang)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	f := html.New(html.Standalone(false))

	s := styles.Get(style)
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}
	return f.Format(writer, s, it)
}

func line2html(line ast.Line) (string, bool) {
	switch v := line.(type) {
	case *ast.EmptyLine:
		return `<div class="empty-line"></div>`, false
	case *ast.Text:
		return fmt.Sprintf("<p>%s</p>", es(string(v.Bytes()))), false
	case *ast.Link:
		urlReplace(v)
		label := v.Label
		if label == "" {
			label = v.Url
		}
		return fmt.Sprintf(`<p><a href="%s">%s</a></p>`, v.Url, es(label)), false
	case *ast.Heading:
		if v.Level == 1 {
			return fmt.Sprintf("<h1>%s</h1>", es(string(v.Bytes()))), false
		} else if v.Level == 2 {
			return fmt.Sprintf("<h2>%s</h2>", es(string(v.Bytes()))), false
		} else {
			return fmt.Sprintf("<h3>%s</h3>", es(string(v.Bytes()))), false
		}
	case *ast.ListItem:
		return fmt.Sprintf("    <li>%s</li>", es(string(v.Bytes()))), true
	case *ast.Blockquote:
		return fmt.Sprintf("<blockquote><p>%s</p></blockquote>", es(string(v.Bytes()))), false
	case *ast.Preformatted:
		preText := string(v.Bytes())
		if alt := v.AltText; alt != "" && *codeHighlightStyle != "" {
			var code strings.Builder
			// err := quick.Highlight(&code, preText, alt, "html", *codeHighlightStyle)
			err := highlight(&code, preText, alt, *codeHighlightStyle)
			if err != nil {
				fmt.Printf("Error while highlight code: %s\n", err)
			}
			return code.String(), false
		}
		return fmt.Sprintf("<pre>%s</pre>", preText), false
	}
	return "", false
}

func outputFile(p string, content []byte) error {
	err := os.MkdirAll(filepath.Dir(p), 0755)
	if err != nil {
		fmt.Println("failed to create dir", err)
		os.Exit(1)
	}
	file, err := os.Create(p)
	if err != nil {
		fmt.Println("failed to write", p, err)
		os.Exit(1)
	}

	file.Write(content)
	return nil
}

const defaultTemplate string = `<!DOCTYPE html>
<html>
  <head>
    <meta name="generator" content="gem2site">
    <meta charset="utf-8">
    <style>
      div.empty-line {
        height: 0.5em;
        margin:0;
        padding:0;
      }
    </style>
	{{ if .ExternalCssFile }}
      <link href="{{.ExternalCssFile}}" rel="stylesheet"/>
	{{ else }}
    <style>
	{{ .DefaultCss }}
	</style>
	{{ end }}

    <title>{{ .Title }}</title>
  </head>
  <body>
    <main>
      <article>
        {{ .Content }}
      </article>
    </main>
  </body>
</html>`

const defaultCss string = `body {
  color: #171717;
  font-family: 'Garamond', Georgia, serif, 'Noto Color Emoji', 'Apple Color Emoji', 'Segoe UI Emoji';
}

article {
  margin: 0 auto;
  max-width: 720px;
  line-height: 1.3;
}

h1,h2,h3 {
  color: #ba3925;
  text-rendering: optimizeLegibility;
  font-family: "Open Sans", sans-serif;
}

pre {
    background-color: #eee;
    padding: 0.5rem 0.5rem;
    margin: 25px 0;
    max-width: 100%;
    overflow-x: auto;
    border: solid 1px lightgray;
    border-radius: 4px;
	font-size: 14px;
}

p {
  text-align: justify;
  margin: 0.25rem;
}

article > p:first-child > a {
    text-decoration: none;
    font-size: 200%;
    font-weight: bold;
    color: black;
}
`

func processFile(p string, tmplString string, externalCssFile string) []byte {
	bytes, err := os.ReadFile(p)
	if err != nil {
		fmt.Println("Error read file:", err)
		os.Exit(1)
	}

	if path.Ext(p) != GMI_EXT {
		return bytes
	}

	lines := gemtext.Parse(bytes)
	tmpl, err := template.New("page").Parse(tmplString)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		os.Exit(1)
	}

	var content strings.Builder
	inList := false
	title := ""
	for _, line := range lines {
		if title == "" {
			h1, ok := line.(*ast.Heading)
			if ok && h1.Level == 1 {
				title = string(h1.Bytes())
			}
		}

		text, isItem := line2html(line)
		// entering and leaving a list
		if !inList && isItem {
			content.WriteString("<ul>\n")
		} else if inList && !isItem {
			content.WriteString("</ul>\n")
		}
		content.WriteString(text)
		content.WriteString("\n")
		inList = isItem
	}

	if title == "" {
		title = "Page"
	}

	var page strings.Builder
	data := struct {
		Title           string
		Content         template.HTML
		ExternalCssFile string
		DefaultCss      template.CSS
	}{
		Title:           title,
		Content:         template.HTML(content.String()),
		ExternalCssFile: externalCssFile,
		DefaultCss:      template.CSS(defaultCss),
	}

	err = tmpl.Execute(&page, data)
	if err != nil {
		fmt.Println("Execute template failed:", err)
		os.Exit(1)
	}

	return []byte(page.String())
}

var externalTmpl = flag.String("tmpl", "", "path of external template file")
var externalCss = flag.String("css", "", "use external css file instead of internal default css. value should be urlPath of the external css file UrlPath, which will be used as value of href attribute of <link> element in head.")
var dumpTmpl = flag.Bool("dump", false, "print default template and css, then exit")
var codeHighlightStyle = flag.String("hl", "github", "code highlighting style, see https://github.com/alecthomas/chroma for details.  Default is 'github', pass empty string to disable code Highlighting.")

func main() {
	flag.Parse()
	if *dumpTmpl {
		fmt.Println(defaultTemplate)
		fmt.Println()
		fmt.Println(defaultCss)
		return
	}

	if flag.NArg() != 2 {
		fmt.Printf("Usage: %s [flags] <source> <dest>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	src := flag.Arg(0)
	dest := flag.Arg(1)

	tmpl := defaultTemplate
	if *externalTmpl != "" {
		tmplBytes, err := os.ReadFile(*externalTmpl)
		if err != nil {
			fmt.Println("Cannot read alternative template: ", err)
			os.Exit(1)
		}
		tmpl = string(tmplBytes)
	}

	filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(src, p)
			out := filepath.Join(dest, rel)
			if path.Ext(out) == GMI_EXT {
				out = strings.TrimSuffix(out, GMI_EXT) + ".html"
			}
			outStat, err := os.Stat(out)
			if err != nil && !os.IsNotExist(err) {
				fmt.Printf("Unable to check stat of target file %s, %s\n", out, err)
				return err
			}
			if err == nil && outStat.ModTime().After(info.ModTime()) {
				fmt.Printf("Skip target %s as already up to date.\n", out)
				return nil
			}
			content := processFile(p, tmpl, *externalCss)
			outputFile(out, content)
		}
		return nil
	})
}
