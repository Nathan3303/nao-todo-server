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
	clients          map[int64]map[chan []byte]string // channel → token
	mu               sync.RWMutex
	SessionValidator func(userId int64, token string) bool
}

var (
	hub  *Hub
	once sync.Once
)

// GetHub 获取 Hub 单例
func GetHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			clients: make(map[int64]map[chan []byte]string),
		}
	})
	return hub
}

// Publish 向指定用户推送提醒事件
func (h *Hub) Publish(userId int64, event ReminderEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.publishRaw(userId, data)
}

// PublishJSON 向指定用户推送任意 JSON 序列化的事件
// 用于领域事件触发的 SSE 通知（如项目删除、任务级联等）
func (h *Hub) PublishJSON(userId int64, event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.publishRaw(userId, data)
}

// publishRaw 向指定用户的所有有效通道推送原始数据
func (h *Hub) publishRaw(userId int64, data []byte) {
	h.mu.RLock()
	channels, ok := h.clients[userId]
	h.mu.RUnlock()
	if !ok {
		return
	}
	// 收集需要清理的无效通道
	var invalidChs []chan []byte
	h.mu.RLock()
	for ch, token := range channels {
		if h.SessionValidator != nil && !h.SessionValidator(userId, token) {
			invalidChs = append(invalidChs, ch)
			continue
		}
		select {
		case ch <- data:
		default:
		}
	}
	h.mu.RUnlock()
	// 清理无效通道
	if len(invalidChs) > 0 {
		h.mu.Lock()
		for _, ch := range invalidChs {
			if _, exists := channels[ch]; exists {
				delete(channels, ch)
				close(ch)
			}
		}
		if len(channels) == 0 {
			delete(h.clients, userId)
		}
		h.mu.Unlock()
	}
}

// Subscribe 订阅用户事件通道
func (h *Hub) Subscribe(userId int64, token string) chan []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan []byte, 10)
	if h.clients[userId] == nil {
		h.clients[userId] = make(map[chan []byte]string)
	}
	h.clients[userId][ch] = token
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
