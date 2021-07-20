package health

import (
    "net/http"

    "github.com/growbak/hub/utils"
)

type HealthHandler struct {
}

func (c HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    utils.ResponseOk(w, 200, "Sersan API running")
}
