package parser

import (
	"bytes"
	"crawler-go/internal/storage"
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(currURL string, content []byte, tokenToParse int) (storage.Webpage, []string) {
	z := html.NewTokenizer(bytes.NewReader(content))
	tokenCount := 0
	bodyStarted := false
	textLen := 0
	wp := storage.Webpage{URL: currURL}
	links := make([]string, 0)

	for {
		if z.Next() == html.ErrorToken || tokenCount > tokenToParse {
			break
		}
		t := z.Token()

		if t.Type == html.StartTagToken {
			switch t.Data {
			case "title":
				z.Next()
				wp.Title = z.Token().Data
			case "body":
				bodyStarted = true
			case "script", "style":
				z.Next() // skip contents
			case "a":
				if href := absoluteHref(t, currURL); href != "" {
					links = append(links, href)
				}
			}
		}

		if bodyStarted && t.Type == html.TextToken && textLen < tokenToParse {
			txt := strings.TrimSpace(t.Data)
			wp.Content += txt
			textLen += len(txt)
		}
		tokenCount++
	}

	// print links for debugging
	// if len(links) > 0 {
	// 	fmt.Printf("Found %d links in %s\n", len(links), currURL)
	// 	for _, l := range links {
	// 		fmt.Printf("  - %s\n", l)
	// 	}
	// }

	return wp, links
}

func absoluteHref(tok html.Token, currURL string) string {
	for _, a := range tok.Attr {
		if a.Key == "href" {
			abs := ResolveLink(currURL, a.Val)
			if abs != "" {
				return abs
			}
			break
		}
	}
	return ""
}

