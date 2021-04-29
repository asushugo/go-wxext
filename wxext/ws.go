package wxext

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
)

type iws interface {
	Conn() (chan<- map[string]interface{}, <-chan map[string]interface{}, <-chan bool, error)
	Close() error
}

type ws struct {
	Addr string
	Port uint16
	Path string
	Name string
	Key  string
	Send chan map[string]interface{}
	Recv chan map[string]interface{}
	conn *websocket.Conn
}

func newWS(Addr string, Port uint16, Name string, Key string) *ws {
	return &ws{
		Addr: Addr,
		Port: Port,
		Name: Name,
		Key:  Key,
		Path: "wx",
	}
}

func (w *ws) Conn() (chan<- map[string]interface{}, <-chan map[string]interface{}, <-chan bool, error) {
	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%s:%d", w.Addr, w.Port), Path: fmt.Sprintf("%s?name=%s&key=%s", w.Path, w.Name, w.Key)}
	URL, err := url.QueryUnescape(u.String())
	if err != nil {
		return nil, nil, nil, err
	}
	c, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		return nil, nil, nil, err
	} else {
		w.conn = c
		w.Send = make(chan map[string]interface{}, 100)
		w.Recv = make(chan map[string]interface{}, 100)
	}
	errChan := make(chan bool, 1)

	go func() {
		for {
			var message map[string]interface{}
			err := c.ReadJSON(&message)
			if err != nil {
				errChan <- false
				continue
			} else if len(message) == 0 {
				errChan <- false
				continue
			}
			w.Recv <- message
		}
	}()

	go func() {
		for m := range w.Send {
			_ = c.WriteJSON(m)
		}
	}()
	return w.Send, w.Recv, errChan, nil
}

func (w *ws) Close() error {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	return w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
