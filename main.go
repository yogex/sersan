package main

import (
    "log"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/facebookgo/inject"
    "github.com/patrickmn/go-cache"
    "github.com/salestock/sersan/config"
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
    log.Printf("Sersan-api started in port: " + conf.Port)
    err = http.ListenAndServe(":"+conf.Port, r)
    if err != nil {
        log.Printf("%v", err)
    }
}
