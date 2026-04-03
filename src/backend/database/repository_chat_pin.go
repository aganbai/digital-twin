package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ChatPinRepository 聊天置顶数据访问
type ChatPinRepository struct {
	db *Database
}

// NewChatPinRepository 创建置顶仓库
func NewChatPinRepository(db *Database) *ChatPinRepository {
	return &ChatPinRepository{db: db}
}

// CreateChatPin 创建置顶记录
func (r *ChatPinRepository) CreateChatPin(pin *ChatPin) (int64, error) {
	result, err := r.db.DB.Exec(`
		INSERT INTO chat_pins (user_id, user_role, target_type, target_id, persona_id, pinned_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, user_role, target_type, target_id, persona_id) DO UPDATE SET
			pinned_at = excluded.pinned_at`,
		pin.UserID, pin.UserRole, pin.TargetType, pin.TargetID, pin.PersonaID, time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("创建置顶记录失败: %w", err)
	}
	return result.LastInsertId()
}

// DeleteChatPin 删除置顶记录
func (r *ChatPinRepository) DeleteChatPin(userID int64, userRole string, targetType string, targetID int64, personaID int64) error {
	_, err := r.db.DB.Exec(`
		DELETE FROM chat_pins
		WHERE user_id = ? AND user_role = ? AND target_type = ? AND target_id = ? AND persona_id = ?`,
		userID, userRole, targetType, targetID, personaID,
	)
	if err != nil {
		return fmt.Errorf("删除置顶记录失败: %w", err)
	}
	return nil
}

// GetChatPinsByUser 获取用户的所有置顶记录
func (r *ChatPinRepository) GetChatPinsByUser(userID int64, userRole string, personaID int64) ([]ChatPin, error) {
	rows, err := r.db.DB.Query(`
		SELECT id, user_id, user_role, target_type, target_id, persona_id, pinned_at
		FROM chat_pins
		WHERE user_id = ? AND user_role = ? AND persona_id = ?
		ORDER BY pinned_at DESC`,
		userID, userRole, personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询置顶记录失败: %w", err)
	}
	defer rows.Close()

	var pins []ChatPin
	for rows.Next() {
		var pin ChatPin
		err := rows.Scan(&pin.ID, &pin.UserID, &pin.UserRole, &pin.TargetType, &pin.TargetID, &pin.PersonaID, &pin.PinnedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描置顶记录失败: %w", err)
		}
		pins = append(pins, pin)
	}

	return pins, nil
}

// IsPinned 检查是否已置顶
func (r *ChatPinRepository) IsPinned(userID int64, userRole string, targetType string, targetID int64, personaID int64) (bool, error) {
	var count int
	err := r.db.DB.QueryRow(`
		SELECT COUNT(*) FROM chat_pins
		WHERE user_id = ? AND user_role = ? AND target_type = ? AND target_id = ? AND persona_id = ?`,
		userID, userRole, targetType, targetID, personaID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查置顶状态失败: %w", err)
	}
	return count > 0, nil
}

// GetPinnedTargets 获取用户置顶的目标ID列表
func (r *ChatPinRepository) GetPinnedTargets(userID int64, userRole string, targetType string, personaID int64) ([]int64, error) {
	rows, err := r.db.DB.Query(`
		SELECT target_id FROM chat_pins
		WHERE user_id = ? AND user_role = ? AND target_type = ? AND persona_id = ?
		ORDER BY pinned_at DESC`,
		userID, userRole, targetType, personaID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询置顶目标失败: %w", err)
	}
	defer rows.Close()

	var targetIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("扫描置顶目标失败: %w", err)
		}
		targetIDs = append(targetIDs, id)
	}

	return targetIDs, nil
}

// GetChatPinByID 根据ID获取置顶记录
func (r *ChatPinRepository) GetChatPinByID(id int64) (*ChatPin, error) {
	row := r.db.DB.QueryRow(`
		SELECT id, user_id, user_role, target_type, target_id, persona_id, pinned_at
		FROM chat_pins WHERE id = ?`, id)

	pin := &ChatPin{}
	err := row.Scan(&pin.ID, &pin.UserID, &pin.UserRole, &pin.TargetType, &pin.TargetID, &pin.PersonaID, &pin.PinnedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询置顶记录失败: %w", err)
	}
	return pin, nil
}

// DeleteChatPinByID 根据ID删除置顶记录
func (r *ChatPinRepository) DeleteChatPinByID(id int64) error {
	_, err := r.db.DB.Exec(`DELETE FROM chat_pins WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("删除置顶记录失败: %w", err)
	}
	return nil
}
