package database

import (
	"testing"
)

// ==================== GetPersonaDashboard 测试 ====================

func TestGetPersonaDashboard_BasicInfo(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, _ := userRepo.Create(&User{
		Username: "teacher_dashboard1",
		Password: "password",
		Role:     "teacher",
		Nickname: "仪表盘教师",
		School:   "仪表盘学校",
	})

	personaID, _ := personaRepo.Create(&Persona{
		UserID:      userID,
		Role:        "teacher",
		Nickname:    "仪表盘教师分身",
		School:      "仪表盘学校",
		Description: "测试仪表盘",
	})

	dashboard, err := personaRepo.GetPersonaDashboard(personaID)
	if err != nil {
		t.Fatalf("获取仪表盘失败: %v", err)
	}

	// 验证分身基本信息
	if dashboard.Persona == nil {
		t.Fatal("分身信息不应为 nil")
	}
	if dashboard.Persona.ID != personaID {
		t.Errorf("分身ID不匹配: got %d, want %d", dashboard.Persona.ID, personaID)
	}
	if dashboard.Persona.Nickname != "仪表盘教师分身" {
		t.Errorf("昵称不匹配: got %q", dashboard.Persona.Nickname)
	}

	// 验证初始统计
	if dashboard.PendingCount != 0 {
		t.Errorf("初始 PendingCount 应为 0: got %d", dashboard.PendingCount)
	}
	if len(dashboard.Classes) != 0 {
		t.Errorf("初始班级列表应为空: got %d", len(dashboard.Classes))
	}
}

func TestGetPersonaDashboard_WithPendingRelations(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_dash_pending", Password: "p", Role: "teacher", Nickname: "教师DashPending"})
	student1UserID, _ := userRepo.Create(&User{Username: "student_dash_pending1", Password: "p", Role: "student", Nickname: "学生1"})
	student2UserID, _ := userRepo.Create(&User{Username: "student_dash_pending2", Password: "p", Role: "student", Nickname: "学生2"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身DashPending", School: "学校DashPending"})
	student1PersonaID, _ := personaRepo.Create(&Persona{UserID: student1UserID, Role: "student", Nickname: "学生分身1"})
	student2PersonaID, _ := personaRepo.Create(&Persona{UserID: student2UserID, Role: "student", Nickname: "学生分身2"})

	// 创建2个 pending 关系
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, student1UserID, teacherPersonaID, student1PersonaID, "pending", "student")
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, student2UserID, teacherPersonaID, student2PersonaID, "pending", "student")

	dashboard, err := personaRepo.GetPersonaDashboard(teacherPersonaID)
	if err != nil {
		t.Fatalf("获取仪表盘失败: %v", err)
	}

	if dashboard.PendingCount != 2 {
		t.Errorf("PendingCount 应为 2: got %d", dashboard.PendingCount)
	}
}

func TestGetPersonaDashboard_WithClasses(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_dash_classes", Password: "p", Role: "teacher", Nickname: "教师DashClasses"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_dash_classes", Password: "p", Role: "student", Nickname: "学生DashClasses"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身DashClasses", School: "学校DashClasses"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身DashClasses"})

	// 创建2个班级
	class1ID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "班级A", Description: "描述A"})
	class2ID, _ := classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "班级B", Description: "描述B"})

	// 班级1添加成员
	_, _ = classRepo.AddMember(class1ID, studentPersonaID)

	// 停用班级2
	_, _ = classRepo.ToggleClass(class2ID, 0)

	dashboard, err := personaRepo.GetPersonaDashboard(teacherPersonaID)
	if err != nil {
		t.Fatalf("获取仪表盘失败: %v", err)
	}

	if len(dashboard.Classes) != 2 {
		t.Fatalf("应有2个班级: got %d", len(dashboard.Classes))
	}

	// 验证班级信息
	for _, c := range dashboard.Classes {
		if c.ID == class1ID {
			if c.MemberCount != 1 {
				t.Errorf("班级A成员数应为 1: got %d", c.MemberCount)
			}
			if !c.IsActive {
				t.Error("班级A应为活跃状态")
			}
		}
		if c.ID == class2ID {
			if c.MemberCount != 0 {
				t.Errorf("班级B成员数应为 0: got %d", c.MemberCount)
			}
			if c.IsActive {
				t.Error("班级B应为停用状态")
			}
		}
	}
}

func TestGetPersonaDashboard_Stats(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)
	classRepo := NewClassRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	teacherUserID, _ := userRepo.Create(&User{Username: "teacher_dash_stats", Password: "p", Role: "teacher", Nickname: "教师DashStats"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_dash_stats", Password: "p", Role: "student", Nickname: "学生DashStats"})

	teacherPersonaID, _ := personaRepo.Create(&Persona{UserID: teacherUserID, Role: "teacher", Nickname: "教师分身DashStats", School: "学校DashStats"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身DashStats"})

	// 创建 approved 关系
	_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")

	// 创建班级
	_, _ = classRepo.Create(&Class{PersonaID: teacherPersonaID, Name: "统计班级"})

	// 插入知识条目（迭代8迁移到 knowledge_items 表）
	_, _ = db.DB.Exec(
		`INSERT INTO knowledge_items (teacher_id, persona_id, title, content, item_type, status) VALUES (?, ?, ?, ?, 'text', 'active')`,
		teacherUserID, teacherPersonaID, "测试文档", "内容",
	)

	dashboard, err := personaRepo.GetPersonaDashboard(teacherPersonaID)
	if err != nil {
		t.Fatalf("获取仪表盘失败: %v", err)
	}

	// 验证统计信息
	if dashboard.Stats == nil {
		t.Fatal("Stats 不应为 nil")
	}

	totalStudents, ok := dashboard.Stats["total_students"]
	if !ok {
		t.Fatal("Stats 应包含 total_students")
	}
	if totalStudents != 1 {
		t.Errorf("total_students 应为 1: got %v", totalStudents)
	}

	totalClasses, ok := dashboard.Stats["total_classes"]
	if !ok {
		t.Fatal("Stats 应包含 total_classes")
	}
	if totalClasses != 1 {
		t.Errorf("total_classes 应为 1: got %v", totalClasses)
	}

	totalDocuments, ok := dashboard.Stats["total_documents"]
	if !ok {
		t.Fatal("Stats 应包含 total_documents")
	}
	if totalDocuments != 1 {
		t.Errorf("total_documents 应为 1: got %v", totalDocuments)
	}
}

func TestGetPersonaDashboard_NonExistentPersona(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)

	_, err := personaRepo.GetPersonaDashboard(99999)
	if err == nil {
		t.Fatal("不存在的分身应返回错误")
	}
}

func TestGetPersonaDashboard_EmptyDashboard(t *testing.T) {
	db := setupTestDB(t)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	userID, _ := userRepo.Create(&User{Username: "teacher_dash_empty", Password: "p", Role: "teacher", Nickname: "教师Empty"})
	personaID, _ := personaRepo.Create(&Persona{UserID: userID, Role: "teacher", Nickname: "教师分身Empty", School: "学校Empty"})

	dashboard, err := personaRepo.GetPersonaDashboard(personaID)
	if err != nil {
		t.Fatalf("获取仪表盘失败: %v", err)
	}

	// 验证空仪表盘
	if dashboard.PendingCount != 0 {
		t.Errorf("PendingCount 应为 0: got %d", dashboard.PendingCount)
	}
	if len(dashboard.Classes) != 0 {
		t.Errorf("Classes 应为空: got %d", len(dashboard.Classes))
	}
	if dashboard.LatestShare != nil {
		t.Error("LatestShare 应为 nil")
	}

	// Stats 应有值但都为 0
	if dashboard.Stats == nil {
		t.Fatal("Stats 不应为 nil")
	}
	if dashboard.Stats["total_students"] != 0 {
		t.Errorf("total_students 应为 0: got %v", dashboard.Stats["total_students"])
	}
}
