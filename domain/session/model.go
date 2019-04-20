package session

import (
    "net/url"

    "github.com/salestock/sersan/service"
)

// Session session
type Session struct {
    Caps service.Caps
    URL  *url.URL
}

// Browser browser
type Browser struct {
    Caps    service.Caps `json:"desiredCapabilities"`
    W3CCaps struct {
        Caps  service.Caps `json:"alwaysMatch"`
    } `json:"capabilities"`
}
