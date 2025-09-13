package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"
    
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
    log.Printf("https enabled: %v", cfg.UseHTTPS)
    if cfg.UseHTTPS {
        log.Printf("cert file: %s, key file: %s", cfg.CertFile, cfg.KeyFile)
    }
    log.Printf("verbose logging enabled - all requests will be logged")
    
    manager := tunnel.NewManager()
    handler := proxy.NewHandler(manager, cfg.Domain)
    
    // logging middleware
    loggingHandler := func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            log.Printf("[REQUEST] %s %s from %s - User-Agent: %s", r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
            if r.URL.RawQuery != "" {
                log.Printf("[REQUEST] Query: %s", r.URL.RawQuery)
            }
            
            next(w, r)
            
            duration := time.Since(start)
            log.Printf("[RESPONSE] %s %s completed in %v", r.Method, r.URL.Path, duration)
        }
    }
    
    // websocket endpoint for tunnel connections
    http.HandleFunc("/tunnel", loggingHandler(manager.HandleWebSocket))
    
    // response handler for client responses
    http.HandleFunc("/response", loggingHandler(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            log.Printf("[ERROR] Invalid method %s for /response endpoint", r.Method)
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        
        var resp proxy.Response
        if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
            log.Printf("[ERROR] Failed to decode response JSON: %v", err)
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }
        
        log.Printf("[RESPONSE] Handling response for request ID: %s", resp.RequestID)
        handler.HandleResponse(&resp)
        w.WriteHeader(http.StatusOK)
    }))
    
    // catch-all handler for proxying requests
    http.HandleFunc("/", loggingHandler(handler.ServeHTTP))
    
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