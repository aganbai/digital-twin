package database

import (
	"testing"
)

// ==================== ListTeachersForStudent 测试 ====================

func TestListTeachersForStudent_BasicQuery(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)

	// 创建2个教师和1个学生
	teacher1UserID, _ := userRepo.Create(&User{Username: "teacher_list1", Password: "p", Role: "teacher", Nickname: "教师1", School: "学校1"})
	teacher2UserID, _ := userRepo.Create(&User{Username: "teacher_list2", Password: "p", Role: "teacher", Nickname: "教师2", School: "学校2"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_list1", Password: "p", Role: "student", Nickname: "学生1"})

	teacher1PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher1UserID, Role: "teacher", Nickname: "教师分身1", School: "学校1"})
	teacher2PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher2UserID, Role: "teacher", Nickname: "教师分身2", School: "学校2"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身1"})

	// 创建 approved 关系
	_, _ = relationRepo.CreateWithPersonas(teacher1UserID, studentUserID, teacher1PersonaID, studentPersonaID, "approved", "teacher")
	_, _ = relationRepo.CreateWithPersonas(teacher2UserID, studentUserID, teacher2PersonaID, studentPersonaID, "approved", "teacher")

	teachers, total, err := userRepo.ListTeachersForStudent(studentPersonaID, 0, 20)
	if err != nil {
		t.Fatalf("查询教师列表失败: %v", err)
	}

	if total != 2 {
		t.Errorf("total 应为 2: got %d", total)
	}
	if len(teachers) != 2 {
		t.Errorf("应返回 2 个教师: got %d", len(teachers))
	}
}

func TestListTeachersForStudent_OnlyApproved(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)

	teacher1UserID, _ := userRepo.Create(&User{Username: "teacher_only_approved1", Password: "p", Role: "teacher", Nickname: "教师OA1", School: "学校OA1"})
	teacher2UserID, _ := userRepo.Create(&User{Username: "teacher_only_approved2", Password: "p", Role: "teacher", Nickname: "教师OA2", School: "学校OA2"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_only_approved", Password: "p", Role: "student", Nickname: "学生OA"})

	teacher1PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher1UserID, Role: "teacher", Nickname: "教师分身OA1", School: "学校OA1"})
	teacher2PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher2UserID, Role: "teacher", Nickname: "教师分身OA2", School: "学校OA2"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身OA"})

	// 一个 approved，一个 pending
	_, _ = relationRepo.CreateWithPersonas(teacher1UserID, studentUserID, teacher1PersonaID, studentPersonaID, "approved", "teacher")
	_, _ = relationRepo.CreateWithPersonas(teacher2UserID, studentUserID, teacher2PersonaID, studentPersonaID, "pending", "student")

	teachers, total, err := userRepo.ListTeachersForStudent(studentPersonaID, 0, 20)
	if err != nil {
		t.Fatalf("查询教师列表失败: %v", err)
	}

	if total != 1 {
		t.Errorf("total 应为 1（仅 approved）: got %d", total)
	}
	if len(teachers) != 1 {
		t.Errorf("应返回 1 个教师: got %d", len(teachers))
	}
}

func TestListTeachersForStudent_ExcludeInactive(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)

	teacher1UserID, _ := userRepo.Create(&User{Username: "teacher_excl_inact1", Password: "p", Role: "teacher", Nickname: "教师EI1", School: "学校EI1"})
	teacher2UserID, _ := userRepo.Create(&User{Username: "teacher_excl_inact2", Password: "p", Role: "teacher", Nickname: "教师EI2", School: "学校EI2"})
	teacher3UserID, _ := userRepo.Create(&User{Username: "teacher_excl_inact3", Password: "p", Role: "teacher", Nickname: "教师EI3", School: "学校EI3"})
	studentUserID, _ := userRepo.Create(&User{Username: "student_excl_inact", Password: "p", Role: "student", Nickname: "学生EI"})

	teacher1PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher1UserID, Role: "teacher", Nickname: "教师分身EI1", School: "学校EI1"})
	teacher2PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher2UserID, Role: "teacher", Nickname: "教师分身EI2", School: "学校EI2"})
	teacher3PersonaID, _ := personaRepo.Create(&Persona{UserID: teacher3UserID, Role: "teacher", Nickname: "教师分身EI3", School: "学校EI3"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身EI"})

	relID1, _ := relationRepo.CreateWithPersonas(teacher1UserID, studentUserID, teacher1PersonaID, studentPersonaID, "approved", "teacher")
	_, _ = relationRepo.CreateWithPersonas(teacher2UserID, studentUserID, teacher2PersonaID, studentPersonaID, "approved", "teacher")
	_, _ = relationRepo.CreateWithPersonas(teacher3UserID, studentUserID, teacher3PersonaID, studentPersonaID, "approved", "teacher")

	// 停用第一个关系的 is_active
	_, _ = relationRepo.ToggleRelation(relID1, 0)

	// 停用第三个教师分身的 is_active（模拟教师分身被停用）
	_ = personaRepo.SetActive(teacher3PersonaID, 0)

	teachers, total, err := userRepo.ListTeachersForStudent(studentPersonaID, 0, 20)
	if err != nil {
		t.Fatalf("查询教师列表失败: %v", err)
	}

	// 教师1：关系 is_active=0 → 排除
	// 教师2：关系 approved + is_active=1，分身 is_active=1 → 保留
	// 教师3：关系 approved + is_active=1，但分身 is_active=0 → 排除
	if total != 1 {
		t.Errorf("total 应为 1（排除停用关系和停用分身）: got %d", total)
	}
	if len(teachers) != 1 {
		t.Errorf("应返回 1 个教师: got %d", len(teachers))
	}
}

func TestListTeachersForStudent_EmptyResult(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	studentUserID, _ := userRepo.Create(&User{Username: "student_empty_list", Password: "p", Role: "student", Nickname: "学生Empty"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Empty"})

	teachers, total, err := userRepo.ListTeachersForStudent(studentPersonaID, 0, 20)
	if err != nil {
		t.Fatalf("查询教师列表失败: %v", err)
	}

	if total != 0 {
		t.Errorf("total 应为 0: got %d", total)
	}
	if len(teachers) != 0 {
		t.Errorf("应返回空列表: got %d", len(teachers))
	}
}

func TestListTeachersForStudent_NonExistentPersona(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 使用不存在的分身ID
	teachers, total, err := userRepo.ListTeachersForStudent(99999, 0, 20)
	if err != nil {
		t.Fatalf("不应返回错误: %v", err)
	}

	if total != 0 {
		t.Errorf("total 应为 0: got %d", total)
	}
	if len(teachers) != 0 {
		t.Errorf("应返回空列表: got %d", len(teachers))
	}
}

func TestListTeachersForStudent_Pagination(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	relationRepo := NewRelationRepository(db.DB)

	studentUserID, _ := userRepo.Create(&User{Username: "student_page", Password: "p", Role: "student", Nickname: "学生Page"})
	studentPersonaID, _ := personaRepo.Create(&Persona{UserID: studentUserID, Role: "student", Nickname: "学生分身Page"})

	// 创建3个教师并建立关系
	for i := 1; i <= 3; i++ {
		teacherUserID, _ := userRepo.Create(&User{
			Username: "teacher_page_" + string(rune('0'+i)),
			Password: "p",
			Role:     "teacher",
			Nickname: "教师Page" + string(rune('0'+i)),
			School:   "学校Page" + string(rune('0'+i)),
		})
		teacherPersonaID, _ := personaRepo.Create(&Persona{
			UserID:   teacherUserID,
			Role:     "teacher",
			Nickname: "教师分身Page" + string(rune('0'+i)),
			School:   "学校Page" + string(rune('0'+i)),
		})
		_, _ = relationRepo.CreateWithPersonas(teacherUserID, studentUserID, teacherPersonaID, studentPersonaID, "approved", "teacher")
	}

	// 第一页，每页2条
	teachers, total, err := userRepo.ListTeachersForStudent(studentPersonaID, 0, 2)
	if err != nil {
		t.Fatalf("查询第一页失败: %v", err)
	}
	if total != 3 {
		t.Errorf("total 应为 3: got %d", total)
	}
	if len(teachers) != 2 {
		t.Errorf("第一页应返回 2 条: got %d", len(teachers))
	}

	// 第二页
	teachers2, _, err := userRepo.ListTeachersForStudent(studentPersonaID, 2, 2)
	if err != nil {
		t.Fatalf("查询第二页失败: %v", err)
	}
	if len(teachers2) != 1 {
		t.Errorf("第二页应返回 1 条: got %d", len(teachers2))
	}
}
