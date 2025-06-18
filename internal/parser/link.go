// internal/parser/link.go
package parser

import (
	"net/url"
	"strings"
)

// schemes we refuse to crawl
var badScheme = map[string]struct{}{
	"mailto":     {},
	"javascript": {},
	"tel":        {},
	"data":       {},
}

// ResolveLink converts a raw <a href="…"> into an absolute URL string.
// It returns "" if the link should be ignored.
func ResolveLink(base, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return ""
	}

	// Parse the base only once—cheap enough here; could cache per-host later.
	bu, err := url.Parse(base)
	if err != nil {
		return ""
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	// Disallow unsupported or dangerous schemes.
	if ref.Scheme != "" {
		if _, bad := badScheme[strings.ToLower(ref.Scheme)]; bad {
			return ""
		}
		if ref.Scheme != "http" && ref.Scheme != "https" {
			return ""
		}
	}

	abs := bu.ResolveReference(ref)
	abs.Fragment = "" // drop #section

	// Normalise: add a slash if path is empty so hashes match consistently.
	if abs.Path == "" {
		abs.Path = "/"
	}
	return abs.String()
}
