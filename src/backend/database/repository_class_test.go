package database

import (
	"testing"
)

// ==================== ToggleClass 测试 ====================

func TestToggleClass_DisableActive(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_toggle_class",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师A",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身A",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 创建班级
	classID, err := classRepo.Create(&Class{
		PersonaID:   personaID,
		Name:        "测试班级",
		Description: "用于测试启停",
	})
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}

	// 停用班级
	result, err := classRepo.ToggleClass(classID, 0)
	if err != nil {
		t.Fatalf("停用班级失败: %v", err)
	}

	if result.ID != classID {
		t.Errorf("班级ID不匹配: got %d, want %d", result.ID, classID)
	}
	if result.Name != "测试班级" {
		t.Errorf("班级名称不匹配: got %q, want %q", result.Name, "测试班级")
	}
	if result.IsActive {
		t.Error("停用后 IsActive 应为 false")
	}
	if result.AffectedStudents != 0 {
		t.Errorf("无成员时 AffectedStudents 应为 0: got %d", result.AffectedStudents)
	}

	// 验证数据库中的状态
	class, err := classRepo.GetByID(classID)
	if err != nil {
		t.Fatalf("查询班级失败: %v", err)
	}
	if class.IsActive != 0 {
		t.Errorf("数据库中 is_active 应为 0: got %d", class.IsActive)
	}
}

func TestToggleClass_EnableInactive(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_toggle_enable",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师B",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身B",
		School:   "测试学校B",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 创建班级并先停用
	classID, err := classRepo.Create(&Class{
		PersonaID:   personaID,
		Name:        "停用班级",
		Description: "先停用再启用",
	})
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}

	_, err = classRepo.ToggleClass(classID, 0)
	if err != nil {
		t.Fatalf("停用班级失败: %v", err)
	}

	// 重新启用
	result, err := classRepo.ToggleClass(classID, 1)
	if err != nil {
		t.Fatalf("启用班级失败: %v", err)
	}

	if !result.IsActive {
		t.Error("启用后 IsActive 应为 true")
	}

	// 验证数据库中的状态
	class, err := classRepo.GetByID(classID)
	if err != nil {
		t.Fatalf("查询班级失败: %v", err)
	}
	if class.IsActive != 1 {
		t.Errorf("数据库中 is_active 应为 1: got %d", class.IsActive)
	}
}

func TestToggleClass_WithMembers(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师
	teacherUserID, err := userRepo.Create(&User{
		Username: "teacher_toggle_members",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师C",
	})
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	teacherPersonaID, err := personaRepo.Create(&Persona{
		UserID:   teacherUserID,
		Role:     "teacher",
		Nickname: "教师分身C",
		School:   "测试学校C",
	})
	if err != nil {
		t.Fatalf("创建教师分身失败: %v", err)
	}

	// 创建学生
	studentUserID, err := userRepo.Create(&User{
		Username: "student_toggle_members",
		Password: "password",
		Role:     "student",
		Nickname: "学生A",
	})
	if err != nil {
		t.Fatalf("创建学生失败: %v", err)
	}

	studentPersonaID, err := personaRepo.Create(&Persona{
		UserID:   studentUserID,
		Role:     "student",
		Nickname: "学生分身A",
	})
	if err != nil {
		t.Fatalf("创建学生分身失败: %v", err)
	}

	// 创建班级并添加成员
	classID, err := classRepo.Create(&Class{
		PersonaID:   teacherPersonaID,
		Name:        "有成员班级",
		Description: "含成员的班级",
	})
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}

	_, err = classRepo.AddMember(classID, studentPersonaID)
	if err != nil {
		t.Fatalf("添加成员失败: %v", err)
	}

	// 停用班级
	result, err := classRepo.ToggleClass(classID, 0)
	if err != nil {
		t.Fatalf("停用班级失败: %v", err)
	}

	if result.AffectedStudents != 1 {
		t.Errorf("AffectedStudents 应为 1: got %d", result.AffectedStudents)
	}
	if result.IsActive {
		t.Error("停用后 IsActive 应为 false")
	}
}

func TestToggleClass_NonExistentClass(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)

	// 不存在的班级ID
	_, err := classRepo.ToggleClass(99999, 0)
	if err == nil {
		t.Fatal("不存在的班级应返回错误")
	}
}

func TestToggleClass_IdempotentDisable(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_toggle_idempotent",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师D",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身D",
		School:   "测试学校D",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	classID, err := classRepo.Create(&Class{
		PersonaID:   personaID,
		Name:        "幂等测试班级",
		Description: "测试重复操作",
	})
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}

	// 连续两次停用，不应报错
	_, err = classRepo.ToggleClass(classID, 0)
	if err != nil {
		t.Fatalf("第一次停用失败: %v", err)
	}

	result, err := classRepo.ToggleClass(classID, 0)
	if err != nil {
		t.Fatalf("第二次停用失败: %v", err)
	}
	if result.IsActive {
		t.Error("第二次停用后 IsActive 仍应为 false")
	}
}

func TestToggleClass_ListReflectsStatus(t *testing.T) {
	db := setupTestDB(t)
	classRepo := NewClassRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_toggle_list",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师E",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身E",
		School:   "测试学校E",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 创建两个班级
	class1ID, err := classRepo.Create(&Class{
		PersonaID:   personaID,
		Name:        "活跃班级",
		Description: "保持活跃",
	})
	if err != nil {
		t.Fatalf("创建班级1失败: %v", err)
	}

	class2ID, err := classRepo.Create(&Class{
		PersonaID:   personaID,
		Name:        "停用班级",
		Description: "将被停用",
	})
	if err != nil {
		t.Fatalf("创建班级2失败: %v", err)
	}

	// 停用班级2
	_, err = classRepo.ToggleClass(class2ID, 0)
	if err != nil {
		t.Fatalf("停用班级2失败: %v", err)
	}

	// 查询列表，验证状态
	classes, err := classRepo.ListByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询班级列表失败: %v", err)
	}

	if len(classes) != 2 {
		t.Fatalf("应有2个班级: got %d", len(classes))
	}

	// 按创建时间排序，class1 在前
	for _, c := range classes {
		if c.ID == class1ID && !c.IsActive {
			t.Error("班级1应为活跃状态")
		}
		if c.ID == class2ID && c.IsActive {
			t.Error("班级2应为停用状态")
		}
	}
}
