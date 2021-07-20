package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/facebookgo/inject"
	cache "github.com/patrickmn/go-cache"
	"github.com/salestock/sersan/config"
	"github.com/salestock/sersan/lib"
)

func main() {
	conf := config.Get()

	// Display some important configuration items
	log.Printf("[INIT] Node selector: %s:%s", conf.NodeSelectorKey, conf.NodeSelectorValue)
	log.Printf("[INIT] Cpu request - limit: %s-%s", conf.CPURequest, conf.CPULimit)
	log.Printf("[INIT] Memory request - limit: %s-%s", conf.MemoryRequest, conf.MemoryLimit)

	// Load grid config
	gridConfig := lib.GetGridConfig()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Printf("Could not get current directory:%s", err)
	}
	log.Printf("Current directory: %v", dir)
	err = gridConfig.Load(filepath.Join(dir, conf.GridConfigFile))
	if err != nil {
		log.Printf("Could not load grid config file: %v", err)
	}

	// Tuned http round tripper
	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, ok := defaultRoundTripper.(*http.Transport)
	if !ok {
		log.Printf("defaultRoundTripper not an *http.Transport")
	}
	tunedTransport := *defaultTransportPointer
	tunedTransport.MaxIdleConns = conf.MaxIdleConns
	tunedTransport.MaxIdleConnsPerHost = conf.MaxIdleConnsPerHost
	tunedTransport.MaxConnsPerHost = conf.MaxConnsPerHost

	// Setup dependency injection
	var rh RootHandler
	c := cache.New(time.Duration(conf.CacheTimeout)*time.Minute, time.Duration(conf.CacheTimeout)*time.Duration(2)*time.Minute)
	err = inject.Populate(&rh, c, &tunedTransport)
	if err != nil {
		log.Printf("%v", err)
	}

	// Setup router
	r := CreateRouter(rh)

	// Serve
	var srv http.Server
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		if err = srv.Shutdown(context.Background()); err != nil {
			log.Printf("Sersan API shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()
	srv.Addr = ":" + conf.Port
	srv.Handler = r
	log.Printf("Sersan API started in port: %v", conf.Port)
	if err = srv.ListenAndServe(); err != nil {
		log.Printf("Error starting Sersan API: %v", err)
	}

	<-idleConnsClosed
}
