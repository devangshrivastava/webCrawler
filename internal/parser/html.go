package parser

import (
	"bytes"
	"strings"
	"golang.org/x/net/html"
	"crawler-go/internal/storage"
	
)

func ParseHTML(currURL string, content []byte) (storage.Webpage, []string) {
	z := html.NewTokenizer(bytes.NewReader(content))
	tokenCount := 0
	bodyStarted := false
	textLen := 0
	wp := storage.Webpage{URL: currURL}
	links := make([]string, 0)

	for {
		if z.Next() == html.ErrorToken || tokenCount > 500 {
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

		if bodyStarted && t.Type == html.TextToken && textLen < 500 {
			txt := strings.TrimSpace(t.Data)
			wp.Content += txt
			textLen += len(txt)
		}
		tokenCount++
	}
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

