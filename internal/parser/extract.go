// internal/parser/extract.go
package parser

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// Extract cleans <body> and returns (title, content, words)
func Extract(htmlBody []byte, maxWords int) (string, string, int) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBody))
	if err != nil {
		return "", "", 0
	}

	title := strings.TrimSpace(doc.Find("title").First().Text())

	// Remove noisy nodes
	doc.Find("script, style, nav, header, footer, aside").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// Gather text from semantic tags
	sb := strings.Builder{}
	doc.Find("main, article, p, h1, h2, h3, h4, h5, h6, li").Each(func(i int, s *goquery.Selection) {
		sb.WriteString(" ")
		sb.WriteString(strings.TrimSpace(s.Text()))
	})

	words := splitWords(sb.String())
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	return title, strings.Join(words, " "), len(words)
}

// helper: unicode-aware split
func splitWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
}
