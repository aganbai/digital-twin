package manager

import (
	"database/sql"
	"fmt"

	"digital-twin/src/harness/core"
	"digital-twin/src/plugins/auth"
	"digital-twin/src/plugins/dialogue"
	"digital-twin/src/plugins/knowledge"
	"digital-twin/src/plugins/memory"
)

// CreatePlugin 根据插件类型创建插件实例
func CreatePlugin(name string, pluginType string, db *sql.DB) (core.Plugin, error) {
	switch pluginType {
	case "auth":
		return auth.NewAuthPlugin(name, db), nil
	case "knowledge":
		return knowledge.NewKnowledgePlugin(name, db), nil
	case "memory":
		return memory.NewMemoryPlugin(name, db), nil
	case "dialogue":
		return dialogue.NewDialoguePlugin(name, db), nil
	default:
		return nil, fmt.Errorf("未知的插件类型: %s", pluginType)
	}
}
