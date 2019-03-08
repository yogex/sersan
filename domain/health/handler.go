package health

import (
	"net/http"

	"github.com/salestock/sersan/utils"
)

type HealthHandler struct {
}

func (c HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.ResponseOk(w, 200, "Sersan API running")
}
