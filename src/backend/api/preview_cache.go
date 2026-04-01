package api

import (
	"sync"
	"time"
)

// ======================== V2.0 迭代3 预览缓存 ========================

// previewCacheEntry 预览缓存条目
type previewCacheEntry struct {
	Content   string                   `json:"content"`
	Title     string                   `json:"title"`
	Tags      string                   `json:"tags"`
	Chunks    []map[string]interface{} `json:"chunks"`
	DocType   string                   `json:"doc_type"`
	SourceURL string                   `json:"source_url,omitempty"`
	CreatedAt time.Time                `json:"created_at"`
}

var (
	previewCache     sync.Map
	previewCacheOnce sync.Once
)

const (
	previewCacheExpiry = 30 * time.Minute
	previewCacheMax    = 100
)

// initPreviewCacheCleanup 初始化预览缓存清理 goroutine
func initPreviewCacheCleanup() {
	previewCacheOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				now := time.Now()
				previewCache.Range(func(key, value interface{}) bool {
					entry, ok := value.(*previewCacheEntry)
					if ok && now.Sub(entry.CreatedAt) > previewCacheExpiry {
						previewCache.Delete(key)
					}
					return true
				})
			}
		}()
	})
}

// getPreviewCacheSize 获取缓存当前大小
func getPreviewCacheSize() int {
	count := 0
	previewCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
