package ws

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer.
	_WRITE_WAIT = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	// This must be longer than pingPeriod.
	_PONG_WAIT = 120 * time.Second

	// Send pings to peer with this period. Must be less than _PONG_WAIT.
	_PING_PERIOD = (_PONG_WAIT * 9) / 10

	// Maximum message size allowed from peer.
	_MAX_MESSAGE_SIZE = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	UserId string      // UserId is the id of the user whouse client
	Send   chan []byte // Send is a chennel for sending data

	hub       *Hub
	conn      *websocket.Conn
	unregFunc UnregisterFunc
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// NOTE:
		// Allow all connections for development purposes.
		// In production, you should implement proper origin checking.
		return true
	},
	// TODO: add error check func
}

// NewClient Registers new client to hub
func NewClient(hub *Hub, conn *websocket.Conn, userId string, unregFunc UnregisterFunc) error {
	zap.S().Debugf("Registering new client to hub %s", hub.hubId)
	client := &Client{UserId: userId, hub: hub, conn: conn, Send: make(chan []byte, 256), unregFunc: unregFunc}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
	return nil
}

// ReaderFunc used to proces the message
type ReaderFunc func(*Hub, []byte)
type UnregisterFunc func()

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		if c.hub.isEmpty() {
			c.unregFunc()
		}
	}()
	c.conn.SetReadLimit(_MAX_MESSAGE_SIZE)
	c.conn.SetReadDeadline(time.Now().Add(_PONG_WAIT))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(_PONG_WAIT)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				zap.S().Errorf("Error reading message, error: %w", err)
			} else {
				zap.S().Infof("Closing the connection with id: %d", c.UserId)
			}
			break
		}
		c.hub.Handler.HandleMsg(message)
	}

}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(_PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(_WRITE_WAIT))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(_WRITE_WAIT))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
