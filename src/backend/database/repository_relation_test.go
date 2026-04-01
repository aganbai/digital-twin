package database

import (
	"testing"
)

// ==================== ToggleRelation 测试 ====================

func TestToggleRelation_DisableActive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师和学生
	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_rel_toggle1", Password: "p", Role: "teacher", Nickname: "教师"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_rel_toggle1", Password: "p", Role: "student", Nickname: "学生"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身", School: "学校"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身"})

	// 创建 approved 关系
	relID, err := relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")
	if err != nil {
		t.Fatalf("创建关系失败: %v", err)
	}

	// 停用关系
	result, err := relationRepo.ToggleRelation(relID, 0)
	if err != nil {
		t.Fatalf("停用关系失败: %v", err)
	}

	if result.ID != relID {
		t.Errorf("关系ID不匹配: got %d, want %d", result.ID, relID)
	}
	if result.IsActive {
		t.Error("停用后 IsActive 应为 false")
	}

	// 验证数据库状态
	rel, err := relationRepo.GetByID(relID)
	if err != nil {
		t.Fatalf("查询关系失败: %v", err)
	}
	if rel.IsActive != 0 {
		t.Errorf("数据库中 is_active 应为 0: got %d", rel.IsActive)
	}
}

func TestToggleRelation_EnableInactive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_rel_toggle2", Password: "p", Role: "teacher", Nickname: "教师2"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_rel_toggle2", Password: "p", Role: "student", Nickname: "学生2"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身2", School: "学校2"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身2"})

	relID, _ := relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 先停用再启用
	_, _ = relationRepo.ToggleRelation(relID, 0)
	result, err := relationRepo.ToggleRelation(relID, 1)
	if err != nil {
		t.Fatalf("启用关系失败: %v", err)
	}

	if !result.IsActive {
		t.Error("启用后 IsActive 应为 true")
	}

	rel, _ := relationRepo.GetByID(relID)
	if rel.IsActive != 1 {
		t.Errorf("数据库中 is_active 应为 1: got %d", rel.IsActive)
	}
}

func TestToggleRelation_NonExistent(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)

	_, err := relationRepo.ToggleRelation(99999, 0)
	if err == nil {
		t.Fatal("不存在的关系应返回错误")
	}
}

func TestToggleRelation_StudentNickname(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_rel_nick", Password: "p", Role: "teacher", Nickname: "教师Nick"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_rel_nick", Password: "p", Role: "student", Nickname: "小明"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Nick", School: "学校Nick"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Nick"})

	relID, _ := relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	result, err := relationRepo.ToggleRelation(relID, 0)
	if err != nil {
		t.Fatalf("停用关系失败: %v", err)
	}

	// 验证返回了学生昵称
	if result.StudentNickname != "小明" {
		t.Errorf("学生昵称不匹配: got %q, want %q", result.StudentNickname, "小明")
	}
	if result.StudentPersonaID != studentPersonaID {
		t.Errorf("学生分身ID不匹配: got %d, want %d", result.StudentPersonaID, studentPersonaID)
	}
}

// ==================== CheckChatPermission 测试 ====================

func TestCheckChatPermission_AllActive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师和学生
	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_ok", Password: "p", Role: "teacher", Nickname: "教师Perm"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_ok", Password: "p", Role: "student", Nickname: "学生Perm"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Perm", School: "学校Perm"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Perm"})

	// 创建 approved 关系
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 创建班级并添加学生
	classID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "Perm班级"})
	_, _ = classRepo.AddMember(classID, studentPersonaID)

	// 所有条件满足，应返回 nil
	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr != nil {
		t.Fatalf("所有条件满足时不应返回错误: code=%d, msg=%s", permErr.Code, permErr.Message)
	}
}

func TestCheckChatPermission_NoRelation(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_norel", Password: "p", Role: "teacher", Nickname: "教师NoRel"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_norel", Password: "p", Role: "student", Nickname: "学生NoRel"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身NoRel", School: "学校NoRel"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身NoRel"})

	// 无关系
	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr == nil {
		t.Fatal("无关系时应返回错误")
	}
	if permErr.Code != 40007 {
		t.Errorf("错误码应为 40007: got %d", permErr.Code)
	}
}

func TestCheckChatPermission_PendingRelation(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_pending", Password: "p", Role: "teacher", Nickname: "教师Pending"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_pending", Password: "p", Role: "student", Nickname: "学生Pending"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Pending", School: "学校Pending"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Pending"})

	// 创建 pending 关系
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "pending", "student")

	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr == nil {
		t.Fatal("pending 关系应返回错误")
	}
	if permErr.Code != 40007 {
		t.Errorf("错误码应为 40007: got %d", permErr.Code)
	}
}

func TestCheckChatPermission_RelationInactive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_relinact", Password: "p", Role: "teacher", Nickname: "教师RelInact"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_relinact", Password: "p", Role: "student", Nickname: "学生RelInact"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身RelInact", School: "学校RelInact"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身RelInact"})

	relID, _ := relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 停用关系
	_, _ = relationRepo.ToggleRelation(relID, 0)

	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr == nil {
		t.Fatal("关系停用时应返回错误")
	}
	if permErr.Code != 40027 {
		t.Errorf("错误码应为 40027: got %d", permErr.Code)
	}
}

func TestCheckChatPermission_PersonaInactive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_pinact", Password: "p", Role: "teacher", Nickname: "教师PInact"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_pinact", Password: "p", Role: "student", Nickname: "学生PInact"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身PInact", School: "学校PInact"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身PInact"})

	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 停用教师分身
	personaRepo.SetActive(teacherPersonaID, 0)

	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr == nil {
		t.Fatal("教师分身停用时应返回错误")
	}
	if permErr.Code != 40025 {
		t.Errorf("错误码应为 40025: got %d", permErr.Code)
	}
}

func TestCheckChatPermission_AllClassesInactive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_classinact", Password: "p", Role: "teacher", Nickname: "教师ClassInact"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_classinact", Password: "p", Role: "student", Nickname: "学生ClassInact"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身ClassInact", School: "学校ClassInact"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身ClassInact"})

	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 创建班级，添加学生，然后停用班级
	classID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "停用班级Perm"})
	_, _ = classRepo.AddMember(classID, studentPersonaID)
	_, _ = classRepo.ToggleClass(classID, 0)

	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr == nil {
		t.Fatal("所有班级停用时应返回错误")
	}
	if permErr.Code != 40026 {
		t.Errorf("错误码应为 40026: got %d", permErr.Code)
	}
}

func TestCheckChatPermission_OneClassActiveOneInactive(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_mixed", Password: "p", Role: "teacher", Nickname: "教师Mixed"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_mixed", Password: "p", Role: "student", Nickname: "学生Mixed"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身Mixed", School: "学校Mixed"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Mixed"})

	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 创建两个班级，一个活跃一个停用
	class1ID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "活跃班级Mixed"})
	_, _ = classRepo.AddMember(class1ID, studentPersonaID)

	class2ID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "停用班级Mixed"})
	_, _ = classRepo.AddMember(class2ID, studentPersonaID)
	_, _ = classRepo.ToggleClass(class2ID, 0)

	// 只要有一个班级活跃，就应该允许对话
	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr != nil {
		t.Fatalf("有活跃班级时不应返回错误: code=%d, msg=%s", permErr.Code, permErr.Message)
	}
}

func TestCheckChatPermission_NoClass(t *testing.T) {
	db := setupTestDB(t)
	relationRepo := NewRelationRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_perm_noclass", Password: "p", Role: "teacher", Nickname: "教师NoClass"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_perm_noclass", Password: "p", Role: "student", Nickname: "学生NoClass"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身NoClass", School: "学校NoClass"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身NoClass"})

	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 没有班级时，只要关系和分身都活跃，就允许对话
	permErr := relationRepo.CheckChatPermission(teacherPersonaID, studentPersonaID)
	if permErr != nil {
		t.Fatalf("无班级时不应因班级检查而拒绝: code=%d, msg=%s", permErr.Code, permErr.Message)
	}
}
