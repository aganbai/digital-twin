package manager

import (
	"fmt"
	"sync"
	"time"

	"digital-twin/src/harness/core"
)

// MemoryEventBus 内存事件总线实现
type MemoryEventBus struct {
	subscribers map[string][]func(core.Event)
	mu          sync.RWMutex
}

// NewMemoryEventBus 创建新的内存事件总线
func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		subscribers: make(map[string][]func(core.Event)),
	}
}

// Publish 发布事件
func (eb *MemoryEventBus) Publish(event core.Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	handlers, exists := eb.subscribers[event.Name()]
	if !exists {
		return nil // 没有订阅者
	}

	for _, handler := range handlers {
		go handler(event) // 异步处理
	}

	return nil
}

// Subscribe 订阅事件
func (eb *MemoryEventBus) Subscribe(eventName string, handler func(core.Event)) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscribers[eventName] = append(eb.subscribers[eventName], handler)
	return nil
}

// Unsubscribe 取消订阅
func (eb *MemoryEventBus) Unsubscribe(eventName string, handler func(core.Event)) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	handlers, exists := eb.subscribers[eventName]
	if !exists {
		return fmt.Errorf("no subscribers for event %s", eventName)
	}

	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			eb.subscribers[eventName] = append(handlers[:i], handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("handler not found for event %s", eventName)
}

// HealthCheckEvent 健康检查事件
type HealthCheckEvent struct {
	name      string
	data      interface{}
	timestamp time.Time
}

// NewHealthCheckEvent 创建健康检查事件
func NewHealthCheckEvent(data interface{}) *HealthCheckEvent {
	return &HealthCheckEvent{
		name:      "health_check",
		data:      data,
		timestamp: time.Now(),
	}
}

func (e *HealthCheckEvent) Name() string {
	return e.name
}

func (e *HealthCheckEvent) Data() interface{} {
	return e.data
}

func (e *HealthCheckEvent) Timestamp() time.Time {
	return e.timestamp
}