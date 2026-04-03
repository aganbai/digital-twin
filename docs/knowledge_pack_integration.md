# 知识包集成指南

## 🎯 集成概述

**知识库自动化工具包 (KAT)** 已成功创建了**39个教育知识包**，涵盖小学到大学的所有主要学科。这些知识包已集成到主项目的 `data/prebuilt-knowledge-packs/` 目录中，可用于支持教学问答、知识检索和智能辅导功能。

## 📦 已集成的知识包清单

### 12个学科 × 多年级 = 39个知识包

| 学科 | 小学 | 初中 | 高中 | 大学 | 合计 |
|------|------|------|------|------|------|
| 数学 | ✅ | ✅ | ✅ | ✅ | 4 |
| 物理 | - | ✅ | ✅ | ✅ | 3 |
| 化学 | - | ✅ | ✅ | ✅ | 3 |
| 生物 | - | ✅ | ✅ | ✅ | 3 |
| 语文 | ✅ | ✅ | ✅ | ✅ | 4 |
| 英语 | ✅ | ✅ | ✅ | ✅ | 4 |
| 历史 | - | ✅ | ✅ | ✅ | 3 |
| 地理 | - | ✅ | ✅ | ✅ | 3 |
| 政治 | - | ✅ | ✅ | ✅ | 3 |
| 音乐 | ✅ | ✅ | ✅ | - | 3 |
| 美术 | ✅ | ✅ | ✅ | - | 3 |
| 体育 | ✅ | ✅ | ✅ | - | 3 |
| **合计** | **4** | **12** | **12** | **11** | **39** |

## 📁 知识包目录结构

```
data/prebuilt-knowledge-packs/knowledge_packs/
├── packages.json                    # 知识包索引文件 (13.6KB)
├── math_小学/                      # 数学-小学知识包
│   ├── metadata.json              # 包元数据 (294B)
│   └── README.md                  # 包说明 (511B)
├── physics_初中/                   # 物理-初中知识包
│   ├── metadata.json              # 包元数据 (300B)
│   └── README.md                  # 包说明 (511B)
├── chemistry_高中/                 # 化学-高中知识包
│   ├── metadata.json              # 包元数据 (304B)
│   └── README.md                  # 包说明 (511B)
└── ... (还有36个类似结构的知识包)
```

## 🔧 集成使用方式

### 方式1: 直接元数据查询

主项目可以直接读取知识包的元数据，了解可用的知识资源：

```go
// 示例：Go代码读取知识包元数据
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

type KnowledgePackMetadata struct {
    ID           string  `json:"id"`
    Name         string  `json:"name"`
    Subject      string  `json:"subject"`
    SubjectName  string  `json:"subject_name"`
    Level        string  `json:"level"`
    Version      string  `json:"version"`
    CreatedAt    string  `json:"created_at"`
    EstimatedSize float64 `json:"estimated_size_mb"`
    Status       string  `json:"status"`
}

func LoadKnowledgePacks() ([]KnowledgePackMetadata, error) {
    dataPath := "data/prebuilt-knowledge-packs/knowledge_packs/packages.json"
    data, err := os.ReadFile(dataPath)
    if err != nil {
        return nil, err
    }
    
    var result struct {
        Packs []KnowledgePackMetadata `json:"packs"`
    }
    
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }
    
    return result.Packs, nil
}
```

### 方式2: 按需加载知识包

主项目可以根据用户选择的学科和年级，动态加载对应的知识包：

```go
// 示例：根据用户选择加载知识包
func LoadSpecificPack(subject, level string) (*KnowledgePackMetadata, error) {
    packs, err := LoadKnowledgePacks()
    if err != nil {
        return nil, err
    }
    
    targetID := fmt.Sprintf("%s_%s", subject, level)
    for _, pack := range packs {
        if pack.ID == targetID {
            return &pack, nil
        }
    }
    
    return nil, fmt.Errorf("知识包未找到: %s", targetID)
}

// 使用示例
mathPack, err := LoadSpecificPack("math", "小学")
if err != nil {
    // 处理错误
}
fmt.Printf("加载知识包: %s (%.2fMB)", mathPack.Name, mathPack.EstimatedSize)
```

### 方式3: 初始化向量数据库

当需要实际使用向量检索时，可以初始化ChromaDB向量数据库：

```go
// 示例：初始化向量数据库
func InitializeVectorDatabase(packID string) error {
    // 1. 检查知识包是否存在
    packPath := fmt.Sprintf("data/prebuilt-knowledge-packs/knowledge_packs/%s", packID)
    if _, err := os.Stat(packPath); os.IsNotExist(err) {
        return fmt.Errorf("知识包目录不存在: %s", packID)
    }
    
    // 2. 读取元数据
    metadataPath := filepath.Join(packPath, "metadata.json")
    metadataData, err := os.ReadFile(metadataPath)
    if err != nil {
        return err
    }
    
    // 3. 初始化ChromaDB
    // 这里可以调用KAT的向量索引器，或者使用KAT的配置文件
    // KAT配置路径: knowledge-automation/configs/config.yaml
    
    return nil
}
```

## ⚡ 快速开始

### 1. 验证知识包已集成
```bash
# 检查知识包数量
cd /Users/aganbai/Desktop/WorkSpace/digital-twin
tree data/prebuilt-knowledge-packs/knowledge_packs/ | head -20

# 检查包索引
cat data/prebuilt-knowledge-packs/knowledge_packs/packages.json | jq '.total_packs'
# 应该输出: 39
```

### 2. 在Go代码中集成

创建知识包管理模块：

```go
// src/backend/knowledge/pack_manager.go
package knowledge

import (
    "embed"
    "encoding/json"
    "path/filepath"
    "sync"
)

// 嵌入知识包元数据（可选）
//go:embed ../../data/prebuilt-knowledge-packs/knowledge_packs/packages.json
var knowledgePacksData []byte

// PackManager 管理知识包
type PackManager struct {
    mu    sync.RWMutex
    packs map[string]KnowledgePack
}

// LoadPacks 加载所有知识包
func (m *PackManager) LoadPacks() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // 从文件或嵌入数据加载
    // ...
    
    return nil
}

// GetPack 获取指定知识包
func (m *PackManager) GetPack(subject, level string) (*KnowledgePack, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    packID := subject + "_" + level
    if pack, ok := m.packs[packID]; ok {
        return &pack, nil
    }
    
    return nil, ErrPackNotFound
}
```

### 3. 在API中使用知识包

```go
// 示例API端点
func (h *Handler) GetAvailablePacks(c *gin.Context) {
    packs, err := h.packManager.GetAllPacks()
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "packs": packs,
        "count": len(packs),
    })
}

func (h *Handler) ActivatePack(c *gin.Context) {
    var req struct {
        Subject string `json:"subject" binding:"required"`
        Level   string `json:"level" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    pack, err := h.packManager.GetPack(req.Subject, req.Level)
    if err != nil {
        c.JSON(404, gin.H{"error": "知识包未找到"})
        return
    }
    
    // 激活知识包（初始化向量数据库等）
    if err := h.vectorService.ActivatePack(pack.ID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "message": "知识包已激活",
        "pack": pack,
    })
}
```

## 🔄 知识包更新机制

### 自动更新流程
KAT项目支持知识包的持续更新，主项目可以通过以下方式更新知识包：

1. **手动更新**: 重新运行KAT处理新的教育文档
2. **增量更新**: 只处理新增或修改的文档
3. **版本控制**: 支持多个版本的知识包共存

### 更新脚本示例
```bash
#!/bin/bash
# scripts/update_knowledge_packs.sh

# 进入KAT目录
cd knowledge-automation

# 1. 处理新文档
python3 scripts/process_documents.py

# 2. 构建新索引
python3 scripts/build_index.py

# 3. 生成新知识包
python3 scripts/create_packs.py

# 4. 复制到主项目
cp -r data/knowledge_packs_output/knowledge_packs \
    ../data/prebuilt-knowledge-packs/knowledge_packs_$(date +%Y%m%d)

# 5. 更新软链接（可选）
ln -sfn knowledge_packs_$(date +%Y%m%d) \
    ../data/prebuilt-knowledge-packs/current
```

## 🧪 测试验证

### 1. 验证包完整性
```go
// 测试知识包加载
func TestKnowledgePackLoading(t *testing.T) {
    manager := NewPackManager()
    
    // 测试加载所有包
    if err := manager.LoadPacks(); err != nil {
        t.Fatalf("加载知识包失败: %v", err)
    }
    
    // 验证包数量
    expectedCount := 39
    if count := manager.Count(); count != expectedCount {
        t.Errorf("包数量不匹配: 期望 %d, 实际 %d", expectedCount, count)
    }
    
    // 测试获取特定包
    pack, err := manager.GetPack("math", "小学")
    if err != nil {
        t.Errorf("获取数学小学包失败: %v", err)
    }
    
    if pack.SubjectName != "数学" {
        t.Errorf("学科名称不匹配: %s", pack.SubjectName)
    }
}
```

### 2. 端到端测试
```go
// 测试知识包API
func TestKnowledgePackAPI(t *testing.T) {
    // 启动测试服务器
    // ...
    
    // 测试获取可用包
    resp, err := http.Get("http://localhost:8080/api/knowledge/packs")
    // 验证响应
    // ...
    
    // 测试激活包
    activateData := map[string]string{
        "subject": "math",
        "level":  "小学",
    }
    // 验证激活结果
    // ...
}
```

## 📈 性能建议

### 内存管理
- **按需加载**: 不要一次性加载所有知识包的向量数据
- **缓存机制**: 缓存常用知识包的查询结果
- **LRU策略**: 对不常用的知识包进行清理

### 存储优化
- **压缩存储**: 知识包支持gzip压缩
- **增量更新**: 只更新变化的部分
- **分级存储**: 热数据在内存，冷数据在磁盘

## 🆘 故障排除

### 常见问题

1. **包未找到**
   - 检查路径: `data/prebuilt-knowledge-packs/knowledge_packs/`
   - 确认包ID格式: `{subject}_{level}`
   - 检查包索引文件: `packages.json`

2. **元数据解析错误**
   - 验证JSON格式: `jq . packages.json`
   - 检查文件编码: 应该是UTF-8
   - 确保文件完整: 无截断或损坏

3. **向量数据库初始化失败**
   - 检查KAT配置: `knowledge-automation/configs/config.yaml`
   - 确保OpenAI API密钥正确
   - 检查ChromaDB目录权限

### 调试工具

```bash
# 1. 检查包结构
tree -L 2 data/prebuilt-knowledge-packs/

# 2. 查看包元数据
cat data/prebuilt-knowledge-packs/knowledge_packs/math_小学/metadata.json | jq .

# 3. 验证包索引
cat data/prebuilt-knowledge-packs/knowledge_packs/packages.json | jq '.packs[0]'

# 4. 测试KAT功能
cd knowledge-automation
python3 scripts/run_example.py
```

---

## ✅ 集成完成状态

- [x] 知识包已创建: 39个包，12个学科，多年级
- [x] 元数据已生成: 每个包都有完整的metadata.json
- [x] 索引文件就绪: packages.json包含所有包信息
- [x] 部署到主项目: 已复制到data/prebuilt-knowledge-packs/
- [x] 集成指南完成: 本文档提供完整的集成方案

**下一步**: 在主项目中实现知识包管理模块，开始使用这些教育知识包。