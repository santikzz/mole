package tunnel

import (
    "log"
    "net/http"
    "sync"
    
    "github.com/gorilla/websocket"
)

type Manager struct {
    tunnels  map[string]*websocket.Conn
    mutex    sync.RWMutex
    upgrader websocket.Upgrader
}

func NewManager() *Manager {
    return &Manager{
        tunnels: make(map[string]*websocket.Conn),
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // allow all origins for development
            },
        },
    }
}

func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := m.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("websocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()
    
    // expect first message to contain subdomain
    var msg struct {
        Type      string `json:"type"`
        Subdomain string `json:"subdomain"`
    }
    
    if err := conn.ReadJSON(&msg); err != nil {
        log.Printf("failed to read initial message: %v", err)
        return
    }
    
    if msg.Type != "register" {
        log.Printf("expected register message, got: %s", msg.Type)
        return
    }
    
    subdomain := msg.Subdomain
    if subdomain == "" {
        log.Printf("subdomain is required")
        return
    }
    
    // register the tunnel
    m.mutex.Lock()
    m.tunnels[subdomain] = conn
    m.mutex.Unlock()
    
    log.Printf("tunnel registered for subdomain: %s", subdomain)
    
    // send confirmation
    conn.WriteJSON(map[string]interface{}{
        "type":    "registered",
        "subdomain": subdomain,
    })
    
    // keep connection alive and handle cleanup
    for {
        if _, _, err := conn.NextReader(); err != nil {
            break
        }
    }
    
    // cleanup when connection closes
    m.mutex.Lock()
    delete(m.tunnels, subdomain)
    m.mutex.Unlock()
    
    log.Printf("tunnel closed for subdomain: %s", subdomain)
}

func (m *Manager) GetTunnel(subdomain string) *websocket.Conn {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.tunnels[subdomain]
}