package main

import (
    "github.com/salestock/sersan/domain/health"
    "github.com/salestock/sersan/domain/session"
)

// RootHandler should list all the handler that we will use
type RootHandler struct {
    *session.SessionHandler `inject:""`
    *health.HealthHandler   `inject:""`
}
