package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    
    "mole/server/config"
    "mole/server/proxy"
    "mole/server/tunnel"
)

func main() {
    
	cfg, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    
    log.Printf("starting mole server on port %d for domain %s", cfg.Port, cfg.Domain)
    
    manager := tunnel.NewManager()
    handler := proxy.NewHandler(manager, cfg.Domain)
    
    // websocket endpoint for tunnel connections
    http.HandleFunc("/tunnel", manager.HandleWebSocket)
    
    // response handler for client responses
    http.HandleFunc("/response", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        var resp proxy.Response
        if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        
        handler.HandleResponse(&resp)
        w.WriteHeader(http.StatusOK)
    })
    
    // catch-all handler for proxying requests
    http.HandleFunc("/", handler.ServeHTTP)
    
    // start server
    addr := fmt.Sprintf(":%d", cfg.Port)
    if cfg.UseHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
        log.Printf("starting https server on %s", addr)
        log.Fatal(http.ListenAndServeTLS(addr, cfg.CertFile, cfg.KeyFile, nil))
    } else {
        log.Printf("starting http server on %s", addr)
        log.Fatal(http.ListenAndServe(addr, nil))
    }
}