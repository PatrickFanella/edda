package domain

import (
	"regexp"
	"strings"
)

var nonLocationNameWord = regexp.MustCompile(`[^a-z0-9]+`)

// CanonicalLocationName returns a conservative key for detecting obvious same-location variants.
// It is not a broad fuzzy matcher.
func CanonicalLocationName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = nonLocationNameWord.ReplaceAllString(name, " ")
	parts := strings.Fields(name)
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "the" || part == "a" || part == "an" {
			continue
		}
		out = append(out, singularizeLocationToken(part))
	}
	return strings.Join(out, " ")
}

func SameCanonicalLocationName(a, b string) bool {
	ca := CanonicalLocationName(a)
	cb := CanonicalLocationName(b)
	return ca != "" && ca == cb
}

func singularizeLocationToken(token string) string {
	if len(token) <= 3 {
		return token
	}
	if strings.HasSuffix(token, "ies") && len(token) > 4 {
		return strings.TrimSuffix(token, "ies") + "y"
	}
	if strings.HasSuffix(token, "s") && !strings.HasSuffix(token, "ss") {
		return strings.TrimSuffix(token, "s")
	}
	return token
}
