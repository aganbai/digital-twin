// 知识包管理模块
// 用于管理从KAT (Knowledge Automation Toolkit) 生成的预置教育知识包
package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// 包元数据结构
type PackMetadata struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Subject       string  `json:"subject"`
	SubjectName   string  `json:"subject_name"`
	Level         string  `json:"level"`
	Version       string  `json:"version"`
	CreatedAt     string  `json:"created_at"`
	EstimatedSize float64 `json:"estimated_size_mb"`
	VectorCount   int     `json:"vector_count"`
	ChunkCount    int     `json:"chunk_count"`
	Status        string  `json:"status"`
}

// 包索引结构
type PackIndex struct {
	Version     string         `json:"version"`
	CreatedAt   string         `json:"created_at"`
	TotalPacks  int            `json:"total_packs"`
	TotalSize   float64        `json:"total_size_mb"`
	Compression bool           `json:"compression"`
	Encryption  bool           `json:"encryption"`
	Packs       []PackMetadata `json:"packs"`
}

// 包管理器配置
type PackManagerConfig struct {
	BasePath       string `json:"base_path"` // 知识包基础路径
	EnableCache    bool   `json:"enable_cache"`
	CacheDuration  int    `json:"cache_duration"` // 缓存时长（秒）
	DefaultSubject string `json:"default_subject"`
	DefaultLevel   string `json:"default_level"`
}

// 包管理器
type PackManager struct {
	config    PackManagerConfig
	mu        sync.RWMutex
	index     *PackIndex
	packs     map[string]*PackMetadata // id -> metadata
	bySubject map[string][]string      // subject -> [packIDs]
	byLevel   map[string][]string      // level -> [packIDs]
}

// 默认配置
var DefaultConfig = PackManagerConfig{
	BasePath:       "data/prebuilt-knowledge-packs/knowledge_packs",
	EnableCache:    true,
	CacheDuration:  3600, // 1小时
	DefaultSubject: "math",
	DefaultLevel:   "小学",
}

// 创建新的包管理器
func NewPackManager(config PackManagerConfig) (*PackManager, error) {
	if config.BasePath == "" {
		config.BasePath = DefaultConfig.BasePath
	}

	manager := &PackManager{
		config:    config,
		packs:     make(map[string]*PackMetadata),
		bySubject: make(map[string][]string),
		byLevel:   make(map[string][]string),
	}

	// 加载包索引
	if err := manager.loadIndex(); err != nil {
		return nil, fmt.Errorf("加载知识包索引失败: %v", err)
	}

	return manager, nil
}

// 加载包索引
func (m *PackManager) loadIndex() error {
	indexPath := filepath.Join(m.config.BasePath, "packages.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("读取索引文件失败: %v", err)
	}

	var index PackIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("解析索引JSON失败: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.index = &index

	// 构建索引
	for i, pack := range index.Packs {
		m.packs[pack.ID] = &index.Packs[i]

		// 按学科索引
		if _, ok := m.bySubject[pack.Subject]; !ok {
			m.bySubject[pack.Subject] = []string{}
		}
		m.bySubject[pack.Subject] = append(m.bySubject[pack.Subject], pack.ID)

		// 按年级索引
		if _, ok := m.byLevel[pack.Level]; !ok {
			m.byLevel[pack.Level] = []string{}
		}
		m.byLevel[pack.Level] = append(m.byLevel[pack.Level], pack.ID)
	}

	return nil
}

// 获取包总数
func (m *PackManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.index == nil {
		return 0
	}
	return m.index.TotalPacks
}

// 获取所有包
func (m *PackManager) GetAllPacks() []PackMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.index == nil {
		return []PackMetadata{}
	}

	// 返回副本
	packs := make([]PackMetadata, len(m.index.Packs))
	copy(packs, m.index.Packs)
	return packs
}

// 获取包
func (m *PackManager) GetPack(packID string) (*PackMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pack, ok := m.packs[packID]
	if !ok {
		return nil, fmt.Errorf("知识包不存在: %s", packID)
	}

	// 返回副本
	packCopy := *pack
	return &packCopy, nil
}

// 按学科和年级获取包
func (m *PackManager) GetPackBySubjectLevel(subject, level string) (*PackMetadata, error) {
	packID := fmt.Sprintf("%s_%s", subject, level)
	return m.GetPack(packID)
}

// 按学科查询包
func (m *PackManager) GetPacksBySubject(subject string) []PackMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	packIDs, ok := m.bySubject[subject]
	if !ok {
		return []PackMetadata{}
	}

	packs := make([]PackMetadata, 0, len(packIDs))
	for _, id := range packIDs {
		if pack, ok := m.packs[id]; ok {
			packs = append(packs, *pack)
		}
	}

	// 按年级排序
	sort.Slice(packs, func(i, j int) bool {
		levelOrder := map[string]int{
			"小学": 1,
			"初中": 2,
			"高中": 3,
			"大学": 4,
		}
		iOrder, jOrder := levelOrder[packs[i].Level], levelOrder[packs[j].Level]
		return iOrder < jOrder
	})

	return packs
}

// 按年级查询包
func (m *PackManager) GetPacksByLevel(level string) []PackMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	packIDs, ok := m.byLevel[level]
	if !ok {
		return []PackMetadata{}
	}

	packs := make([]PackMetadata, 0, len(packIDs))
	for _, id := range packIDs {
		if pack, ok := m.packs[id]; ok {
			packs = append(packs, *pack)
		}
	}

	// 按学科排序
	sort.Slice(packs, func(i, j int) bool {
		subjectOrder := map[string]int{
			"math":      1,
			"physics":   2,
			"chemistry": 3,
			"biology":   4,
			"chinese":   5,
			"english":   6,
			"history":   7,
			"geography": 8,
			"politics":  9,
			"music":     10,
			"art":       11,
			"pe":        12,
		}
		iOrder, jOrder := subjectOrder[packs[i].Subject], subjectOrder[packs[j].Subject]
		return iOrder < jOrder
	})

	return packs
}

// 获取所有学科
func (m *PackManager) GetAllSubjects() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subjects := make([]string, 0, len(m.bySubject))
	for subject := range m.bySubject {
		subjects = append(subjects, subject)
	}

	// 按预定义顺序排序
	subjectOrder := map[string]int{
		"math":      1,
		"physics":   2,
		"chemistry": 3,
		"biology":   4,
		"chinese":   5,
		"english":   6,
		"history":   7,
		"geography": 8,
		"politics":  9,
		"music":     10,
		"art":       11,
		"pe":        12,
	}

	sort.Slice(subjects, func(i, j int) bool {
		iOrder, jOrder := subjectOrder[subjects[i]], subjectOrder[subjects[j]]
		return iOrder < jOrder
	})

	return subjects
}

// 获取所有年级
func (m *PackManager) GetAllLevels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	levels := make([]string, 0, len(m.byLevel))
	for level := range m.byLevel {
		levels = append(levels, level)
	}

	// 按预定义顺序排序
	levelOrder := map[string]int{
		"小学": 1,
		"初中": 2,
		"高中": 3,
		"大学": 4,
	}

	sort.Slice(levels, func(i, j int) bool {
		iOrder, jOrder := levelOrder[levels[i]], levelOrder[levels[j]]
		return iOrder < jOrder
	})

	return levels
}

// 获取学科的中文名称
func (m *PackManager) GetSubjectName(subject string) string {
	subjectNames := map[string]string{
		"math":      "数学",
		"physics":   "物理",
		"chemistry": "化学",
		"biology":   "生物",
		"chinese":   "语文",
		"english":   "英语",
		"history":   "历史",
		"geography": "地理",
		"politics":  "政治",
		"music":     "音乐",
		"art":       "美术",
		"pe":        "体育",
	}

	if name, ok := subjectNames[subject]; ok {
		return name
	}
	return subject
}

// 检查包是否存在
func (m *PackManager) HasPack(packID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.packs[packID]
	return ok
}

// 检查学科和年级是否存在包
func (m *PackManager) HasSubjectLevel(subject, level string) bool {
	packID := fmt.Sprintf("%s_%s", subject, level)
	return m.HasPack(packID)
}

// 重新加载索引
func (m *PackManager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清空现有数据
	m.index = nil
	m.packs = make(map[string]*PackMetadata)
	m.bySubject = make(map[string][]string)
	m.byLevel = make(map[string][]string)

	// 重新加载
	return m.loadIndex()
}

// 获取统计信息
func (m *PackManager) GetStats() gin.H {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.index == nil {
		return gin.H{
			"loaded":        false,
			"total_packs":   0,
			"total_size_mb": 0.0,
		}
	}

	return gin.H{
		"loaded":         true,
		"total_packs":    m.index.TotalPacks,
		"total_size_mb":  m.index.TotalSize,
		"compression":    m.index.Compression,
		"encryption":     m.index.Encryption,
		"version":        m.index.Version,
		"created_at":     m.index.CreatedAt,
		"subjects_count": len(m.bySubject),
		"levels_count":   len(m.byLevel),
	}
}

// 获取包路径
func (m *PackManager) GetPackPath(packID string) (string, error) {
	if !m.HasPack(packID) {
		return "", fmt.Errorf("知识包不存在: %s", packID)
	}

	return filepath.Join(m.config.BasePath, packID), nil
}

// 获取包的元数据文件路径
func (m *PackManager) GetMetadataPath(packID string) (string, error) {
	packPath, err := m.GetPackPath(packID)
	if err != nil {
		return "", err
	}

	return filepath.Join(packPath, "metadata.json"), nil
}

// 获取包的README路径
func (m *PackManager) GetReadmePath(packID string) (string, error) {
	packPath, err := m.GetPackPath(packID)
	if err != nil {
		return "", err
	}

	return filepath.Join(packPath, "README.md"), nil
}

// 验证包完整性
func (m *PackManager) ValidatePack(packID string) error {
	metadataPath, err := m.GetMetadataPath(packID)
	if err != nil {
		return err
	}

	// 检查文件是否存在
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return fmt.Errorf("元数据文件不存在: %s", metadataPath)
	}

	// 读取并验证元数据
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("读取元数据文件失败: %v", err)
	}

	var metadata PackMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("解析元数据JSON失败: %v", err)
	}

	// 验证ID匹配
	if metadata.ID != packID {
		return fmt.Errorf("元数据ID不匹配: 期望 %s, 实际 %s", packID, metadata.ID)
	}

	// 验证必需字段
	requiredFields := []struct {
		field string
		value string
	}{
		{"subject", metadata.Subject},
		{"subject_name", metadata.SubjectName},
		{"level", metadata.Level},
		{"status", metadata.Status},
	}

	for _, field := range requiredFields {
		if field.value == "" {
			return fmt.Errorf("必需字段为空: %s", field.field)
		}
	}

	return nil
}

// 验证所有包
func (m *PackManager) ValidateAll() map[string]error {
	m.mu.RLock()
	packIDs := make([]string, 0, len(m.packs))
	for id := range m.packs {
		packIDs = append(packIDs, id)
	}
	m.mu.RUnlock()

	results := make(map[string]error)
	for _, packID := range packIDs {
		results[packID] = m.ValidatePack(packID)
	}

	return results
}

// 转换为API响应格式
func (m *PackManager) ToAPIResponse() gin.H {
	return gin.H{
		"stats":       m.GetStats(),
		"subjects":    m.GetAllSubjects(),
		"levels":      m.GetAllLevels(),
		"subject_map": m.GetSubjectMap(),
		"packs":       m.GetAllPacks(),
	}
}

// 获取学科映射表
func (m *PackManager) GetSubjectMap() map[string]string {
	subjectMap := make(map[string]string)

	for subject := range m.bySubject {
		subjectMap[subject] = m.GetSubjectName(subject)
	}

	return subjectMap
}

// 搜索包
func (m *PackManager) SearchPacks(query string) []PackMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if query == "" {
		return []PackMetadata{}
	}

	query = strings.ToLower(query)
	results := make([]PackMetadata, 0)

	for _, pack := range m.packs {
		// 在名称、学科名称、年级中搜索
		if strings.Contains(strings.ToLower(pack.Name), query) ||
			strings.Contains(strings.ToLower(pack.SubjectName), query) ||
			strings.Contains(strings.ToLower(pack.Level), query) ||
			strings.Contains(strings.ToLower(pack.Subject), query) {
			results = append(results, *pack)
		}
	}

	return results
}

// 根据用户信息推荐包
func (m *PackManager) RecommendPacks(userInfo gin.H) []PackMetadata {
	// 从userInfo中提取推荐信息
	subject := userInfo["preferred_subject"].(string)
	level := userInfo["grade_level"].(string)

	// 优先推荐匹配的包
	if pack, err := m.GetPackBySubjectLevel(subject, level); err == nil {
		return []PackMetadata{*pack}
	}

	// 其次推荐同学科的其他年级
	packs := m.GetPacksBySubject(subject)
	if len(packs) > 0 {
		return packs[:min(3, len(packs))]
	}

	// 最后推荐默认包
	defaultPacks := m.GetPacksByLevel(m.config.DefaultLevel)
	if len(defaultPacks) > 0 {
		return defaultPacks[:min(3, len(defaultPacks))]
	}

	return []PackMetadata{}
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 将字符串转换为int64
func parseInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
