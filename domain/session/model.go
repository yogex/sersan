package session

import (
	"net/url"

	"github.com/salestock/sersan/lib"
)

// Session session
type Session struct {
	Caps lib.Caps
	URL  *url.URL
}

// Browser browser
type Browser struct {
	Caps    lib.Caps `json:"desiredCapabilities"`
	W3CCaps struct {
		Caps lib.Caps `json:"alwaysMatch"`
	} `json:"capabilities"`
}
