package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port     int
    Domain   string
    CertFile string
    KeyFile  string
    UseHTTPS bool
}

func Load() (*Config, error) {
    cfg := &Config{}
    
    // load from environment variables
    if port := os.Getenv("MOLE_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            cfg.Port = p
        }
    }
    
    cfg.Domain = os.Getenv("MOLE_DOMAIN")
    cfg.CertFile = os.Getenv("MOLE_CERT_FILE")
    cfg.KeyFile = os.Getenv("MOLE_KEY_FILE")
    
    if https := os.Getenv("MOLE_USE_HTTPS"); https == "true" {
        cfg.UseHTTPS = true
    }
    
    // set defaults
    if cfg.Port == 0 {
        cfg.Port = 3000
    }
    if cfg.Domain == "" {
        cfg.Domain = "localhost"
    }
    
    // automatically set certificate paths if HTTPS is enabled but paths not specified
    if cfg.UseHTTPS && cfg.CertFile == "" {
        cfg.CertFile = "/etc/letsencrypt/live/" + cfg.Domain + "/fullchain.pem"
    }
    if cfg.UseHTTPS && cfg.KeyFile == "" {
        cfg.KeyFile = "/etc/letsencrypt/live/" + cfg.Domain + "/privkey.pem"
    }
    
    return cfg, nil
}