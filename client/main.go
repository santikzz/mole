package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    
    "mole/client/config"
    "mole/client/forwarder"
    "mole/client/tunnel"
)

func main() {
    
	if len(os.Args) < 3 || os.Args[1] != "http" {
        fmt.Println("usage: mole http <port> [-d subdomain]")
        os.Exit(1)
    }
    
    // parse local port
    localPort, err := strconv.Atoi(os.Args[2])
    if err != nil {
        log.Fatalf("invalid port: %v", err)
    }
    
    cfg, subdomainOverride, _, err := config.Load()
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }
    
    // determine subdomain
    subdomain := cfg.Subdomain
    if subdomainOverride != nil {
        subdomain = *subdomainOverride
    }
    
    if subdomain == "" {
        log.Fatalf("subdomain is required (set in config.json or use -d flag)")
    }
    
    // create forwarder
    fwd := forwarder.NewForwarder(localPort)
    
    // create tunnel client
    serverURL := fmt.Sprintf("%s:%d", cfg.Server, cfg.Port)
    client := tunnel.NewClient(serverURL, subdomain, fwd)
    
    // connect to server
    if err := client.Connect(); err != nil {
        log.Fatalf("failed to connect: %v", err)
    }
    defer client.Close()
    
    protocol := "http"
    if cfg.UseHTTPS {
        protocol = "https"
    }
    log.Printf("forwarding http://localhost:%d to %s://%s.%s", localPort, protocol, subdomain, cfg.Server)
    
    // handle shutdown gracefully
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c // wait for signal
        log.Println("shutting down...")
        client.Close()
        os.Exit(0)
    }()
    
    // start listening for requests
    if err := client.Listen(); err != nil {
        log.Fatalf("tunnel error: %v", err)
    }
}