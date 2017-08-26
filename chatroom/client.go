package chatroom

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/ChatRoom1234/ChatChatChat/db"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline           = []byte{'\n'}
	space             = []byte{' '}
	usernameSeparator = []byte(": ")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the room.
type Client struct {
	room *Room

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the room.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(accessKey string) {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		userID, username := db.GetUserByKey(accessKey)
		// username = []byte(username)
		if userID == 0 {
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		db.AddMessage(userID, message)

		message = append([]byte(username+": "), message...)
		c.room.broadcast <- message
	}
}

// writePump pumps messages from the room to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The room closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(room *Room, w http.ResponseWriter, r *http.Request) {
	accessKey, err := r.Cookie("access_key")
	if err != nil {
		log.Printf("error getting access_key: %v ", err)
		return
	}

	log.Printf("access_key: %s ", accessKey.Value)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{room: room, conn: conn, send: make(chan []byte, 256)}
	client.room.register <- client

	go client.writePump()
	go func(accessKey string) {
		client.readPump(accessKey)
	}(accessKey.Value)
	// go func(accessKey string) {
	// 	username := db.GetUserByKey(accessKey)
	// 	client.readPump(append([]byte(username), usernameSeparator...))
	// }(accessKey.Value)
}
