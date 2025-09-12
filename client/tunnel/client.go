package tunnel

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
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
    u := url.URL{Scheme: "ws", Host: c.serverURL, Path: "/tunnel"}
    
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
    resp, err := c.forwarder.Forward(req.Method, req.URL, req.Headers, req.Body)
    if err != nil {
        // send error response
        errorResp := &Response{
            ID:         req.ID,
            StatusCode: 500,
            Headers:    map[string]string{"Content-Type": "text/plain"},
            Body:       []byte(fmt.Sprintf("forwarding error: %v", err)),
        }
        c.sendResponse(errorResp)
        return
    }
    
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
    // send response back via http post to avoid websocket write conflicts
    jsonData, _ := json.Marshal(resp)
    
    responseURL := fmt.Sprintf("http://%s/response", c.serverURL)
    http.Post(responseURL, "application/json", bytes.NewBuffer(jsonData))
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