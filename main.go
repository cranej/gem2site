package main

import (
	"fmt"
	"git.sr.ht/~justinsantoro/gemtext"
	"git.sr.ht/~justinsantoro/gemtext/ast"
	"io"
	"os"
)

func line2html(line ast.Line) (string, bool) {
	switch v := line.(type) {
	case *ast.EmptyLine:
		return "<br/>", false
	case *ast.Text:
		return fmt.Sprintf("<p>%s</p>", string(v.Bytes())), false
	case *ast.Link:
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
	// TODO: use first level 1 heading as title
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

	inList := false
	for _, line := range lines {
		text, isItem := line2html(line)
		// entering and leaving a list
		if !inList && isItem {
			fmt.Println("<ul>")
		} else if inList && !isItem {
			fmt.Println("</ul>")
		}
		fmt.Println(text)
		inList = isItem
	}

	fmt.Println("</body></html>")
}
