package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

func RandID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 10)
	seed := time.Now().UnixNano()
	for i := range b {
		seed = (seed*1664525 + 1013904223) % 4294967296
		idx := int(math.Abs(float64(seed % int64(len(letters)))))
		b[i] = letters[idx]
	}
	return string(b)
}

func UrlEscape(s string) string {
	r := strings.NewReplacer(
		" ", "%20", "!", "%21", "\"", "%22", "#", "%23",
		"$", "%24", "%", "%25", "&", "%26", "'", "%27",
		"(", "%28", ")", "%29", "*", "%2A", "+", "%2B",
		",", "%2C", "/", "%2F", ":", "%3A", ";", "%3B",
		"<", "%3C", "=", "%3D", ">", "%3E", "?", "%3F",
		"@", "%40", "[", "%5B", "\\", "%5C", "]", "%5D",
		"^", "%5E", "`", "%60", "{", "%7B", "|", "%7C",
		"}", "%7D", "~", "%7E",
	)
	return r.Replace(s)
}

func JsonIndent(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json indent: %w", err)
	}
	return string(b), nil
}
