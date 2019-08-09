package noolite

//type Client struct {
//conn *Connection
//handlers [commandCount]ClientHandler
//stop ch struct{}
//}

//type ClientHandler func(resp *Response) (*Request, error)

//func NewClient(c *Connection, mode byte) *Client {
//return &Client{
//conn: c,
//stop: make(chan struct{})
//}
//}

//func (c *Client) Start() {
//for {
//select {
//case <-c.stop:
//return
//default:

//}
//}
//}

//func (c *Client) RegisterHandler(command byte, h ClientHandler) {
//if command > commandCount {
//return
//}
//c.handlers[command] = h
//}
