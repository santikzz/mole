package config

import (
    "flag"
    "os"
    "strconv"
    
    "github.com/joho/godotenv"
)

type Config struct {
    Port     int
    Domain   string
    CertFile string
    KeyFile  string
    UseHTTPS bool
}

func Load() (*Config, error) {
    // load .env file if it exists
    godotenv.Load()
    
    // parse command line flags
    var portFlag = flag.Int("port", 0, "server port (overrides MOLE_PORT)")
    var domainFlag = flag.String("domain", "", "server domain (overrides MOLE_DOMAIN)")
    flag.Parse()
    
    cfg := &Config{}
    
    // load from environment variables
    if port := os.Getenv("MOLE_PORT"); port != "" {
        if p, err := strconv.Atoi(port); err == nil {
            cfg.Port = p
        }
    }
    
    cfg.Domain = os.Getenv("MOLE_DOMAIN")
    
    // override with command line flags if provided
    if *portFlag != 0 {
        cfg.Port = *portFlag
    }
    if *domainFlag != "" {
        cfg.Domain = *domainFlag
    }
    cfg.CertFile = os.Getenv("MOLE_CERT_FILE")
    cfg.KeyFile = os.Getenv("MOLE_KEY_FILE")
    
    if https := os.Getenv("MOLE_USE_HTTPS"); https == "true" {
        cfg.UseHTTPS = true
    }
    
    // set defaults
    if cfg.Port == 0 {
        cfg.Port = 80
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