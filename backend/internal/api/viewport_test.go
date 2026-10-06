package api

import (
	"net/http/httptest"
	"testing"
)

func TestViewportValidation(t *testing.T) {
	for _, tc := range []struct {
		query string
		ok    bool
	}{
		{"", true}, {"?x0=0&y0=0&x1=32&y1=32&follow=1", true},
		{"?x0=0", false}, {"?x0=NaN&y0=0&x1=1&y1=1", false},
		{"?x0=0&y0=0&x1=0&y1=1", false}, {"?x0=0&y0=0&x1=1&y1=1&follow=-1", false},
		{"?x0=0&y0=0&x1=Inf&y1=1", false},
	} {
		_, err := parseViewport(httptest.NewRequest("GET", "/stream"+tc.query, nil))
		if (err == nil) != tc.ok {
			t.Errorf("%s: %v", tc.query, err)
		}
	}
}
