package forwarder

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
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
    
    // create request
    req, err := http.NewRequest(method, localURL, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }
    
    // add headers
    for key, value := range headers {
        // skip certain headers that might cause issues
        if key == "Host" || key == "Connection" {
            continue
        }
        req.Header.Set(key, value)
    }
    
    // make request
    resp, err := f.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()
    
    // read response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %v", err)
    }
    
    // prepare response headers
    respHeaders := make(map[string]string)
    for key, values := range resp.Header {
        if len(values) > 0 {
            respHeaders[key] = values[0]
        }
    }
    
    return &Response{
        StatusCode: resp.StatusCode,
        Headers:    respHeaders,
        Body:       respBody,
    }, nil
}