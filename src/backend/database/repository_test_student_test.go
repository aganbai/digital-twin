package database

import (
	"testing"
)

// ======================== V2.0 迭代11 M4 自测学生测试 ========================

func TestUserRepository_CreateTestStudent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db.DB)

	// 先创建一个教师用户
	teacher := &User{
		Username: "test_teacher_m4",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "测试教师",
	}
	teacherID, err := repo.Create(teacher)
	if err != nil {
		t.Fatalf("创建教师失败: %v", err)
	}

	// 创建自测学生
	testStudent := &User{
		Username:      "teacher_" + string(rune(teacherID)) + "_test",
		Password:      "hashed_test_password",
		Role:          "student",
		Nickname:      "测试学生",
		IsTestStudent: true,
		TestTeacherID: teacherID,
	}

	userID, err := repo.CreateTestStudent(testStudent)
	if err != nil {
		t.Fatalf("创建自测学生失败: %v", err)
	}
	if userID <= 0 {
		t.Error("返回的用户ID应该大于0")
	}

	// 查询自测学生
	found, err := repo.FindByTestTeacherID(teacherID)
	if err != nil {
		t.Fatalf("查询自测学生失败: %v", err)
	}
	if found == nil {
		t.Fatal("应该找到自测学生")
	}
	if found.TestTeacherID != teacherID {
		t.Errorf("test_teacher_id 不匹配: 期望 %d, 实际 %d", teacherID, found.TestTeacherID)
	}
	if !found.IsTestStudent {
		t.Error("is_test_student 应该为 true")
	}
	if found.Nickname != "测试学生" {
		t.Errorf("昵称不匹配: 期望 '测试学生', 实际 '%s'", found.Nickname)
	}
}

func TestUserRepository_FindByTestTeacherID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db.DB)

	// 查询不存在的教师ID
	found, err := repo.FindByTestTeacherID(99999)
	if err != nil {
		t.Fatalf("查询不应该报错: %v", err)
	}
	if found != nil {
		t.Error("不应该找到自测学生")
	}
}

func TestPersonaRepository_SearchStudentPersonas_ExcludeTestStudent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	// 创建普通学生用户
	normalStudent := &User{
		Username: "normal_student_search",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "普通学生",
	}
	normalStudentID, _ := userRepo.Create(normalStudent)

	// 创建普通学生分身
	normalPersona := &Persona{
		UserID:   normalStudentID,
		Role:     "student",
		Nickname: "张三",
		School:   "测试学校",
	}
	normalPersonaID, _ := personaRepo.Create(normalPersona)

	// 创建教师用户
	teacher := &User{
		Username: "teacher_for_search_test",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "搜索测试教师",
	}
	teacherID, _ := userRepo.Create(teacher)

	// 创建自测学生用户
	testStudent := &User{
		Username:      "teacher_search_test",
		Password:      "hashed_password",
		Role:          "student",
		Nickname:      "测试学生",
		IsTestStudent: true,
		TestTeacherID: teacherID,
	}
	testStudentID, _ := userRepo.CreateTestStudent(testStudent)

	// 创建自测学生分身
	testPersona := &Persona{
		UserID:   testStudentID,
		Role:     "student",
		Nickname: "测试学生张三", // 使用类似关键词
		School:   "",
	}
	personaRepo.Create(testPersona)

	// 搜索学生分身（应该排除自测学生）
	results, total, err := personaRepo.SearchStudentPersonas("张", 0, 10)
	if err != nil {
		t.Fatalf("搜索学生分身失败: %v", err)
	}

	// 验证结果中不包含自测学生
	for _, p := range results {
		if p.ID == testPersona.ID {
			t.Error("搜索结果不应该包含自测学生")
		}
	}

	// 应该只找到普通学生
	if total != 1 {
		t.Errorf("搜索结果数量不匹配: 期望 1, 实际 %d", total)
	}

	if len(results) > 0 && results[0].ID != normalPersonaID {
		t.Errorf("搜索结果应该是普通学生分身")
	}

	t.Logf("搜索结果: total=%d, results=%v", total, results)
}

func TestPersonaRepository_GetStudentPersonaByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	// 创建学生用户
	student := &User{
		Username: "student_get_test",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "获取测试学生",
	}
	studentID, _ := userRepo.Create(student)

	// 创建学生分身
	persona := &Persona{
		UserID:   studentID,
		Role:     "student",
		Nickname: "学生分身",
		School:   "测试学校",
	}
	personaID, _ := personaRepo.Create(persona)

	// 查询学生分身
	found, err := personaRepo.GetStudentPersonaByUserID(studentID)
	if err != nil {
		t.Fatalf("查询学生分身失败: %v", err)
	}
	if found == nil {
		t.Fatal("应该找到学生分身")
	}
	if found.ID != personaID {
		t.Errorf("分身ID不匹配: 期望 %d, 实际 %d", personaID, found.ID)
	}
	if found.Role != "student" {
		t.Errorf("角色应该是 student, 实际 %s", found.Role)
	}
}

func TestClassRepository_ListClassesByStudentPersonaID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	classRepo := NewClassRepository(db.DB)

	// 创建教师
	teacher := &User{
		Username: "teacher_class_list",
		Password: "hashed_password",
		Role:     "teacher",
		Nickname: "班级列表教师",
	}
	teacherID, _ := userRepo.Create(teacher)

	// 创建教师分身
	teacherPersona := &Persona{
		UserID:   teacherID,
		Role:     "teacher",
		Nickname: "班级列表教师分身",
		School:   "测试学校",
	}
	teacherPersonaID, _ := personaRepo.Create(teacherPersona)

	// 创建学生
	student := &User{
		Username: "student_class_list",
		Password: "hashed_password",
		Role:     "student",
		Nickname: "班级列表学生",
	}
	studentID, _ := userRepo.Create(student)

	// 创建学生分身
	studentPersona := &Persona{
		UserID:   studentID,
		Role:     "student",
		Nickname: "班级列表学生分身",
	}
	studentPersonaID, _ := personaRepo.Create(studentPersona)

	// 创建班级
	class1 := &Class{
		PersonaID:   teacherPersonaID,
		Name:        "测试班级1",
		Description: "描述1",
	}
	classID1, _ := classRepo.Create(class1)

	class2 := &Class{
		PersonaID:   teacherPersonaID,
		Name:        "测试班级2",
		Description: "描述2",
	}
	classID2, _ := classRepo.Create(class2)

	// 将学生加入班级
	classRepo.AddMember(classID1, studentPersonaID)
	classRepo.AddMember(classID2, studentPersonaID)

	// 查询学生所在班级
	classes, err := classRepo.ListClassesByStudentPersonaID(studentPersonaID)
	if err != nil {
		t.Fatalf("查询学生班级失败: %v", err)
	}
	if len(classes) != 2 {
		t.Errorf("班级数量不匹配: 期望 2, 实际 %d", len(classes))
	}

	// 验证班级名称
	classNames := make(map[string]bool)
	for _, c := range classes {
		classNames[c.Name] = true
	}
	if !classNames["测试班级1"] || !classNames["测试班级2"] {
		t.Error("应该包含两个测试班级")
	}

	t.Logf("学生所在班级: %v", classes)
}
