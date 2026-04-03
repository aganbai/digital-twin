package database

import (
	"testing"
	"time"
)

// ==================== GetShareInfo 测试 ====================

func TestGetShareInfo_ValidShareCode(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_info", Password: "p", Role: "teacher", Nickname: "张老师", School: "XX中学", Description: "数学老师"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "张老师", School: "XX中学", Description: "数学老师"})

	// 创建通用分享码（target=0）
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "TESTCODE",
		TargetStudentPersonaID: 0,
		ExpiresAt:              &expiresAt,
		MaxUses:                0,
	}
	_, err := shareRepo.Create(share)
	if err != nil {
		t.Fatalf("创建分享码失败: %v", err)
	}

	// 查询分享码信息
	info, err := shareRepo.GetShareInfo("TESTCODE")
	if err != nil {
		t.Fatalf("查询分享码信息失败: %v", err)
	}

	if info == nil {
		t.Fatal("分享码信息不应为 nil")
	}
	if !info.IsValid {
		t.Errorf("分享码应有效，reason: %s", info.Reason)
	}
	if info.TeacherPersonaID != teacherPersonaID {
		t.Errorf("教师分身ID不匹配: got %d, want %d", info.TeacherPersonaID, teacherPersonaID)
	}
	if info.TeacherNickname != "张老师" {
		t.Errorf("教师昵称不匹配: got %q, want %q", info.TeacherNickname, "张老师")
	}
	if info.TeacherSchool != "XX中学" {
		t.Errorf("教师学校不匹配: got %q, want %q", info.TeacherSchool, "XX中学")
	}
	if info.TargetStudentPersonaID != 0 {
		t.Errorf("通用分享码 target 应为 0: got %d", info.TargetStudentPersonaID)
	}
}

func TestGetShareInfo_NotExist(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)

	info, err := shareRepo.GetShareInfo("NOTEXIST")
	if err != nil {
		t.Fatalf("查询不应返回错误: %v", err)
	}
	if info == nil {
		t.Fatal("info 不应为 nil")
	}
	if info.IsValid {
		t.Error("不存在的分享码应标记为无效")
	}
	if info.Reason != "分享码不存在" {
		t.Errorf("原因不匹配: got %q", info.Reason)
	}
}

func TestGetShareInfo_Expired(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_expired", Password: "p", Role: "teacher", Nickname: "教师Expired"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师Expired", School: "学校"})

	// 创建已过期的分享码
	expiresAt := time.Now().Add(-1 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "EXPIRED1",
		ExpiresAt:        &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	info, err := shareRepo.GetShareInfo("EXPIRED1")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if info.IsValid {
		t.Error("过期分享码应标记为无效")
	}
	if info.Reason != "分享码已过期" {
		t.Errorf("原因不匹配: got %q", info.Reason)
	}
}

func TestGetShareInfo_Deactivated(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_deact", Password: "p", Role: "teacher", Nickname: "教师Deact"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师Deact", School: "学校"})

	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "DEACTIV1",
		ExpiresAt:        &expiresAt,
	}
	shareID, _ := shareRepo.Create(share)

	// 停用分享码
	_ = shareRepo.Deactivate(shareID)

	info, err := shareRepo.GetShareInfo("DEACTIV1")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if info.IsValid {
		t.Error("停用的分享码应标记为无效")
	}
	if info.Reason != "分享码已停用" {
		t.Errorf("原因不匹配: got %q", info.Reason)
	}
}

func TestGetShareInfo_WithTargetStudent(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_target", Password: "p", Role: "teacher", Nickname: "张老师"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "张老师", School: "XX中学", Description: "数学老师"})

	studentUserID, _ := userRepo.Create(&User{Username: "student_share_target", Password: "p", Role: "student", Nickname: "李四"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "李四"})

	// 创建定向分享码
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "TARGET01",
		TargetStudentPersonaID: studentPersonaID,
		ExpiresAt:              &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	info, err := shareRepo.GetShareInfo("TARGET01")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if !info.IsValid {
		t.Error("分享码应有效")
	}
	if info.TargetStudentPersonaID != studentPersonaID {
		t.Errorf("目标学生ID不匹配: got %d, want %d", info.TargetStudentPersonaID, studentPersonaID)
	}
	if info.TargetStudentNickname != "李四" {
		t.Errorf("目标学生昵称不匹配: got %q, want %q", info.TargetStudentNickname, "李四")
	}
}

func TestGetShareInfo_MaxUsesReached(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_maxuse", Password: "p", Role: "teacher", Nickname: "教师MaxUse"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师MaxUse", School: "学校"})

	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "MAXUSE01",
		ExpiresAt:        &expiresAt,
		MaxUses:          1,
	}
	shareID, _ := shareRepo.Create(share)

	// 使用一次
	_ = shareRepo.IncrementUsedCount(shareID)

	info, err := shareRepo.GetShareInfo("MAXUSE01")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if info.IsValid {
		t.Error("使用次数达上限的分享码应标记为无效")
	}
	if info.Reason != "分享码使用次数已达上限" {
		t.Errorf("原因不匹配: got %q", info.Reason)
	}
}

// ==================== IsApprovedByPersonas 用于 join_status 判断的测试 ====================

func TestJoinStatus_AlreadyJoined(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_joined", Password: "p", Role: "teacher", Nickname: "教师Joined"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_joined", Password: "p", Role: "student", Nickname: "学生Joined"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Joined", School: "学校"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Joined"})

	// 创建 approved 关系
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 创建分享码
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "JOINED01",
		ExpiresAt:        &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	// 模拟 join_status 判断逻辑
	shareRecord, _ := shareRepo.GetByCode("JOINED01")
	if shareRecord == nil {
		t.Fatal("分享码不应为 nil")
	}

	approved, err := relationRepo.IsApprovedByPersonas(shareRecord.TeacherPersonaID, studentPersonaID)
	if err != nil {
		t.Fatalf("查询关系失败: %v", err)
	}
	if !approved {
		t.Error("已有 approved 关系，应返回 true")
	}

	// join_status 应为 already_joined
	joinStatus := determineJoinStatus(shareRecord, studentPersonaID, approved)
	if joinStatus != "already_joined" {
		t.Errorf("join_status 应为 already_joined: got %q", joinStatus)
	}
}

func TestJoinStatus_CanJoin_UniversalCode(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_canjoin", Password: "p", Role: "teacher", Nickname: "教师CanJoin"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_canjoin", Password: "p", Role: "student", Nickname: "学生CanJoin"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身CanJoin", School: "学校"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身CanJoin"})

	// 创建通用分享码（target=0）
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "CANJOIN1",
		TargetStudentPersonaID: 0,
		ExpiresAt:              &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	shareRecord, _ := shareRepo.GetByCode("CANJOIN1")
	approved, _ := relationRepo.IsApprovedByPersonas(shareRecord.TeacherPersonaID, studentPersonaID)

	joinStatus := determineJoinStatus(shareRecord, studentPersonaID, approved)
	if joinStatus != "can_join" {
		t.Errorf("通用分享码 + 无关系应为 can_join: got %q", joinStatus)
	}
}

func TestJoinStatus_CanJoin_TargetMatch(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_target_match", Password: "p", Role: "teacher", Nickname: "教师Match"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_target_match", Password: "p", Role: "student", Nickname: "学生Match"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Match", School: "学校"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Match"})

	// 创建定向分享码，目标就是当前学生
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "MATCH001",
		TargetStudentPersonaID: studentPersonaID,
		ExpiresAt:              &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	shareRecord, _ := shareRepo.GetByCode("MATCH001")
	approved, _ := relationRepo.IsApprovedByPersonas(shareRecord.TeacherPersonaID, studentPersonaID)

	joinStatus := determineJoinStatus(shareRecord, studentPersonaID, approved)
	if joinStatus != "can_join" {
		t.Errorf("定向分享码 + 目标匹配应为 can_join: got %q", joinStatus)
	}
}

func TestJoinStatus_NotTarget(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_not_target", Password: "p", Role: "teacher", Nickname: "教师NotTarget"})
	targetStudentUserID, _ := userRepo.Create(&User{Username: "target_student", Password: "p", Role: "student", Nickname: "目标学生"})
	otherStudentUserID, _ := userRepo.Create(&User{Username: "other_student", Password: "p", Role: "student", Nickname: "其他学生"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身NotTarget", School: "学校"})
	targetStudentPersonaID, _ := personaRepo.Create(&Persona{UserID: targetStudentUserID, Role: "student", Nickname: "目标学生分身"})
	otherStudentPersonaID, _ := personaRepo.Create(&Persona{UserID: otherStudentUserID, Role: "student", Nickname: "其他学生分身"})
	_ = targetStudentPersonaID // 仅用于分享码创建

	// 创建定向分享码，目标是 targetStudent
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "NOTTGT01",
		TargetStudentPersonaID: targetStudentPersonaID,
		ExpiresAt:              &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	shareRecord, _ := shareRepo.GetByCode("NOTTGT01")
	// otherStudent 尝试查看
	approved, _ := relationRepo.IsApprovedByPersonas(shareRecord.TeacherPersonaID, otherStudentPersonaID)

	joinStatus := determineJoinStatus(shareRecord, otherStudentPersonaID, approved)
	if joinStatus != "not_target" {
		t.Errorf("定向分享码 + 非目标学生应为 not_target: got %q", joinStatus)
	}
}

// determineJoinStatus 辅助函数：模拟 handler 中的 join_status 判断逻辑
// 用于测试验证逻辑正确性
func determineJoinStatus(share *PersonaShare, studentPersonaID int64, approved bool) string {
	if approved {
		return "already_joined"
	}
	if share.TargetStudentPersonaID > 0 && share.TargetStudentPersonaID != studentPersonaID {
		return "not_target"
	}
	return "can_join"
}

// ==================== GetByCode 测试 ====================

func TestGetByCode_Success(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_getbycode", Password: "p", Role: "teacher", Nickname: "教师GetByCode"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身GetByCode", School: "学校"})

	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID:       teacherPersonaID,
		ShareCode:              "GETCODE1",
		TargetStudentPersonaID: 0,
		ExpiresAt:              &expiresAt,
		MaxUses:                10,
	}
	_, _ = shareRepo.Create(share)

	result, err := shareRepo.GetByCode("GETCODE1")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if result == nil {
		t.Fatal("结果不应为 nil")
	}
	if result.ShareCode != "GETCODE1" {
		t.Errorf("分享码不匹配: got %q", result.ShareCode)
	}
	if result.TeacherPersonaID != teacherPersonaID {
		t.Errorf("教师分身ID不匹配: got %d, want %d", result.TeacherPersonaID, teacherPersonaID)
	}
	if result.MaxUses != 10 {
		t.Errorf("最大使用次数不匹配: got %d, want 10", result.MaxUses)
	}
	if result.UsedCount != 0 {
		t.Errorf("初始使用次数应为 0: got %d", result.UsedCount)
	}
}

func TestGetByCode_NotFound(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)

	result, err := shareRepo.GetByCode("NOCODE01")
	if err != nil {
		t.Fatalf("查询不应返回错误: %v", err)
	}
	if result != nil {
		t.Error("不存在的分享码应返回 nil")
	}
}

// ==================== IncrementUsedCount 测试 ====================

func TestIncrementUsedCount(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_incr", Password: "p", Role: "teacher", Nickname: "教师Incr"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Incr", School: "学校"})

	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "INCR0001",
		ExpiresAt:        &expiresAt,
	}
	shareID, _ := shareRepo.Create(share)

	// 增加使用次数
	err := shareRepo.IncrementUsedCount(shareID)
	if err != nil {
		t.Fatalf("增加使用次数失败: %v", err)
	}

	// 验证
	result, _ := shareRepo.GetByCode("INCR0001")
	if result.UsedCount != 1 {
		t.Errorf("使用次数应为 1: got %d", result.UsedCount)
	}

	// 再增加一次
	_ = shareRepo.IncrementUsedCount(shareID)
	result, _ = shareRepo.GetByCode("INCR0001")
	if result.UsedCount != 2 {
		t.Errorf("使用次数应为 2: got %d", result.UsedCount)
	}
}

// ==================== GetShareInfo 与 Class 联合测试 ====================

func TestGetShareInfo_WithClass(t *testing.T) {
	db := setupTestDB(t)
	shareRepo := NewShareRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_share_class", Password: "p", Role: "teacher", Nickname: "张老师"})
	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "张老师", School: "XX中学", Description: "数学老师"})

	// 创建班级
	classID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "高一1班"})

	// 创建带班级的分享码
	expiresAt := time.Now().Add(24 * time.Hour)
	share := &PersonaShare{
		TeacherPersonaID: teacherPersonaID,
		ShareCode:        "CLASS001",
		ClassID:          &classID,
		ExpiresAt:        &expiresAt,
	}
	_, _ = shareRepo.Create(share)

	info, err := shareRepo.GetShareInfo("CLASS001")
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if !info.IsValid {
		t.Error("分享码应有效")
	}
	if info.ClassName != "高一1班" {
		t.Errorf("班级名称不匹配: got %q, want %q", info.ClassName, "高一1班")
	}
}
