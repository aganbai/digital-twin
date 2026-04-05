package database

import (
	"testing"
	"time"
)

// ==================== CourseNotificationRepository 测试 ====================

func TestCourseNotification_CreateAndGetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseNotificationRepository(db.DB)

	notification := &CourseNotification{
		CourseItemID: 1,
		ClassID:      100,
		TeacherID:    10,
		PersonaID:    20,
		PushType:     "in_app",
		Status:       "pending",
	}

	id, err := repo.Create(notification)
	if err != nil {
		t.Fatalf("创建课程推送通知失败: %v", err)
	}
	if id == 0 {
		t.Error("返回的ID不应为0")
	}

	// 查询验证
	got, err := repo.GetByID(id)
	if err != nil {
		t.Fatalf("查询课程推送通知失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回通知记录")
	}
	if got.CourseItemID != 1 {
		t.Errorf("CourseItemID 应为 1: got %d", got.CourseItemID)
	}
	if got.PushType != "in_app" {
		t.Errorf("PushType 应为 'in_app': got %s", got.PushType)
	}
}

func TestCourseNotification_GetByClassID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseNotificationRepository(db.DB)

	// 创建多条记录
	for i := 1; i <= 3; i++ {
		_, _ = repo.Create(&CourseNotification{
			CourseItemID: int64(i),
			ClassID:      100,
			TeacherID:    10,
			PersonaID:    20,
			PushType:     "in_app",
			Status:       "pending",
		})
	}
	// 创建不同班级的记录
	_, _ = repo.Create(&CourseNotification{
		CourseItemID: 99,
		ClassID:      200,
		TeacherID:    10,
		PersonaID:    20,
		PushType:     "wechat",
		Status:       "pending",
	})

	notifications, total, err := repo.GetByClassID(100, 0, 10)
	if err != nil {
		t.Fatalf("查询课程推送通知列表失败: %v", err)
	}
	if total != 3 {
		t.Errorf("total 应为 3: got %d", total)
	}
	if len(notifications) != 3 {
		t.Errorf("应返回 3 条记录: got %d", len(notifications))
	}
}

func TestCourseNotification_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseNotificationRepository(db.DB)

	id, _ := repo.Create(&CourseNotification{
		CourseItemID: 1,
		ClassID:      100,
		TeacherID:    10,
		PersonaID:    20,
		PushType:     "wechat",
		Status:       "pending",
	})

	err := repo.UpdateStatus(id, "sent")
	if err != nil {
		t.Fatalf("更新推送状态失败: %v", err)
	}

	got, _ := repo.GetByID(id)
	if got.Status != "sent" {
		t.Errorf("Status 应为 'sent': got %s", got.Status)
	}
}

func TestCourseNotification_GetPendingByPersona(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCourseNotificationRepository(db.DB)

	personaID := int64(20)

	// 创建待推送记录
	for i := 1; i <= 2; i++ {
		_, _ = repo.Create(&CourseNotification{
			CourseItemID: int64(i),
			ClassID:      100,
			TeacherID:    10,
			PersonaID:    personaID,
			PushType:     "in_app",
			Status:       "pending",
		})
	}
	// 创建已发送记录
	id, _ := repo.Create(&CourseNotification{
		CourseItemID: 3,
		ClassID:      100,
		TeacherID:    10,
		PersonaID:    personaID,
		PushType:     "in_app",
		Status:       "sent",
	})
	_ = repo.UpdateStatus(id, "sent")

	notifications, err := repo.GetPendingByPersona(personaID, 10)
	if err != nil {
		t.Fatalf("查询待推送通知失败: %v", err)
	}
	if len(notifications) != 2 {
		t.Errorf("应返回 2 条待推送记录: got %d", len(notifications))
	}
}

// ==================== WxSubscriptionRepository 测试 ====================

func TestWxSubscription_UpsertAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWxSubscriptionRepository(db.DB)

	userID := int64(1)
	templateID := "template_001"

	// 首次创建订阅
	err := repo.Upsert(userID, templateID, true)
	if err != nil {
		t.Fatalf("创建微信订阅失败: %v", err)
	}

	// 查询验证
	sub, err := repo.GetByUserAndTemplate(userID, templateID)
	if err != nil {
		t.Fatalf("查询微信订阅失败: %v", err)
	}
	if sub == nil {
		t.Fatal("应返回订阅记录")
	}
	if !sub.IsSubscribed {
		t.Error("IsSubscribed 应为 true")
	}

	// 更新订阅状态为取消
	err = repo.Upsert(userID, templateID, false)
	if err != nil {
		t.Fatalf("更新微信订阅失败: %v", err)
	}

	sub, _ = repo.GetByUserAndTemplate(userID, templateID)
	if sub.IsSubscribed {
		t.Error("IsSubscribed 应为 false")
	}
}

func TestWxSubscription_GetByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWxSubscriptionRepository(db.DB)

	userID := int64(1)

	// 创建多个模板订阅
	for i := 1; i <= 3; i++ {
		_ = repo.Upsert(userID, "template_"+string(rune('0'+i)), true)
	}

	subscriptions, err := repo.GetByUser(userID)
	if err != nil {
		t.Fatalf("查询用户订阅列表失败: %v", err)
	}
	if len(subscriptions) != 3 {
		t.Errorf("应返回 3 条订阅记录: got %d", len(subscriptions))
	}
}

func TestWxSubscription_IsSubscribed(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWxSubscriptionRepository(db.DB)

	userID := int64(1)
	templateID := "template_001"

	// 未订阅时应返回 false
	isSubscribed, err := repo.IsSubscribed(userID, templateID)
	if err != nil {
		t.Fatalf("检查订阅状态失败: %v", err)
	}
	if isSubscribed {
		t.Error("未订阅时应返回 false")
	}

	// 订阅后检查
	_ = repo.Upsert(userID, templateID, true)
	isSubscribed, _ = repo.IsSubscribed(userID, templateID)
	if !isSubscribed {
		t.Error("订阅后应返回 true")
	}
}

// ==================== SessionTitleRepository 测试 ====================

func TestSessionTitle_CreateAndGetBySessionID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	title := &SessionTitle{
		SessionID:        "session_001",
		StudentPersonaID: 1,
		TeacherPersonaID: 2,
		Title:            "数学辅导课程",
	}

	id, err := repo.Create(title)
	if err != nil {
		t.Fatalf("创建会话标题失败: %v", err)
	}
	if id == 0 {
		t.Error("返回的ID不应为0")
	}

	// 查询验证
	got, err := repo.GetBySessionID("session_001")
	if err != nil {
		t.Fatalf("查询会话标题失败: %v", err)
	}
	if got == nil {
		t.Fatal("应返回标题记录")
	}
	if got.Title != "数学辅导课程" {
		t.Errorf("Title 应为 '数学辅导课程': got %s", got.Title)
	}
}

func TestSessionTitle_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	_, _ = repo.Create(&SessionTitle{
		SessionID:        "session_002",
		StudentPersonaID: 1,
		TeacherPersonaID: 2,
		Title:            "原标题",
	})

	err := repo.Update("session_002", "新标题")
	if err != nil {
		t.Fatalf("更新会话标题失败: %v", err)
	}

	got, _ := repo.GetBySessionID("session_002")
	if got.Title != "新标题" {
		t.Errorf("Title 应为 '新标题': got %s", got.Title)
	}
}

func TestSessionTitle_Upsert(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	// 首次创建
	err := repo.Upsert("session_003", 1, 2, "首次标题")
	if err != nil {
		t.Fatalf("Upsert 创建失败: %v", err)
	}

	got, _ := repo.GetBySessionID("session_003")
	if got.Title != "首次标题" {
		t.Errorf("Title 应为 '首次标题': got %s", got.Title)
	}

	// 更新
	err = repo.Upsert("session_003", 1, 2, "更新后的标题")
	if err != nil {
		t.Fatalf("Upsert 更新失败: %v", err)
	}

	got, _ = repo.GetBySessionID("session_003")
	if got.Title != "更新后的标题" {
		t.Errorf("Title 应为 '更新后的标题': got %s", got.Title)
	}
}

func TestSessionTitle_GetByStudentPersona(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	studentPersonaID := int64(1)

	// 创建多个会话标题
	for i := 1; i <= 3; i++ {
		_ = repo.Upsert("student_session_"+string(rune('0'+i)), studentPersonaID, 2, "标题"+string(rune('0'+i)))
	}
	// 创建其他学生的会话
	_ = repo.Upsert("other_session", 99, 2, "其他学生标题")

	titles, total, err := repo.GetByStudentPersona(studentPersonaID, 0, 10)
	if err != nil {
		t.Fatalf("查询学生会话标题失败: %v", err)
	}
	if total != 3 {
		t.Errorf("total 应为 3: got %d", total)
	}
	if len(titles) != 3 {
		t.Errorf("应返回 3 条记录: got %d", len(titles))
	}
}

func TestSessionTitle_GetByTeacherPersona(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	teacherPersonaID := int64(2)

	// 创建多个会话标题
	for i := 1; i <= 2; i++ {
		_ = repo.Upsert("teacher_session_"+string(rune('0'+i)), 1, teacherPersonaID, "标题"+string(rune('0'+i)))
	}

	titles, total, err := repo.GetByTeacherPersona(teacherPersonaID, 0, 10)
	if err != nil {
		t.Fatalf("查询教师会话标题失败: %v", err)
	}
	if total != 2 {
		t.Errorf("total 应为 2: got %d", total)
	}
	if len(titles) != 2 {
		t.Errorf("应返回 2 条记录: got %d", len(titles))
	}
}

func TestSessionTitle_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	_ = repo.Upsert("session_to_delete", 1, 2, "待删除标题")

	err := repo.Delete("session_to_delete")
	if err != nil {
		t.Fatalf("删除会话标题失败: %v", err)
	}

	got, _ := repo.GetBySessionID("session_to_delete")
	if got != nil {
		t.Error("删除后应返回 nil")
	}
}

// ==================== 画像隐私保护测试 ====================

func TestUserRepository_ProfileSnapshotPrivacy(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 创建用户
	userID, err := userRepo.Create(&User{
		Username: "privacy_test_user",
		Password: "p",
		Role:     "student",
		Nickname: "隐私测试用户",
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 设置用户画像快照
	testSnapshot := `{"age":10,"interests":["math","science"],"level":"advanced"}`
	err = userRepo.UpdateProfileSnapshot(userID, testSnapshot)
	if err != nil {
		t.Fatalf("更新用户画像失败: %v", err)
	}

	// 测试 GetByID - 应过滤 profile_snapshot
	user, err := userRepo.GetByID(userID)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if user.ProfileSnapshot != "" {
		t.Errorf("GetByID 应过滤 profile_snapshot，但返回了: %s", user.ProfileSnapshot)
	}

	// 测试 GetByUsername - 应过滤 profile_snapshot
	user2, err := userRepo.GetByUsername("privacy_test_user")
	if err != nil {
		t.Fatalf("根据用户名查询失败: %v", err)
	}
	if user2.ProfileSnapshot != "" {
		t.Errorf("GetByUsername 应过滤 profile_snapshot，但返回了: %s", user2.ProfileSnapshot)
	}

	// 测试 GetByOpenID - 应过滤 profile_snapshot（设置openid后测试）
	_, _ = db.DB.Exec(`UPDATE users SET openid = ? WHERE id = ?`, "test_openid_123", userID)
	user3, err := userRepo.GetByOpenID("test_openid_123")
	if err != nil {
		t.Fatalf("根据OpenID查询失败: %v", err)
	}
	if user3.ProfileSnapshot != "" {
		t.Errorf("GetByOpenID 应过滤 profile_snapshot，但返回了: %s", user3.ProfileSnapshot)
	}
}

func TestUserRepository_ProfileSnapshotPrivacy_EmptyUser(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewUserRepository(db.DB)

	// 创建用户（无画像）
	userID, _ := userRepo.Create(&User{
		Username: "no_profile_user",
		Password: "p",
		Role:     "student",
		Nickname: "无画像用户",
	})

	// 即使没有设置画像，返回时也应该是空字符串
	user, _ := userRepo.GetByID(userID)
	if user.ProfileSnapshot != "" {
		t.Errorf("未设置画像的用户，profile_snapshot 应为空字符串: got %s", user.ProfileSnapshot)
	}
}

func TestSessionTitle_UpdatedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSessionTitleRepository(db.DB)

	// 创建会话标题
	_ = repo.Upsert("session_time_test", 1, 2, "原始标题")
	time.Sleep(100 * time.Millisecond) // 等待一小段时间

	// 更新标题
	_ = repo.Upsert("session_time_test", 1, 2, "更新标题")

	got, _ := repo.GetBySessionID("session_time_test")
	if got.Title != "更新标题" {
		t.Errorf("Title 应为 '更新标题': got %s", got.Title)
	}
	// 验证 updated_at 大于 created_at
	if !got.UpdatedAt.After(got.CreatedAt) {
		t.Error("UpdatedAt 应大于 CreatedAt")
	}
}

// ==================== V2.0 迭代9 API-104 会话列表测试 ====================

func TestGetSessionsByTeacherPersona_Basic(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)
	titleRepo := NewSessionTitleRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	// 创建学生和教师分身
	studentPersonaID := int64(101)
	teacherPersonaID1 := int64(201)
	teacherPersonaID2 := int64(202)

	// 创建用户（满足外键约束）
	studentID, err := userRepo.Create(&User{
		Username: "test_student",
		Password: "p",
		Role:     "student",
	})
	if err != nil {
		t.Fatalf("创建学生用户失败: %v", err)
	}
	teacherID, err := userRepo.Create(&User{
		Username: "test_teacher",
		Password: "p",
		Role:     "teacher",
	})
	if err != nil {
		t.Fatalf("创建教师用户失败: %v", err)
	}

	// 创建会话1：与教师201的会话
	sessionID1 := "sess_api104_001"
	for i := 1; i <= 3; i++ {
		conv := &Conversation{
			StudentID:        studentID,
			TeacherID:        teacherID,
			TeacherPersonaID: teacherPersonaID1,
			StudentPersonaID: studentPersonaID,
			SessionID:        sessionID1,
			Role:             "user",
			Content:          "消息" + string(rune('0'+i)),
			SenderType:       "student",
		}
		if i == 3 {
			conv.Role = "assistant"
			conv.SenderType = "ai"
		}
		_, err := convRepo.CreateWithSenderType(conv)
		if err != nil {
			t.Fatalf("创建会话1消息%d失败: %v", i, err)
		}
	}
	// 为会话1设置标题
	titleRepo.Upsert(sessionID1, studentPersonaID, teacherPersonaID1, "素描技法问答")

	// 创建会话2：与教师202的会话（最新，无标题）
	sessionID2 := "sess_api104_002"
	_, err = convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID2,
		StudentPersonaID: studentPersonaID,
		SessionID:        sessionID2,
		Role:             "user",
		Content:          "老师，色彩搭配有什么技巧？",
		SenderType:       "student",
	})
	if err != nil {
		t.Fatalf("创建会话2失败: %v", err)
	}

	// 测试：查询所有会话
	sessions, total, err := convRepo.GetSessionsByTeacherPersona(studentPersonaID, 0, 0, 10)
	if err != nil {
		t.Fatalf("查询会话列表失败: %v", err)
	}
	if total != 2 {
		t.Errorf("total 应为 2: got %d", total)
	}
	if len(sessions) != 2 {
		t.Fatalf("应返回 2 条会话: got %d", len(sessions))
	}

	// 验证第一条是最新的会话（无标题）
	if sessions[0].SessionID != sessionID2 {
		t.Errorf("第一条会话应为最新会话 %s: got %s", sessionID2, sessions[0].SessionID)
	}
	if sessions[0].Title != nil {
		t.Errorf("最新会话标题应为 nil: got %v", sessions[0].Title)
	}
	if sessions[0].MessageCount != 1 {
		t.Errorf("会话2消息数应为 1: got %d", sessions[0].MessageCount)
	}
	if sessions[0].LastMessage != "老师，色彩搭配有什么技巧？" {
		t.Errorf("LastMessage 不正确: got %s", sessions[0].LastMessage)
	}
	if sessions[0].LastMessageRole != "user" {
		t.Errorf("LastMessageRole 应为 'user': got %s", sessions[0].LastMessageRole)
	}

	// 验证第二条会话（有标题）
	if sessions[1].SessionID != sessionID1 {
		t.Errorf("第二条会话应为 %s: got %s", sessionID1, sessions[1].SessionID)
	}
	if sessions[1].Title == nil || *sessions[1].Title != "素描技法问答" {
		t.Errorf("会话1标题应为 '素描技法问答': got %v", sessions[1].Title)
	}
	if sessions[1].MessageCount != 3 {
		t.Errorf("会话1消息数应为 3: got %d", sessions[1].MessageCount)
	}

	// 验证两条会话都是活跃的（每个教师的最新会话）
	if !sessions[0].IsActive {
		t.Error("会话2（教师202的最新会话）应为活跃会话")
	}
	if !sessions[1].IsActive {
		t.Error("会话1（教师201的最新会话）应为活跃会话")
	}
}

func TestGetSessionsByTeacherPersona_WithTeacherFilter(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	studentPersonaID := int64(101)
	teacherPersonaID1 := int64(201)
	teacherPersonaID2 := int64(202)

	// 创建用户
	studentID, _ := userRepo.Create(&User{Username: "test_student_f", Password: "p", Role: "student"})
	teacherID, _ := userRepo.Create(&User{Username: "test_teacher_f", Password: "p", Role: "teacher"})

	// 创建与教师201的会话
	convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID1,
		StudentPersonaID: studentPersonaID,
		SessionID:        "sess_filter_001",
		Role:             "user",
		Content:          "教师201的消息",
		SenderType:       "student",
	})

	// 创建与教师202的会话
	convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID2,
		StudentPersonaID: studentPersonaID,
		SessionID:        "sess_filter_002",
		Role:             "user",
		Content:          "教师202的消息",
		SenderType:       "student",
	})

	// 测试：按教师201过滤
	sessions, total, err := convRepo.GetSessionsByTeacherPersona(studentPersonaID, teacherPersonaID1, 0, 10)
	if err != nil {
		t.Fatalf("查询会话列表失败: %v", err)
	}
	if total != 1 {
		t.Errorf("total 应为 1: got %d", total)
	}
	if len(sessions) != 1 {
		t.Fatalf("应返回 1 条会话: got %d", len(sessions))
	}
	if sessions[0].SessionID != "sess_filter_001" {
		t.Errorf("应返回教师201的会话: got %s", sessions[0].SessionID)
	}
	if !sessions[0].IsActive {
		t.Error("过滤后的唯一会话应为活跃会话")
	}
}

func TestGetSessionsByTeacherPersona_MultipleTeachersActive(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	studentPersonaID := int64(101)
	teacherPersonaID1 := int64(201)
	teacherPersonaID2 := int64(202)

	// 创建用户
	studentID, _ := userRepo.Create(&User{Username: "test_student_m", Password: "p", Role: "student"})
	teacherID, _ := userRepo.Create(&User{Username: "test_teacher_m", Password: "p", Role: "teacher"})

	// 创建与教师201的会话
	convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID1,
		StudentPersonaID: studentPersonaID,
		SessionID:        "sess_multi_001",
		Role:             "user",
		Content:          "消息1",
		SenderType:       "student",
	})

	// 创建与教师202的会话
	convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID2,
		StudentPersonaID: studentPersonaID,
		SessionID:        "sess_multi_002",
		Role:             "user",
		Content:          "消息2",
		SenderType:       "student",
	})

	// 测试：每个教师各自最新会话都应为活跃
	sessions, _, _ := convRepo.GetSessionsByTeacherPersona(studentPersonaID, 0, 0, 10)

	// 应该有2个活跃会话（每个教师一个）
	activeCount := 0
	for _, s := range sessions {
		if s.IsActive {
			activeCount++
		}
	}
	if activeCount != 2 {
		t.Errorf("应有 2 个活跃会话: got %d", activeCount)
	}
}

func TestGetSessionsByTeacherPersona_Pagination(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	studentPersonaID := int64(101)
	teacherPersonaID := int64(201)

	// 创建用户
	studentID, _ := userRepo.Create(&User{Username: "test_student_p", Password: "p", Role: "student"})
	teacherID, _ := userRepo.Create(&User{Username: "test_teacher_p", Password: "p", Role: "teacher"})

	// 创建3个会话
	for i := 1; i <= 3; i++ {
		convRepo.CreateWithSenderType(&Conversation{
			StudentID:        studentID,
			TeacherID:        teacherID,
			TeacherPersonaID: teacherPersonaID,
			StudentPersonaID: studentPersonaID,
			SessionID:        "sess_page_" + string(rune('0'+i)),
			Role:             "user",
			Content:          "消息",
			SenderType:       "student",
		})
	}

	// 测试分页
	sessions1, total, _ := convRepo.GetSessionsByTeacherPersona(studentPersonaID, 0, 0, 2) // offset=0, limit=2
	if total != 3 {
		t.Errorf("total 应为 3: got %d", total)
	}
	if len(sessions1) != 2 {
		t.Errorf("第一页应返回 2 条: got %d", len(sessions1))
	}

	sessions2, _, _ := convRepo.GetSessionsByTeacherPersona(studentPersonaID, 0, 2, 2) // offset=2, limit=2
	if len(sessions2) != 1 {
		t.Errorf("第二页应返回 1 条: got %d", len(sessions2))
	}
}

func TestGetSessionsByTeacherPersona_LongMessage(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)
	userRepo := NewUserRepository(db.DB)

	studentPersonaID := int64(101)
	teacherPersonaID := int64(201)

	// 创建用户
	studentID, _ := userRepo.Create(&User{Username: "test_student_l", Password: "p", Role: "student"})
	teacherID, _ := userRepo.Create(&User{Username: "test_teacher_l", Password: "p", Role: "teacher"})

	// 创建超长消息
	longMessage := ""
	for i := 0; i < 150; i++ {
		longMessage += "测"
	}

	convRepo.CreateWithSenderType(&Conversation{
		StudentID:        studentID,
		TeacherID:        teacherID,
		TeacherPersonaID: teacherPersonaID,
		StudentPersonaID: studentPersonaID,
		SessionID:        "sess_long_msg",
		Role:             "user",
		Content:          longMessage,
		SenderType:       "student",
	})

	sessions, _, _ := convRepo.GetSessionsByTeacherPersona(studentPersonaID, 0, 0, 10)

	if len(sessions) != 1 {
		t.Fatalf("应返回 1 条会话")
	}

	// 验证消息被截断至100字符
	runes := []rune(sessions[0].LastMessage)
	if len(runes) > 103 { // 100 + "..."
		t.Errorf("LastMessage 应被截断至 100 字符+省略号: got %d 字符", len(runes))
	}
}

func TestGetSessionsByTeacherPersona_Empty(t *testing.T) {
	db := setupTestDB(t)
	convRepo := NewConversationRepository(db.DB)

	// 不创建任何会话
	sessions, total, err := convRepo.GetSessionsByTeacherPersona(999, 0, 0, 10)
	if err != nil {
		t.Fatalf("查询空会话列表失败: %v", err)
	}
	if total != 0 {
		t.Errorf("total 应为 0: got %d", total)
	}
	if len(sessions) != 0 {
		t.Errorf("应返回空列表: got %d 条", len(sessions))
	}
}
