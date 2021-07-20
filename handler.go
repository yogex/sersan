package main

import (
    "github.com/growbak/hub/domain/health"
    "github.com/growbak/hub/domain/session"
)

// RootHandler should list all the handler that we will use
type RootHandler struct {
    *session.SessionHandler `inject:""`
    *health.HealthHandler   `inject:""`
}
