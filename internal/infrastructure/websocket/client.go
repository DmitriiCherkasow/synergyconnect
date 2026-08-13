package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Время ожидания записи сообщения
	writeWait = 10 * time.Second

	// Время ожидания чтения pong
	pongWait = 60 * time.Second

	// Период отправки ping
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения
	maxMessageSize = 512 * 1024
)

// Client представляет WebSocket соединение пользователя
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	userID   string
	send     chan []byte
	mu       sync.Mutex
	lastPing time.Time
}

// NewClient создает нового клиента
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		userID:   userID,
		send:     make(chan []byte, 256),
		lastPing: time.Now(),
	}
}

// ReadPump читает сообщения из WebSocket соединения
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.mu.Lock()
		c.lastPing = time.Now()
		c.mu.Unlock()
		return nil
	})

	for {
		var msg WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.hub.HandleMessage(c, &msg)
	}
}

// WritePump отправляет сообщения в WebSocket соединение
func (c *Client) WritePump() {
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
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Отправляем все накопившиеся сообщения за один раз
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage отправляет сообщение клиенту
func (c *Client) SendMessage(msg *WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return nil
	}
}

// SendError отправляет ошибку клиенту
func (c *Client) SendError(code, message string) error {
	msg := &WSMessage{
		Type: MsgTypeError,
		Payload: json.RawMessage(`{"code":"` + code + `","message":"` + message + `"}`),
	}
	return c.SendMessage(msg)
}

// Close закрывает соединение
func (c *Client) Close() {
	close(c.send)
	c.conn.Close()
}

// GetUserID возвращает ID пользователя
func (c *Client) GetUserID() string {
	return c.userID
}