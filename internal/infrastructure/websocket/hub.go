package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub управляет всеми WebSocket соединениями
type Hub struct {
	// Регистрация клиентов
	register chan *Client

	// Отмена регистрации клиентов
	unregister chan *Client

	// Клиенты по userID
	clients map[string]*Client

	// Мьютекс для защиты clients
	mu sync.RWMutex

	// Обработчик сообщений
	messageHandler func(*Client, *WSMessage)
}

// NewHub создает новый хаб
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

// SetMessageHandler устанавливает обработчик сообщений
func (h *Hub) SetMessageHandler(handler func(*Client, *WSMessage)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messageHandler = handler
}

// Run запускает основной цикл хаба
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Закрываем старое соединение, если оно есть
			if oldClient, ok := h.clients[client.userID]; ok {
				oldClient.Close()
			}
			h.clients[client.userID] = client
			h.mu.Unlock()

			log.Printf("Client registered: %s (total: %d)", client.userID, h.GetClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
			}
			h.mu.Unlock()

			log.Printf("Client unregistered: %s (total: %d)", client.userID, h.GetClientCount())
		}
	}
}

// HandleMessage обрабатывает входящее сообщение
func (h *Hub) HandleMessage(client *Client, msg *WSMessage) {
	h.mu.RLock()
	handler := h.messageHandler
	h.mu.RUnlock()

	if handler != nil {
		handler(client, msg)
	}
}

// SendToUser отправляет сообщение пользователю
func (h *Hub) SendToUser(userID string, msg *WSMessage) error {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return nil // Пользователь не в сети
	}

	return client.SendMessage(msg)
}

// SendToUsers отправляет сообщение нескольким пользователям
func (h *Hub) SendToUsers(userIDs []string, msg *WSMessage) {
	for _, userID := range userIDs {
		h.SendToUser(userID, msg)
	}
}

// Broadcast отправляет сообщение всем подключенным пользователям
func (h *Hub) Broadcast(msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

// IsUserOnline проверяет, онлайн ли пользователь
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// GetOnlineUsers возвращает список онлайн пользователей
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// GetClientCount возвращает количество подключенных клиентов
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RegisterClient регистрирует клиента
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient отменяет регистрацию клиента
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// UpgradeToWebSocket обновляет HTTP соединение до WebSocket
func (h *Hub) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, userID string) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // В продакшене нужно проверять origin
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := NewClient(h, conn, userID)
	h.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}