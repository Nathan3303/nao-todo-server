package sse

import (
	"encoding/json"
	"sync"
)

// ReminderEvent 提醒事件
type ReminderEvent struct {
	Type        string `json:"type"`
	TaskId      string `json:"taskId"`
	TaskName    string `json:"taskName"`
	Description string `json:"description"`
	RemindAt    string `json:"remindAt"`
}

// Hub SSE 连接管理中心
type Hub struct {
	clients map[int64]map[chan []byte]struct{}
	mu      sync.RWMutex
}

var (
	hub  *Hub
	once sync.Once
)

// GetHub 获取 Hub 单例
func GetHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			clients: make(map[int64]map[chan []byte]struct{}),
		}
	})
	return hub
}

// Publish 向指定用户推送事件
// @param userId 用户ID
// @param event 事件
func (h *Hub) Publish(userId int64, event ReminderEvent) {
	h.mu.RLock()
	channels, ok := h.clients[userId]
	h.mu.RUnlock()
	if !ok {
		return
	}
	// 序列化事件
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range channels {
		select {
		case ch <- data:
		default:
		}
	}
}

// Subscribe 订阅用户事件通道
// @param userId 用户ID
// @return chan []byte 事件通道
func (h *Hub) Subscribe(userId int64) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 10)
	if h.clients[userId] == nil {
		h.clients[userId] = make(map[chan []byte]struct{})
	}
	h.clients[userId][ch] = struct{}{}
	return ch
}

// Unsubscribe 取消订阅
// @param userId 用户ID
// @param ch 事件通道
func (h *Hub) Unsubscribe(userId int64, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[userId]; ok {
		delete(clients, ch)
		close(ch)
		if len(clients) == 0 {
			delete(h.clients, userId)
		}
	}
}
