package main

import (
	"net/http"
  "log"
	"os"
	"path/filepath"

	"github.com/salestock/sersan/config"
	"github.com/salestock/sersan/handler"
	"github.com/salestock/sersan/router"

	"github.com/facebookgo/inject"
)

func main() {
	conf := config.Get()

	// Display some important configuration items
  log.Printf("[INIT] Node selector: %s:%s", conf.NodeSelectorKey, conf.NodeSelectorValue)
  log.Printf("[INIT] Cpu request - limit: %s-%s", conf.CPURequest, conf.CPULimit)
  log.Printf("[INIT] Memory request - limit: %s-%s", conf.MemoryRequest, conf.MemoryLimit)

	// Load browser config
	browserConfig := config.GetBrowserConfig()
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
    log.Printf("Could not get current directory:%s", err)
	}
  log.Printf("Current directory: %v", dir)
	err = browserConfig.Load(filepath.Join(dir, conf.BrowserConfigFile))
	if err != nil {
    log.Printf("Could not load browser config file: %v", err)
	}

	// Setup dependency injection
	var rh handler.RootHandler
	err = inject.Populate(&rh)
	if err != nil {
		log.Printf("%v", err)
	}

	// Setup router
	r := router.CreateRouter(rh)

	// Serve
	log.Printf("Sersan-api started in port: " + conf.Port)
	err = http.ListenAndServe(":"+conf.Port, r)
	if err != nil {
		log.Printf("%v", err)
	}
}
