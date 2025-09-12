package proxy

import (
    "crypto/rand"
    "encoding/hex"
    "io"
    "net/http"
    "strings"
    "time"
    
    "mole/server/tunnel"
)

type Handler struct {
    manager    *tunnel.Manager
    baseDomain string
    requests   map[string]chan *Response
}

type Request struct {
    ID      string            `json:"id"`
    Method  string            `json:"method"`
    URL     string            `json:"url"`
    Headers map[string]string `json:"headers"`
    Body    []byte            `json:"body"`
}

type Response struct {
    ID         string            `json:"id"`
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       []byte            `json:"body"`
}

func NewHandler(manager *tunnel.Manager, baseDomain string) *Handler {
    return &Handler{
        manager:    manager,
        baseDomain: baseDomain,
        requests:   make(map[string]chan *Response),
    }
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // extract subdomain from host
    host := r.Host
    subdomain := h.extractSubdomain(host)
    
    if subdomain == "" {
        http.Error(w, "invalid subdomain", http.StatusBadRequest)
        return
    }
    
    // find the tunnel connection
    conn := h.manager.GetTunnel(subdomain)
    if conn == nil {
        http.Error(w, "tunnel not found", http.StatusNotFound)
        return
    }
    
    // generate request id
    requestID := h.generateID()
    
    // read request body
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read request body", http.StatusInternalServerError)
        return
    }
    
    // prepare headers
    headers := make(map[string]string)
    for key, values := range r.Header {
        if len(values) > 0 {
            headers[key] = values[0]
        }
    }
    
    // create request object
    req := &Request{
        ID:      requestID,
        Method:  r.Method,
        URL:     r.URL.String(),
        Headers: headers,
        Body:    body,
    }
    
    // create response channel
    respChan := make(chan *Response, 1)
    h.requests[requestID] = respChan
    
    // send request to client
    if err := conn.WriteJSON(req); err != nil {
        delete(h.requests, requestID)
        http.Error(w, "failed to forward request", http.StatusInternalServerError)
        return
    }
    
    // wait for response with timeout
    select {
    case resp := <-respChan:
        // write response headers
        for key, value := range resp.Headers {
            w.Header().Set(key, value)
        }
        w.WriteHeader(resp.StatusCode)
        w.Write(resp.Body)
        
    case <-time.After(30 * time.Second):
        http.Error(w, "request timeout", http.StatusGatewayTimeout)
    }
    
    // cleanup
    delete(h.requests, requestID)
}

func (h *Handler) HandleResponse(resp *Response) {
    if respChan, exists := h.requests[resp.ID]; exists {
        select {
        case respChan <- resp:
        default:
        }
    }
}

func (h *Handler) extractSubdomain(host string) string {
    // remove port if present
    if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
        host = host[:colonIndex]
    }
    
    // check if it's a subdomain of our base domain
    if !strings.HasSuffix(host, "."+h.baseDomain) && host != h.baseDomain {
        return ""
    }
    
    if host == h.baseDomain {
        return ""
    }
    
    // extract subdomain
    subdomain := strings.TrimSuffix(host, "."+h.baseDomain)
    return subdomain
}

func (h *Handler) generateID() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}