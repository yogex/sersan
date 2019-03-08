package router

import (
	"net"
	"net/http"
	"strings"

	"github.com/salestock/sersan/handler"
)

type request struct {
	*http.Request
}

func (r request) localaddr() string {
	addr := r.Context().Value(http.LocalAddrContextKey).(net.Addr).String()
	_, port, _ := net.SplitHostPort(addr)
	return net.JoinHostPort("127.0.0.1", port)
}

func mux(rh handler.RootHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/session", rh.Create)
	mux.HandleFunc("/session/", rh.Proxy)
	return mux
}

func CreateRouter(rh handler.RootHandler) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("/wd/hub/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		r.URL.Scheme = "http"
		r.URL.Host = (&request{r}).localaddr()
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/wd/hub")
		mux(rh).ServeHTTP(w, r)
	})
	router.HandleFunc("/health", rh.HealthCheck)
	return router
}
