package database

import (
	"encoding/json"
	"testing"
)

// ==================== 教材配置 Repository 测试 (BE-IT13-002) ====================

// TestCurriculumConfigRepository_CreateWithTx_Success 测试通过persona_id创建教材配置
func TestCurriculumConfigRepository_CreateWithTx_Success(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户
	userID, err := userRepo.Create(&User{
		Username: "teacher_curriculum",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师教材配置",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建教师分身
	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身教材",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 开始事务
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 创建教材配置
	textbookVersions, _ := json.Marshal([]string{"人教版", "北师大版"})
	subjects, _ := json.Marshal([]string{"数学", "语文"})
	customTextbooks, _ := json.Marshal([]string{"《小学奥数》"})

	config := &TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_lower",
		Grade:            "三年级",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		Region:           string(customTextbooks),
		CurrentProgress:  "第三单元 乘法初步",
	}

	configID, err := curriculumRepo.CreateWithTx(tx, config)
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}
	if configID <= 0 {
		t.Errorf("配置ID应该大于0: got %d", configID)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证配置已创建
	createdConfig, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if createdConfig == nil {
		t.Fatal("教材配置应该存在")
	}
	if createdConfig.PersonaID != personaID {
		t.Errorf("PersonaID不匹配: got %d, want %d", createdConfig.PersonaID, personaID)
	}
	if createdConfig.GradeLevel != "primary_lower" {
		t.Errorf("GradeLevel错误: got %s, want %s", createdConfig.GradeLevel, "primary_lower")
	}
	if createdConfig.Grade != "三年级" {
		t.Errorf("Grade错误: got %s, want %s", createdConfig.Grade, "三年级")
	}
	if createdConfig.IsActive != 1 {
		t.Errorf("IsActive应该为1: got %d", createdConfig.IsActive)
	}
}

// TestCurriculumConfigRepository_GetActiveByPersonaID 测试通过persona_id查询活跃教材配置
func TestCurriculumConfigRepository_GetActiveByPersonaID(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_get_config",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师查询配置",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身查询",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 场景1: 无配置时返回nil
	config, err := curriculumRepo.GetActiveByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if config != nil {
		t.Error("无配置时应该返回nil")
	}

	// 创建配置
	textbookVersions, _ := json.Marshal([]string{"苏教版"})
	subjects, _ := json.Marshal([]string{"英语"})

	_, err = curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "junior",
		Grade:            "七年级",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		CurrentProgress:  "第一单元",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 场景2: 查询到活跃配置
	config, err = curriculumRepo.GetActiveByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if config == nil {
		t.Fatal("应该查询到教材配置")
	}
	if config.PersonaID != personaID {
		t.Errorf("PersonaID不匹配: got %d, want %d", config.PersonaID, personaID)
	}
	if config.GradeLevel != "junior" {
		t.Errorf("GradeLevel错误: got %s, want %s", config.GradeLevel, "junior")
	}
	if config.CurrentProgress != "第一单元" {
		t.Errorf("CurrentProgress错误: got %s, want %s", config.CurrentProgress, "第一单元")
	}
}

// TestCurriculumConfigRepository_UpsertByPersonaID_Create 测试Upsert创建新配置
func TestCurriculumConfigRepository_UpsertByPersonaID_Create(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_upsert_create",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师Upsert",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身Upsert",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 开启事务
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 使用Upsert创建配置（不存在时创建）
	textbookVersions, _ := json.Marshal([]string{"部编版"})
	subjects, _ := json.Marshal([]string{"语文"})

	config := &TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "senior",
		Grade:            "高一",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		CurrentProgress:  "第二单元 古文阅读",
	}

	configID, err := curriculumRepo.UpsertByPersonaID(tx, config)
	if err != nil {
		t.Fatalf("Upsert教材配置失败: %v", err)
	}
	if configID <= 0 {
		t.Errorf("配置ID应该大于0: got %d", configID)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证配置已创建
	created, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if created.GradeLevel != "senior" {
		t.Errorf("GradeLevel错误: got %s, want %s", created.GradeLevel, "senior")
	}
}

// TestCurriculumConfigRepository_UpsertByPersonaID_Update 测试Upsert更新现有配置
func TestCurriculumConfigRepository_UpsertByPersonaID_Update(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_upsert_update",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师Upsert更新",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身Upsert更新",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 先创建配置
	textbookVersions, _ := json.Marshal([]string{"人教版"})
	subjects, _ := json.Marshal([]string{"数学"})

	configID, err := curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_lower",
		Grade:            "二年级",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		CurrentProgress:  "第一单元",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 开启事务
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 使用Upsert更新配置
	newTextbookVersions, _ := json.Marshal([]string{"人教版", "苏教版"})
	newSubjects, _ := json.Marshal([]string{"数学", "语文", "英语"})

	config := &TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_upper",
		Grade:            "四年级",
		TextbookVersions: string(newTextbookVersions),
		Subjects:         string(newSubjects),
		CurrentProgress:  "第三单元 进阶内容",
	}

	updatedID, err := curriculumRepo.UpsertByPersonaID(tx, config)
	if err != nil {
		t.Fatalf("Upsert更新教材配置失败: %v", err)
	}
	if updatedID != configID {
		t.Errorf("Upsert应该返回原有配置ID: got %d, want %d", updatedID, configID)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证配置已更新
	updated, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if updated.GradeLevel != "primary_upper" {
		t.Errorf("GradeLevel应该更新: got %s, want %s", updated.GradeLevel, "primary_upper")
	}
	if updated.Grade != "四年级" {
		t.Errorf("Grade应该更新: got %s, want %s", updated.Grade, "四年级")
	}
	if updated.CurrentProgress != "第三单元 进阶内容" {
		t.Errorf("CurrentProgress应该更新: got %s, want %s", updated.CurrentProgress, "第三单元 进阶内容")
	}

	// 验证只有一个活跃配置
	allConfigs, err := curriculumRepo.ListByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询配置列表失败: %v", err)
	}
	if len(allConfigs) != 1 {
		t.Errorf("应该只有一个配置: got %d", len(allConfigs))
	}
}

// TestCurriculumConfigRepository_ListByPersonaID 测试获取分身的所有教材配置
func TestCurriculumConfigRepository_ListByPersonaID(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_list_configs",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师列表配置",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身列表",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 初始状态应该无配置
	configs, err := curriculumRepo.ListByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询配置列表失败: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("初始应该无配置: got %d", len(configs))
	}

	// 创建一个配置
	_, err = curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_lower",
		Grade:            "一年级",
		TextbookVersions: "[]",
		Subjects:         "[]",
		CurrentProgress:  "第一单元",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 查询列表
	configs, err = curriculumRepo.ListByPersonaID(personaID)
	if err != nil {
		t.Fatalf("查询配置列表失败: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("应该有一个配置: got %d", len(configs))
	}

	// 验证配置内容
	if configs[0].GradeLevel != "primary_lower" {
		t.Errorf("GradeLevel错误: got %s", configs[0].GradeLevel)
	}
}

// TestCurriculumConfigRepository_UpdateWithTx 测试事务中更新配置
func TestCurriculumConfigRepository_UpdateWithTx(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_update_tx",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师更新事务",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身更新",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 先创建配置
	configID, err := curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "junior",
		Grade:            "七年级",
		TextbookVersions: "[]",
		Subjects:         "[]",
		CurrentProgress:  "",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 开启事务
	tx, err := db.DB.Begin()
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 更新配置
	textbookVersions, _ := json.Marshal([]string{"人教版"})
	subjects, _ := json.Marshal([]string{"数学"})

	config := &TeacherCurriculumConfig{
		ID:               configID,
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "junior",
		Grade:            "八年级",
		TextbookVersions: string(textbookVersions),
		Subjects:         string(subjects),
		Region:           "[]",
		CurrentProgress:  "第二单元",
	}

	err = curriculumRepo.UpdateWithTx(tx, config)
	if err != nil {
		t.Fatalf("更新教材配置失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证更新
	updated, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}
	if updated.Grade != "八年级" {
		t.Errorf("Grade应该更新: got %s, want %s", updated.Grade, "八年级")
	}
	if updated.CurrentProgress != "第二单元" {
		t.Errorf("CurrentProgress应该更新: got %s, want %s", updated.CurrentProgress, "第二单元")
	}
}

// TestCurriculumConfigRepository_UpdateWithTx_Unauthorized 测试无权更新配置
func TestCurriculumConfigRepository_UpdateWithTx_Unauthorized(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户A和分身
	userA, _ := userRepo.Create(&User{
		Username: "teacher_A",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师A",
	})
	personaA, _ := personaRepo.Create(&Persona{
		UserID:   userA,
		Role:     "teacher",
		Nickname: "教师分身A",
		School:   "测试学校",
	})

	// 创建教师用户B
	userB, _ := userRepo.Create(&User{
		Username: "teacher_B",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师B",
	})

	// 用户A创建配置
	configID, _ := curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userA,
		PersonaID:        personaA,
		GradeLevel:       "primary_lower",
		Grade:            "三年级",
		TextbookVersions: "[]",
		Subjects:         "[]",
	})

	// 用户B尝试更新（使用错误的teacher_id）
	config := &TeacherCurriculumConfig{
		ID:               configID,
		TeacherID:        userB, // 错误的所有者
		PersonaID:        personaA,
		GradeLevel:       "primary_lower",
		Grade:            "四年级",
		TextbookVersions: "[]",
		Subjects:         "[]",
		Region:           "[]",
	}

	err := curriculumRepo.UpdateWithTx(nil, config)
	if err == nil {
		t.Error("无权修改应该返回错误")
	}
}

// TestCurriculumConfigRepository_Delete 测试删除教材配置
func TestCurriculumConfigRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, err := userRepo.Create(&User{
		Username: "teacher_delete",
		Password: "password",
		Role:     "teacher",
		Nickname: "教师删除",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	personaID, err := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师分身删除",
		School:   "测试学校",
	})
	if err != nil {
		t.Fatalf("创建分身失败: %v", err)
	}

	// 创建配置
	configID, err := curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_lower",
		Grade:            "一年级",
		TextbookVersions: "[]",
		Subjects:         "[]",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 验证配置存在
	_, err = curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("配置应该存在: %v", err)
	}

	// 删除配置
	err = curriculumRepo.Delete(configID)
	if err != nil {
		t.Fatalf("删除教材配置失败: %v", err)
	}

	// 验证配置已删除
	deleted, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if deleted != nil {
		t.Error("配置应该已删除")
	}
}

// TestCurriculumConfigRepository_AllGradeLevels 测试所有学段类型的存储
func TestCurriculumConfigRepository_AllGradeLevels(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	gradeLevels := []struct {
		level string
		grade string
	}{
		{"preschool", "幼儿园大班"},
		{"primary_lower", "一年级"},
		{"primary_upper", "五年级"},
		{"junior", "七年级"},
		{"senior", "高一"},
		{"university", "大一"},
		{"adult_life", ""},
		{"adult_professional", ""},
	}

	for i, tt := range gradeLevels {
		// 创建教师用户和分身（使用唯一的用户名和昵称）
		nickname := "教师ALL" + string(rune('A'+i))
		userID, err := userRepo.Create(&User{
			Username: "teacher_grade_" + tt.level,
			Password: "password",
			Role:     "teacher",
			Nickname: nickname,
		})
		if err != nil {
			t.Fatalf("创建用户失败: %v", err)
		}
		personaID, err := personaRepo.Create(&Persona{
			UserID:   userID,
			Role:     "teacher",
			Nickname: "分身" + string(rune('A'+i)),
			School:   "测试学校" + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("创建分身失败: %v", err)
		}

		// 创建配置
		config := &TeacherCurriculumConfig{
			TeacherID:        userID,
			PersonaID:        personaID,
			GradeLevel:       tt.level,
			Grade:            tt.grade,
			TextbookVersions: "[]",
			Subjects:         "[]",
		}

		configID, err := curriculumRepo.Create(config)
		if err != nil {
			t.Errorf("创建%s学段配置失败: %v", tt.level, err)
			continue
		}

		// 查询验证
		created, err := curriculumRepo.GetByID(configID)
		if err != nil {
			t.Errorf("查询%s学段配置失败: %v", tt.level, err)
			continue
		}
		if created.GradeLevel != tt.level {
			t.Errorf("%s学段不匹配: got %s", tt.level, created.GradeLevel)
		}
	}
}

// TestCurriculumConfigRepository_JSONFields 测试JSON字段的序列化和反序列化
func TestCurriculumConfigRepository_JSONFields(t *testing.T) {
	db := setupTestDB(t)
	curriculumRepo := NewCurriculumConfigRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建教师用户和分身
	userID, _ := userRepo.Create(&User{
		Username: "teacher_json",
		Password: "password",
		Role:     "teacher",
	})
	personaID, _ := personaRepo.Create(&Persona{
		UserID:   userID,
		Role:     "teacher",
		Nickname: "教师JSON",
		School:   "测试学校",
	})

	// 准备复杂JSON数据
	textbookVersions := []string{"人教版", "北师大版", "苏教版"}
	subjects := []string{"数学", "语文", "英语", "物理", "化学"}
	customTextbooks := []string{"《小学奥数》", "《语文培优》", "《英语听力训练》"}

	textbookVersionsJSON, _ := json.Marshal(textbookVersions)
	subjectsJSON, _ := json.Marshal(subjects)
	customTextbooksJSON, _ := json.Marshal(customTextbooks)

	configID, err := curriculumRepo.Create(&TeacherCurriculumConfig{
		TeacherID:        userID,
		PersonaID:        personaID,
		GradeLevel:       "primary_upper",
		Grade:            "五年级",
		TextbookVersions: string(textbookVersionsJSON),
		Subjects:         string(subjectsJSON),
		Region:           string(customTextbooksJSON),
		CurrentProgress:  "第三单元 综合练习",
	})
	if err != nil {
		t.Fatalf("创建教材配置失败: %v", err)
	}

	// 查询并验证
	created, err := curriculumRepo.GetByID(configID)
	if err != nil {
		t.Fatalf("查询教材配置失败: %v", err)
	}

	// 反序列化验证
	var resultVersions, resultSubjects, resultCustom []string
	json.Unmarshal([]byte(created.TextbookVersions), &resultVersions)
	json.Unmarshal([]byte(created.Subjects), &resultSubjects)
	json.Unmarshal([]byte(created.Region), &resultCustom)

	if len(resultVersions) != 3 {
		t.Errorf("TextbookVersions数量错误: got %d, want 3", len(resultVersions))
	}
	if len(resultSubjects) != 5 {
		t.Errorf("Subjects数量错误: got %d, want 5", len(resultSubjects))
	}
	if len(resultCustom) != 3 {
		t.Errorf("CustomTextbooks数量错误: got %d, want 3", len(resultCustom))
	}

	// 验证具体内容
	found := false
	for _, v := range resultSubjects {
		if v == "数学" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Subjects中应该包含数学")
	}
}
