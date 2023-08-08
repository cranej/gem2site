package main

import (
	"fmt"
	"git.sr.ht/~justinsantoro/gemtext"
	"git.sr.ht/~justinsantoro/gemtext/ast"
	"io"
	"os"
	"strings"
)

func line2html(line ast.Line) string {
	switch v := line.(type) {
	case *ast.EmptyLine:
		return "<br/>"
	case *ast.Text:
		return fmt.Sprintf("<p>%s</p>", string(v.Bytes()))
	case *ast.Link:
		label := v.Label
		if label == "" {
			label = v.Url
		}
		return fmt.Sprintf(`<p><a href="%s">%s</a></p>`, v.Url, label)
	case *ast.Heading:
		if v.Level == 1 {
			return fmt.Sprintf("<h1>%s</h1>", string(v.Bytes()))
		} else if v.Level == 2 {
			return fmt.Sprintf("<h2>%s</h2>", string(v.Bytes()))
		} else {
			return fmt.Sprintf("<h3>%s</h3>", string(v.Bytes()))
		}
	case *ast.ListItem:
		panic("shoudld be here")
	case *ast.Blockquote:
		return fmt.Sprintf("<blockquote><p>%s</p></blockquote>", string(v.Bytes()))
	case *ast.Preformatted:
		return fmt.Sprintf("<pre>%s</pre>", string(v.Bytes()))
	}
	return ""
}

func listLines2html(items []ast.Line) string {
	var b strings.Builder
	b.WriteString("<ul>\n")
	for _, l := range items {
		itemText := fmt.Sprintf("    <li>%s</li>\n", string(l.Bytes()))
		b.WriteString(itemText)
	}
	b.WriteString("</ul>")
	return b.String()
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
	// TODO: title
	fmt.Println(`<!DOCTYPE html>
	<html>
	  <head>
	  	<meta name="generator" content="gem2site">
		<meta charset="utf-8">
		<title>
			cranej&#39;s Webpage
		</title>
		</head>
	  <body>`)

	i := 0
	for i < len(lines) {
		_, isItem := lines[i].(*ast.ListItem)
		if isItem {
			start := i
			i++
			for i < len(lines) {
				_, isItem = lines[i].(*ast.ListItem)
				if !isItem {
					break
				}
				i++
			}

			fmt.Println(listLines2html(lines[start:i]))
		} else {
			fmt.Println(line2html(lines[i]))
			i++
		}
	}

	fmt.Println("</body></html>")
}
