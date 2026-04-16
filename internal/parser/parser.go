package parser

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var nonDigitRegex = regexp.MustCompile(`[^\d\-]`)

// normalize extracts only digits and hyphens from an HTML string.
func normalize(s string) string {
	decoded := html.UnescapeString(s)
	if decoded == "" {
		return ""
	}
	// Replace unicode minus with standard minus
	decoded = strings.ReplaceAll(decoded, "\u2212", "-")
	result := nonDigitRegex.ReplaceAllString(decoded, "")
	return result
}

// ParseInt parses an int from HTML text content.
func ParseInt(s string) int {
	n := normalize(strings.TrimSpace(s))
	if n == "" {
		return -1
	}
	v, err := strconv.Atoi(n)
	if err != nil {
		return -1
	}
	return v
}

// ParseLong parses an int64 from HTML text content.
func ParseLong(s string) int64 {
	n := normalize(strings.TrimSpace(s))
	if n == "" {
		return -1
	}
	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return -1
	}
	return v
}

// ToDuration parses "HH:MM:SS" or "HH:MM:SS (+NNN ms)" time format.
func ToDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	ms := 0
	if idx := strings.Index(s, "(+"); idx != -1 {
		msPart := s[idx+2:]
		ms = ParseInt(msPart)
		s = strings.TrimSpace(s[:idx])
	}

	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h := ParseInt(parts[0])
	m := ParseInt(parts[1])
	sec := ParseInt(parts[2])
	if h < 0 {
		h = 0
	}
	if m < 0 {
		m = 0
	}
	if sec < 0 {
		sec = 0
	}
	if ms < 0 {
		ms = 0
	}

	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second + time.Duration(ms)*time.Millisecond
}

// DocFromHTML parses an HTML string into a goquery document.
func DocFromHTML(htmlStr string) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
}
