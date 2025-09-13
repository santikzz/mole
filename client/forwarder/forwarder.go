package forwarder

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net/http"
    "strconv"
    "time"
)

type Forwarder struct {
    localPort int
    client    *http.Client
}

type Response struct {
    ID         string            `json:"id"`
    StatusCode int               `json:"status_code"`
    Headers    map[string]string `json:"headers"`
    Body       []byte            `json:"body"`
}

func NewForwarder(localPort int) *Forwarder {
    return &Forwarder{
        localPort: localPort,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (f *Forwarder) Forward(method, urlPath string, headers map[string]string, body []byte) (*Response, error) {
    // construct local url
    localURL := fmt.Sprintf("http://localhost:%d%s", f.localPort, urlPath)
    log.Printf("[FORWARDER] Forwarding %s %s to %s", method, urlPath, localURL)
    
    // create request
    var bodyReader io.Reader
    if len(body) > 0 {
        bodyReader = bytes.NewReader(body)
    }
    
    req, err := http.NewRequest(method, localURL, bodyReader)
    if err != nil {
        log.Printf("[FORWARDER] Failed to create request: %v", err)
        return nil, fmt.Errorf("failed to create request: %v", err)
    }
    
    // add headers with proper filtering
    for key, value := range headers {
        normalizedKey := http.CanonicalHeaderKey(key)
        
        // skip problematic headers that should be handled by http client
        switch normalizedKey {
        case "Connection", "Upgrade", "Proxy-Connection", "Transfer-Encoding":
            continue
        case "Content-Length":
            // only set if we have a body and the value is valid
            if len(body) > 0 {
                if contentLength, err := strconv.Atoi(value); err == nil && contentLength == len(body) {
                    req.Header.Set(key, value)
                }
            }
        case "Host":
            // set host to localhost for local forwarding
            req.Host = fmt.Sprintf("localhost:%d", f.localPort)
        default:
            req.Header.Set(key, value)
        }
    }
    
    // ensure proper content-length for requests with body
    if len(body) > 0 && req.Header.Get("Content-Length") == "" {
        req.Header.Set("Content-Length", strconv.Itoa(len(body)))
    }
    
    log.Printf("[FORWARDER] Making request with %d headers, body size: %d bytes", len(req.Header), len(body))
    
    // make request
    resp, err := f.client.Do(req)
    if err != nil {
        log.Printf("[FORWARDER] Request failed: %v", err)
        return nil, fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()
    
    log.Printf("[FORWARDER] Received response: status %d", resp.StatusCode)
    
    // read response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("[FORWARDER] Failed to read response body: %v", err)
        return nil, fmt.Errorf("failed to read response: %v", err)
    }
    
    log.Printf("[FORWARDER] Response body size: %d bytes", len(respBody))
    
    // prepare response headers with proper filtering
    respHeaders := make(map[string]string)
    for key, values := range resp.Header {
        if len(values) > 0 {
            normalizedKey := http.CanonicalHeaderKey(key)
            // include all headers, filtering will be done at proxy level
            respHeaders[normalizedKey] = values[0]
        }
    }
    
    return &Response{
        StatusCode: resp.StatusCode,
        Headers:    respHeaders,
        Body:       respBody,
    }, nil
}