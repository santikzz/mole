package config

import (
    "encoding/json"
    "flag"
    "fmt"
    "os"
)

type Config struct {
    Server    string `json:"server"`
    Port      int    `json:"port"`
    Subdomain string `json:"subdomain"`
    UseHTTPS  bool   `json:"use_https"`
}

func Load() (*Config, *string, *int, error) {
    cfg := &Config{}
    
    // load from config file first
    if data, err := os.ReadFile("config.json"); err == nil {
        json.Unmarshal(data, cfg)
    }
    
    // parse command line arguments
    var subdomain *string
    var localPort *int
    
    if len(os.Args) >= 3 && os.Args[1] == "http" {
        // extract port from "mole http 8000"
        localPortValue := 0
        if len(os.Args) >= 3 {
            if _, err := fmt.Sscanf(os.Args[2], "%d", &localPortValue); err == nil {
                localPort = &localPortValue
            }
        }
        
        // check for flags
        flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
        subdomainFlag := flag.String("d", "", "subdomain to use")
        flag.CommandLine.Parse(os.Args[3:])
        
        if *subdomainFlag != "" {
            subdomain = subdomainFlag
        }
    }
    
    // set defaults
    if cfg.Server == "" {
        cfg.Server = "localhost"
    }
    if cfg.Port == 0 {
        cfg.Port = 80
    }
    
    return cfg, subdomain, localPort, nil
}