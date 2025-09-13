package tunnel

import (
    "fmt"
    "log"
    "net/url"
    "strings"
    
    "github.com/gorilla/websocket"
    
    "mole/client/forwarder"
)

type Client struct {
    serverURL string
    subdomain string
    forwarder *forwarder.Forwarder
    conn      *websocket.Conn
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

func NewClient(serverURL, subdomain string, forwarder *forwarder.Forwarder) *Client {
    return &Client{
        serverURL: serverURL,
        subdomain: subdomain,
        forwarder: forwarder,
    }
}

func (c *Client) Connect() error {
    scheme := "ws"
    if strings.Contains(c.serverURL, "https://") || strings.Contains(c.serverURL, ":443") {
        scheme = "wss"
    }
    u := url.URL{Scheme: scheme, Host: c.serverURL, Path: "/tunnel"}
    
    var err error
    c.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
    if err != nil {
        return fmt.Errorf("failed to connect to server: %v", err)
    }
    
    // register with subdomain
    registerMsg := map[string]interface{}{
        "type":      "register",
        "subdomain": c.subdomain,
    }
    
    if err := c.conn.WriteJSON(registerMsg); err != nil {
        return fmt.Errorf("failed to register: %v", err)
    }
    
    // wait for confirmation
    var response map[string]interface{}
    if err := c.conn.ReadJSON(&response); err != nil {
        return fmt.Errorf("failed to read registration response: %v", err)
    }
    
    if response["type"] != "registered" {
        return fmt.Errorf("registration failed")
    }
    
    log.Printf("tunnel established for %s.%s", c.subdomain, c.extractDomain())
    return nil
}

func (c *Client) Listen() error {
    for {
        var req Request
        if err := c.conn.ReadJSON(&req); err != nil {
            return fmt.Errorf("failed to read request: %v", err)
        }
        
        // forward request to local server
        go c.handleRequest(&req)
    }
}

func (c *Client) handleRequest(req *Request) {
    log.Printf("[CLIENT] Handling request %s: %s %s", req.ID, req.Method, req.URL)
    
    resp, err := c.forwarder.Forward(req.Method, req.URL, req.Headers, req.Body)
    if err != nil {
        log.Printf("[CLIENT] Forwarding failed for request %s: %v", req.ID, err)
        // send error response
        errorResp := &Response{
            ID:         req.ID,
            StatusCode: 502,
            Headers:    map[string]string{"Content-Type": "text/plain"},
            Body:       []byte(fmt.Sprintf("Bad Gateway: %v", err)),
        }
        c.sendResponse(errorResp)
        return
    }
    
    log.Printf("[CLIENT] Forwarding successful for request %s: status %d", req.ID, resp.StatusCode)
    
    // convert forwarder.Response to tunnel.Response
    tunnelResp := &Response{
        ID:         req.ID,
        StatusCode: resp.StatusCode,
        Headers:    resp.Headers,
        Body:       resp.Body,
    }
    c.sendResponse(tunnelResp)
}

func (c *Client) sendResponse(resp *Response) {
    // send response back via websocket for proper tunneling
    log.Printf("[CLIENT] Sending response for request %s: status %d, body size %d bytes", resp.ID, resp.StatusCode, len(resp.Body))
    
    if err := c.conn.WriteJSON(resp); err != nil {
        log.Printf("[CLIENT] Failed to send response for request %s: %v", resp.ID, err)
    } else {
        log.Printf("[CLIENT] Response sent successfully for request %s", resp.ID)
    }
}

func (c *Client) extractDomain() string {
    // simple extraction - assumes server url is "host:port"
    if colonIndex := len(c.serverURL); colonIndex > 0 {
        return c.serverURL[:strings.Index(c.serverURL, ":")]
    }
    return c.serverURL
}

func (c *Client) Close() error {
    if c.conn != nil {
        return c.conn.Close()
    }
    return nil
}