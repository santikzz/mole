package config

import (
    "encoding/json"
    "flag"
    "os"
)

type Config struct {
    Port     int    `json:"port"`
    Domain   string `json:"domain"`
    CertFile string `json:"cert_file"`
    KeyFile  string `json:"key_file"`
    UseHTTPS bool   `json:"use_https"`
}

func Load() (*Config, error) {
    
	cfg := &Config{}
    
    // load from config file first
    if data, err := os.ReadFile("config.json"); err == nil {
        json.Unmarshal(data, cfg)
    }
    
    // override with command line flags
    port     := flag.Int("port", cfg.Port, "server port")
    domain   := flag.String("domain", cfg.Domain, "server domain")
    certFile := flag.String("cert", cfg.CertFile, "ssl certificate file")
    keyFile  := flag.String("key", cfg.KeyFile, "ssl key file")
    useHTTPS := flag.Bool("https", cfg.UseHTTPS, "use https")
    flag.Parse()
    
    if *port != 0 {
        cfg.Port = *port
    }
    if *domain != "" {
        cfg.Domain = *domain
    }
    if *certFile != "" {
        cfg.CertFile = *certFile
    }
    if *keyFile != "" {
        cfg.KeyFile = *keyFile
    }
    cfg.UseHTTPS = *useHTTPS
    
    // set defaults
    if cfg.Port == 0 {
        cfg.Port = 3000
    }
    if cfg.Domain == "" {
        cfg.Domain = "localhost"
    }
    
    return cfg, nil
}