package main

import (
	"fmt"
	"git.sr.ht/~justinsantoro/gemtext"
	"git.sr.ht/~justinsantoro/gemtext/ast"
	"html/template"
	"io"
	"os"
	"strings"
)

// "/en/posts/xxxxx.gmi" -> "/en/posts/xxxxx.html"
func urlReplace(link *ast.Link) {
	if strings.HasPrefix(link.Url, "/") && strings.HasSuffix(link.Url, ".gmi") {
		link.Url = strings.TrimSuffix(link.Url, ".gmi") + ".html"
	}
}

func line2html(line ast.Line) (string, bool) {
	switch v := line.(type) {
	case *ast.EmptyLine:
		return "<br/>", false
	case *ast.Text:
		return fmt.Sprintf("<p>%s</p>", string(v.Bytes())), false
	case *ast.Link:
		urlReplace(v)
		label := v.Label
		if label == "" {
			label = v.Url
		}
		return fmt.Sprintf(`<p><a href="%s">%s</a></p>`, v.Url, label), false
	case *ast.Heading:
		if v.Level == 1 {
			return fmt.Sprintf("<h1>%s</h1>", string(v.Bytes())), false
		} else if v.Level == 2 {
			return fmt.Sprintf("<h2>%s</h2>", string(v.Bytes())), false
		} else {
			return fmt.Sprintf("<h3>%s</h3>", string(v.Bytes())), false
		}
	case *ast.ListItem:
		return fmt.Sprintf("    <li>%s</li>", string(v.Bytes())), true
	case *ast.Blockquote:
		return fmt.Sprintf("<blockquote><p>%s</p></blockquote>", string(v.Bytes())), false
	case *ast.Preformatted:
		return fmt.Sprintf("<pre>%s</pre>", string(v.Bytes())), false
	}
	return "", false
}

func main() {
	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Error open file:", err)
		os.Exit(1)
	}

	text, err := io.ReadAll(in)
	if err != nil {
		fmt.Println("Error read file:", err)
		os.Exit(1)
	}

	lines := gemtext.Parse(text)
	tmpl, err := template.New("page").Parse(`<!DOCTYPE html>
	<html>
	  <head>
	  	<meta name="generator" content="gem2site">
		<meta charset="utf-8">
    	<link href="./site.css" rel="stylesheet"/>
		<title>
		{{ .Title }}
		</title>
		</head>
	  <body>
	  	<main>
		<article>
	  	{{ .Content }}
		</article>
		</main>
	  </body>
	</html>`)
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
		Title   string
		Content template.HTML
	}{
		Title:   title,
		Content: template.HTML(content.String()),
	}

	err = tmpl.Execute(&page, data)
	if err != nil {
		fmt.Println("Execute template failed:", err)
		os.Exit(1)
	}

	fmt.Println(page.String())
}
