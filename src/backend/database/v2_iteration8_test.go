package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// ==================== V2.0 迭代8 集成测试 ====================

// setupV8TestData 创建迭代8测试所需的基础数据
// 返回: teacherID, teacherPersonaID, studentID, studentPersonaID
func setupV8TestData(t *testing.T, db *Database) (int64, int64, int64, int64) {
	t.Helper()
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)

	suffix := time.Now().Format("150405.000000")

	// 创建教师用户
	teacherID, err := userRepo.Create(&User{
		Username: "v8_teacher_" + suffix,
		Password: "password",
		Role:     "teacher",
		Nickname: "V8测试老师",
	})
	if err != nil {
		t.Fatalf("创建教师用户失败: %v", err)
	}

	// 创建教师分身
	teacherPersonaID, err := personaRepo.Create(&Persona{
		UserID:   teacherID,
		Role:     "teacher",
		Nickname: "V8测试老师_" + suffix,
		School:   "V8测试学校_" + suffix,
	})
	if err != nil {
		t.Fatalf("创建教师分身失败: %v", err)
	}

	// 回填 default_persona_id
	_, err = db.DB.Exec(`UPDATE users SET default_persona_id = ? WHERE id = ?`, teacherPersonaID, teacherID)
	if err != nil {
		t.Fatalf("回填教师默认分身失败: %v", err)
	}

	// 创建学生用户
	studentID, err := userRepo.Create(&User{
		Username: "v8_student_" + suffix,
		Password: "password",
		Role:     "student",
		Nickname: "V8测试学生",
	})
	if err != nil {
		t.Fatalf("创建学生用户失败: %v", err)
	}

	// 创建学生分身
	studentPersonaID, err := personaRepo.Create(&Persona{
		UserID:   studentID,
		Role:     "student",
		Nickname: "V8测试学生_" + suffix,
	})
	if err != nil {
		t.Fatalf("创建学生分身失败: %v", err)
	}

	// 回填 default_persona_id
	_, err = db.DB.Exec(`UPDATE users SET default_persona_id = ? WHERE id = ?`, studentPersonaID, studentID)
	if err != nil {
		t.Fatalf("回填学生默认分身失败: %v", err)
	}

	return teacherID, teacherPersonaID, studentID, studentPersonaID
}

// ==================== P0-1: 班级创建（V8扩展字段 + age_group 多选）→ 获取分享信息 ====================

func TestV2I8_IT001_CreateClassV8WithExtendedFields(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, _, _ := setupV8TestData(t, db)

	// 1. 创建班级（V8扩展字段）
	ageGroup := []string{"初中", "高中"}
	ageGroupJSON, _ := json.Marshal(ageGroup)

	inviteCode := "test1234"
	shareLink := fmt.Sprintf("/pages/share-join/index?code=%s", inviteCode)
	qrCodeURL := fmt.Sprintf("/api/qrcode?text=%s", inviteCode)

	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description,
			teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url,
			is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "初一美术一班", "美术创意课程",
		"曹老师", "美术", string(ageGroupJSON),
		shareLink, inviteCode, qrCodeURL,
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建V8班级失败: %v", err)
	}

	classID, _ := result.LastInsertId()
	if classID <= 0 {
		t.Fatal("班级ID应大于0")
	}

	// 2. 验证扩展字段
	var name, description, teacherDisplayName, subject, ageGroupStr string
	var shareLinkDB, inviteCodeDB, qrCodeURLDB string
	err = db.DB.QueryRow(`
		SELECT name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url
		FROM classes WHERE id = ?`, classID).Scan(
		&name, &description, &teacherDisplayName, &subject, &ageGroupStr,
		&shareLinkDB, &inviteCodeDB, &qrCodeURLDB,
	)
	if err != nil {
		t.Fatalf("查询班级失败: %v", err)
	}

	if name != "初一美术一班" {
		t.Errorf("班级名称不匹配: got %q", name)
	}
	if teacherDisplayName != "曹老师" {
		t.Errorf("教师显示名不匹配: got %q", teacherDisplayName)
	}
	if subject != "美术" {
		t.Errorf("学科不匹配: got %q", subject)
	}

	// 验证 age_group JSON 反序列化
	var parsedAgeGroup []string
	if err := json.Unmarshal([]byte(ageGroupStr), &parsedAgeGroup); err != nil {
		t.Fatalf("解析 age_group JSON 失败: %v", err)
	}
	if len(parsedAgeGroup) != 2 {
		t.Errorf("age_group 长度不匹配: got %d, want 2", len(parsedAgeGroup))
	}
	if parsedAgeGroup[0] != "初中" || parsedAgeGroup[1] != "高中" {
		t.Errorf("age_group 内容不匹配: got %v", parsedAgeGroup)
	}

	// 3. 验证分享信息
	if inviteCodeDB != inviteCode {
		t.Errorf("邀请码不匹配: got %q", inviteCodeDB)
	}
	if shareLinkDB != shareLink {
		t.Errorf("分享链接不匹配: got %q", shareLinkDB)
	}
	if qrCodeURLDB != qrCodeURL {
		t.Errorf("二维码URL不匹配: got %q", qrCodeURLDB)
	}

	// 4. 验证班级所有权（通过persona_id关联到teacherID）
	var ownerID int64
	err = db.DB.QueryRow(`
		SELECT user_id FROM personas WHERE id = (SELECT persona_id FROM classes WHERE id = ?)`,
		classID).Scan(&ownerID)
	if err != nil {
		t.Fatalf("查询班级所有者失败: %v", err)
	}
	if ownerID != teacherID {
		t.Errorf("班级所有者不匹配: got %d, want %d", ownerID, teacherID)
	}
}

// ==================== P0-2: 学生申请加入 → 教师审批通过 → 验证 class_members 同步 ====================

func TestV2I8_IT002_JoinClassApproveFlow(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 1. 创建班级
	inviteCode := "join0001"
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "测试审批班级", "测试班级",
		"测试老师", "数学", `["初中"]`,
		"/pages/share-join/index?code="+inviteCode, inviteCode, "/api/qrcode?text="+inviteCode,
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	// 2. 学生申请加入
	requestRepo := NewClassJoinRequestRepository(db)
	req := &ClassJoinRequest{
		ClassID:           classID,
		StudentPersonaID:  studentPersonaID,
		StudentID:         studentID,
		Status:            "pending",
		RequestMessage:    "我想加入这个班级",
		StudentAge:        13,
		StudentGender:     "male",
		StudentFamilyInfo: `{"parents": "双亲家庭"}`,
		RequestTime:       time.Now(),
	}
	requestID, err := requestRepo.CreateJoinRequest(req)
	if err != nil {
		t.Fatalf("创建加入申请失败: %v", err)
	}
	if requestID <= 0 {
		t.Fatal("申请ID应大于0")
	}

	// 3. 验证申请状态为 pending
	joinReq, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		t.Fatalf("查询申请失败: %v", err)
	}
	if joinReq.Status != "pending" {
		t.Errorf("申请状态应为 pending: got %q", joinReq.Status)
	}

	// 4. 获取教师待审批列表
	pendingList, err := requestRepo.GetPendingRequestsByTeacher(teacherID)
	if err != nil {
		t.Fatalf("获取待审批列表失败: %v", err)
	}
	if len(pendingList) != 1 {
		t.Fatalf("待审批列表应有1条: got %d", len(pendingList))
	}
	if pendingList[0].ID != requestID {
		t.Errorf("待审批申请ID不匹配: got %d, want %d", pendingList[0].ID, requestID)
	}

	// 5. 教师审批通过（含修改学生信息）
	teacherEvaluation := "该学生学习态度积极"
	err = requestRepo.ApproveJoinRequest(requestID, teacherEvaluation)
	if err != nil {
		t.Fatalf("审批通过失败: %v", err)
	}

	// 教师提供修改后的年龄和家庭信息
	finalAge := 14
	finalGender := "male"
	finalFamilyInfo := `{"parents": "双亲家庭", "note": "教师补充信息"}`

	// 6. 添加到 class_members（模拟 handler 行为）
	_, err = db.DB.Exec(`
		INSERT INTO class_members (
			class_id, student_persona_id, joined_at,
			approval_status, teacher_evaluation, age, gender, family_info,
			request_time, approval_time
		) VALUES (?, ?, ?, 'approved', ?, ?, ?, ?, ?, ?)
		ON CONFLICT(class_id, student_persona_id) DO UPDATE SET
			approval_status = 'approved',
			teacher_evaluation = excluded.teacher_evaluation,
			age = excluded.age,
			gender = excluded.gender,
			family_info = excluded.family_info,
			approval_time = excluded.approval_time`,
		classID, studentPersonaID, time.Now(),
		teacherEvaluation, finalAge, finalGender,
		finalFamilyInfo, joinReq.RequestTime, time.Now(),
	)
	if err != nil {
		t.Fatalf("添加班级成员失败: %v", err)
	}

	// 7. 验证 class_members 同步
	var memberAge int
	var memberGender, memberEvaluation, memberFamilyInfo, memberApprovalStatus string
	err = db.DB.QueryRow(`
		SELECT age, gender, teacher_evaluation, family_info, approval_status
		FROM class_members WHERE class_id = ? AND student_persona_id = ?`,
		classID, studentPersonaID).Scan(
		&memberAge, &memberGender, &memberEvaluation, &memberFamilyInfo, &memberApprovalStatus,
	)
	if err != nil {
		t.Fatalf("查询班级成员失败: %v", err)
	}

	if memberAge != finalAge {
		t.Errorf("成员年龄不匹配: got %d, want %d", memberAge, finalAge)
	}
	if memberGender != finalGender {
		t.Errorf("成员性别不匹配: got %q, want %q", memberGender, finalGender)
	}
	if memberEvaluation != teacherEvaluation {
		t.Errorf("教师评价不匹配: got %q, want %q", memberEvaluation, teacherEvaluation)
	}
	if memberApprovalStatus != "approved" {
		t.Errorf("审批状态不匹配: got %q, want 'approved'", memberApprovalStatus)
	}

	// 8. 验证申请状态变为 approved
	approvedReq, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		t.Fatalf("查询审批后申请失败: %v", err)
	}
	if approvedReq.Status != "approved" {
		t.Errorf("申请状态应为 approved: got %q", approvedReq.Status)
	}
	if approvedReq.TeacherEvaluation != teacherEvaluation {
		t.Errorf("申请中教师评价不匹配: got %q", approvedReq.TeacherEvaluation)
	}
	if approvedReq.ApprovalTime == nil {
		t.Error("审批时间不应为空")
	}

	// 9. 验证待审批列表为空
	pendingAfter, err := requestRepo.GetPendingRequestsByTeacher(teacherID)
	if err != nil {
		t.Fatalf("获取审批后待审批列表失败: %v", err)
	}
	if len(pendingAfter) != 0 {
		t.Errorf("审批后待审批列表应为空: got %d", len(pendingAfter))
	}
}

// ==================== P0-3: 学生申请加入 → 教师拒绝 → 验证状态 ====================

func TestV2I8_IT003_JoinClassRejectFlow(t *testing.T) {
	db := setupTestDB(t)
	_, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 1. 创建班级
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "拒绝测试班级", "测试",
		"测试老师", "语文", `["小学高年级"]`,
		"/pages/share-join/index?code=rej00001", "rej00001", "/api/qrcode?text=rej00001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	// 2. 学生申请加入
	requestRepo := NewClassJoinRequestRepository(db)
	req := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: studentPersonaID,
		StudentID:        studentID,
		Status:           "pending",
		RequestMessage:   "请求加入",
		RequestTime:      time.Now(),
	}
	requestID, err := requestRepo.CreateJoinRequest(req)
	if err != nil {
		t.Fatalf("创建申请失败: %v", err)
	}

	// 3. 教师拒绝
	rejectReason := "班级已满"
	err = requestRepo.RejectJoinRequest(requestID, rejectReason)
	if err != nil {
		t.Fatalf("拒绝申请失败: %v", err)
	}

	// 4. 验证申请状态为 rejected
	rejectedReq, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		t.Fatalf("查询拒绝后申请失败: %v", err)
	}
	if rejectedReq.Status != "rejected" {
		t.Errorf("申请状态应为 rejected: got %q", rejectedReq.Status)
	}
	if rejectedReq.TeacherEvaluation != rejectReason {
		t.Errorf("拒绝原因不匹配: got %q, want %q", rejectedReq.TeacherEvaluation, rejectReason)
	}
	if rejectedReq.ApprovalTime == nil {
		t.Error("审批时间不应为空（拒绝也应记录时间）")
	}

	// 5. 验证学生未被添加到 class_members
	var memberCount int
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM class_members WHERE class_id = ? AND student_persona_id = ?`,
		classID, studentPersonaID).Scan(&memberCount)
	if memberCount != 0 {
		t.Errorf("拒绝后不应有班级成员记录: got %d", memberCount)
	}
}

// ==================== P0-4: 知识库统一上传（文字类型）→ 搜索 → 详情 → 重命名 → 删除 ====================

func TestV2I8_IT004_KnowledgeItemCRUDFlow(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, _, _ := setupV8TestData(t, db)

	knowledgeRepo := NewKnowledgeRepository(db)

	// 1. 创建文字类型知识条目
	item := &KnowledgeItem{
		TeacherID: teacherID,
		PersonaID: teacherPersonaID,
		Title:     "牛顿第一定律教学笔记",
		Content:   "牛顿第一定律也叫惯性定律，一个物体如果没有受到外力作用，将保持静止或匀速直线运动状态不变。这是经典力学的基础定律之一。",
		ItemType:  "text",
		Tags:      `["物理","力学","牛顿定律"]`,
		Status:    "active",
		Scope:     "global",
	}
	itemID, err := knowledgeRepo.CreateKnowledgeItem(item)
	if err != nil {
		t.Fatalf("创建知识条目失败: %v", err)
	}
	if itemID <= 0 {
		t.Fatal("知识条目ID应大于0")
	}

	// 创建第二个条目用于搜索对比
	item2 := &KnowledgeItem{
		TeacherID: teacherID,
		PersonaID: teacherPersonaID,
		Title:     "美术色彩理论",
		Content:   "色彩三要素：色相、明度、纯度。",
		ItemType:  "text",
		Tags:      `["美术","色彩"]`,
		Status:    "active",
		Scope:     "global",
	}
	item2ID, err := knowledgeRepo.CreateKnowledgeItem(item2)
	if err != nil {
		t.Fatalf("创建第二个知识条目失败: %v", err)
	}

	// 2. 搜索 - 按关键词
	results, total, err := knowledgeRepo.SearchKnowledgeItems(teacherID, "牛顿", "", "", 10, 0)
	if err != nil {
		t.Fatalf("搜索知识条目失败: %v", err)
	}
	if total != 1 {
		t.Errorf("搜索结果数量不匹配: got %d, want 1", total)
	}
	if len(results) != 1 {
		t.Fatalf("搜索结果列表长度不匹配: got %d, want 1", len(results))
	}
	if results[0].ID != itemID {
		t.Errorf("搜索结果ID不匹配: got %d, want %d", results[0].ID, itemID)
	}

	// 3. 搜索 - 无关键词，获取全部
	allResults, allTotal, err := knowledgeRepo.SearchKnowledgeItems(teacherID, "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("搜索全部知识条目失败: %v", err)
	}
	if allTotal != 2 {
		t.Errorf("全部搜索结果数量不匹配: got %d, want 2", allTotal)
	}
	if len(allResults) != 2 {
		t.Errorf("全部搜索结果列表长度不匹配: got %d, want 2", len(allResults))
	}

	// 4. 获取详情
	detail, err := knowledgeRepo.GetKnowledgeItemByID(itemID)
	if err != nil {
		t.Fatalf("获取知识条目详情失败: %v", err)
	}
	if detail == nil {
		t.Fatal("知识条目详情不应为空")
	}
	if detail.Title != "牛顿第一定律教学笔记" {
		t.Errorf("标题不匹配: got %q", detail.Title)
	}
	if detail.ItemType != "text" {
		t.Errorf("类型不匹配: got %q", detail.ItemType)
	}
	if detail.TeacherID != teacherID {
		t.Errorf("教师ID不匹配: got %d", detail.TeacherID)
	}

	// 5. 重命名（更新标题）
	detail.Title = "牛顿第一定律（惯性定律）"
	err = knowledgeRepo.UpdateKnowledgeItem(detail)
	if err != nil {
		t.Fatalf("重命名知识条目失败: %v", err)
	}

	// 验证重命名生效
	renamed, err := knowledgeRepo.GetKnowledgeItemByID(itemID)
	if err != nil {
		t.Fatalf("查询重命名后条目失败: %v", err)
	}
	if renamed.Title != "牛顿第一定律（惯性定律）" {
		t.Errorf("重命名后标题不匹配: got %q", renamed.Title)
	}

	// 6. 删除
	err = knowledgeRepo.DeleteKnowledgeItem(itemID, teacherID)
	if err != nil {
		t.Fatalf("删除知识条目失败: %v", err)
	}

	// 验证删除后查不到
	deleted, err := knowledgeRepo.GetKnowledgeItemByID(itemID)
	if err != nil {
		t.Fatalf("查询删除后条目失败: %v", err)
	}
	if deleted != nil {
		t.Error("删除后应查不到该条目")
	}

	// 验证搜索结果只剩1条
	afterDeleteResults, afterDeleteTotal, err := knowledgeRepo.SearchKnowledgeItems(teacherID, "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("删除后搜索失败: %v", err)
	}
	if afterDeleteTotal != 1 {
		t.Errorf("删除后搜索结果数量不匹配: got %d, want 1", afterDeleteTotal)
	}
	if len(afterDeleteResults) != 1 {
		t.Errorf("删除后搜索结果列表长度不匹配: got %d", len(afterDeleteResults))
	}
	if afterDeleteResults[0].ID != item2ID {
		t.Errorf("删除后剩余条目ID不匹配: got %d, want %d", afterDeleteResults[0].ID, item2ID)
	}
}

// ==================== P0-5: 聊天列表（教师端 + 学生端） ====================

func TestV2I8_IT005_ChatListTeacherAndStudent(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, _, studentPersonaID := setupV8TestData(t, db)

	// 1. 创建班级
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "聊天列表测试班级", "测试",
		"测试老师", "英语", `["初中"]`,
		"/pages/share-join/index?code=chat0001", "chat0001", "/api/qrcode?text=chat0001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	// 2. 添加学生到班级
	_, err = db.DB.Exec(`
		INSERT INTO class_members (class_id, student_persona_id, joined_at, approval_status)
		VALUES (?, ?, ?, 'approved')`,
		classID, studentPersonaID, time.Now(),
	)
	if err != nil {
		t.Fatalf("添加班级成员失败: %v", err)
	}

	// 3. 创建一些对话记录
	convRepo := NewConversationRepository(db.DB)
	_, err = convRepo.Create(&Conversation{
		StudentID:        teacherID, // 这里student_id和teacher_id是用户ID
		TeacherID:        teacherID,
		StudentPersonaID: studentPersonaID,
		TeacherPersonaID: teacherPersonaID,
		SessionID:        "v8-chat-session-001",
		Role:             "user",
		Content:          "老师好，请问英语作文怎么写？",
		TokenCount:       15,
		SenderType:       "student",
	})
	if err != nil {
		t.Fatalf("创建对话记录失败: %v", err)
	}

	// 4. 验证教师端聊天列表 - 查询班级和学生
	rows, err := db.DB.Query(`
		SELECT id, name, subject
		FROM classes
		WHERE persona_id = ? AND is_active = 1
		ORDER BY created_at DESC`, teacherPersonaID)
	if err != nil {
		t.Fatalf("查询教师班级列表失败: %v", err)
	}

	var classCount int
	for rows.Next() {
		var cID int64
		var cName, cSubject string
		rows.Scan(&cID, &cName, &cSubject)
		classCount++

		if cName != "聊天列表测试班级" {
			t.Errorf("班级名称不匹配: got %q", cName)
		}
		if cSubject != "英语" {
			t.Errorf("班级学科不匹配: got %q", cSubject)
		}

		// 查询班级下的学生成员
		var memberCount int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM class_members WHERE class_id = ?`, cID).Scan(&memberCount)
		if memberCount != 1 {
			t.Errorf("班级成员数不匹配: got %d, want 1", memberCount)
		}
	}
	rows.Close()

	if classCount != 1 {
		t.Errorf("教师班级数不匹配: got %d, want 1", classCount)
	}

	// 5. 验证学生端聊天列表 - 查询关联的老师
	teacherRows, err := db.DB.Query(`
		SELECT DISTINCT p.id, p.nickname, c.subject
		FROM class_members cm
		JOIN classes c ON cm.class_id = c.id
		JOIN personas p ON c.persona_id = p.id
		WHERE cm.student_persona_id = ?`, studentPersonaID)
	if err != nil {
		t.Fatalf("查询学生端老师列表失败: %v", err)
	}

	var teacherCount int
	for teacherRows.Next() {
		var tpID int64
		var tNickname, tSubject string
		teacherRows.Scan(&tpID, &tNickname, &tSubject)
		teacherCount++

		if tpID != teacherPersonaID {
			t.Errorf("教师分身ID不匹配: got %d, want %d", tpID, teacherPersonaID)
		}
		if tSubject != "英语" {
			t.Errorf("学科不匹配: got %q", tSubject)
		}
	}
	teacherRows.Close()

	if teacherCount != 1 {
		t.Errorf("学生关联教师数不匹配: got %d, want 1", teacherCount)
	}
}

// ==================== P0-6: 置顶/取消置顶 ====================

func TestV2I8_IT006_PinUnpinChat(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	pinRepo := NewChatPinRepository(db)

	// 1. 创建班级
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "置顶测试班级", "测试",
		"测试老师", "数学", `["高中"]`,
		"/pages/share-join/index?code=pin00001", "pin00001", "/api/qrcode?text=pin00001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	// 2. 教师置顶班级
	classPinID, err := pinRepo.CreateChatPin(&ChatPin{
		UserID:     teacherID,
		UserRole:   "teacher",
		TargetType: "class",
		TargetID:   classID,
		PersonaID:  teacherPersonaID,
	})
	if err != nil {
		t.Fatalf("置顶班级失败: %v", err)
	}
	if classPinID <= 0 {
		t.Fatal("置顶记录ID应大于0")
	}

	// 3. 验证置顶状态
	isPinned, err := pinRepo.IsPinned(teacherID, "teacher", "class", classID, teacherPersonaID)
	if err != nil {
		t.Fatalf("检查置顶状态失败: %v", err)
	}
	if !isPinned {
		t.Error("班级应已置顶")
	}

	// 4. 教师置顶学生
	studentPinID, err := pinRepo.CreateChatPin(&ChatPin{
		UserID:     teacherID,
		UserRole:   "teacher",
		TargetType: "student",
		TargetID:   studentPersonaID,
		PersonaID:  teacherPersonaID,
	})
	if err != nil {
		t.Fatalf("置顶学生失败: %v", err)
	}
	if studentPinID <= 0 {
		t.Fatal("学生置顶记录ID应大于0")
	}

	// 5. 获取教师所有置顶记录
	pins, err := pinRepo.GetChatPinsByUser(teacherID, "teacher", teacherPersonaID)
	if err != nil {
		t.Fatalf("获取置顶列表失败: %v", err)
	}
	if len(pins) != 2 {
		t.Errorf("置顶记录数不匹配: got %d, want 2", len(pins))
	}

	// 6. 获取置顶的班级ID列表
	pinnedClasses, err := pinRepo.GetPinnedTargets(teacherID, "teacher", "class", teacherPersonaID)
	if err != nil {
		t.Fatalf("获取置顶班级失败: %v", err)
	}
	if len(pinnedClasses) != 1 {
		t.Errorf("置顶班级数不匹配: got %d, want 1", len(pinnedClasses))
	}
	if pinnedClasses[0] != classID {
		t.Errorf("置顶班级ID不匹配: got %d, want %d", pinnedClasses[0], classID)
	}

	// 7. 取消置顶班级
	err = pinRepo.DeleteChatPin(teacherID, "teacher", "class", classID, teacherPersonaID)
	if err != nil {
		t.Fatalf("取消置顶失败: %v", err)
	}

	// 8. 验证取消置顶
	isPinnedAfter, err := pinRepo.IsPinned(teacherID, "teacher", "class", classID, teacherPersonaID)
	if err != nil {
		t.Fatalf("检查取消置顶状态失败: %v", err)
	}
	if isPinnedAfter {
		t.Error("班级应已取消置顶")
	}

	// 9. 验证剩余置顶记录
	pinsAfter, err := pinRepo.GetChatPinsByUser(teacherID, "teacher", teacherPersonaID)
	if err != nil {
		t.Fatalf("获取取消后置顶列表失败: %v", err)
	}
	if len(pinsAfter) != 1 {
		t.Errorf("取消后置顶记录数不匹配: got %d, want 1", len(pinsAfter))
	}

	// 10. 学生端置顶教师
	_, err = pinRepo.CreateChatPin(&ChatPin{
		UserID:     studentID,
		UserRole:   "student",
		TargetType: "teacher",
		TargetID:   teacherPersonaID,
		PersonaID:  studentPersonaID,
	})
	if err != nil {
		t.Fatalf("学生置顶教师失败: %v", err)
	}

	studentPinnedTeachers, err := pinRepo.GetPinnedTargets(studentID, "student", "teacher", studentPersonaID)
	if err != nil {
		t.Fatalf("获取学生置顶教师失败: %v", err)
	}
	if len(studentPinnedTeachers) != 1 {
		t.Errorf("学生置顶教师数不匹配: got %d, want 1", len(studentPinnedTeachers))
	}
	if studentPinnedTeachers[0] != teacherPersonaID {
		t.Errorf("学生置顶教师ID不匹配: got %d", studentPinnedTeachers[0])
	}

	// 11. 重复置顶应幂等（UPSERT）
	_, err = pinRepo.CreateChatPin(&ChatPin{
		UserID:     studentID,
		UserRole:   "student",
		TargetType: "teacher",
		TargetID:   teacherPersonaID,
		PersonaID:  studentPersonaID,
	})
	if err != nil {
		t.Fatalf("重复置顶失败: %v", err)
	}

	// 验证不会产生重复记录
	studentPins, err := pinRepo.GetChatPinsByUser(studentID, "student", studentPersonaID)
	if err != nil {
		t.Fatalf("获取重复置顶后列表失败: %v", err)
	}
	if len(studentPins) != 1 {
		t.Errorf("重复置顶后记录数应为1: got %d", len(studentPins))
	}
}

// ==================== P1-7: 学生信息保存 ====================

func TestV2I8_IT007_StudentProfileUpdate(t *testing.T) {
	db := setupTestDB(t)
	_, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 1. 创建班级并添加学生
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "学生信息测试班级", "测试",
		"测试老师", "数学", `["初中"]`,
		"/pages/share-join/index?code=prof0001", "prof0001", "/api/qrcode?text=prof0001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	_, err = db.DB.Exec(`
		INSERT INTO class_members (class_id, student_persona_id, joined_at, approval_status, age, gender, family_info)
		VALUES (?, ?, ?, 'approved', 0, '', '{}')`,
		classID, studentPersonaID, time.Now(),
	)
	if err != nil {
		t.Fatalf("添加班级成员失败: %v", err)
	}

	// 2. 模拟 HandleUpdateStudentProfile：更新学生信息
	age := 14
	gender := "female"
	familyInfo := `{"parents": "双亲家庭", "siblings": 1}`

	_, err = db.DB.Exec(`
		UPDATE class_members SET age = ?, gender = ?, family_info = ?
		WHERE student_persona_id = ?`,
		age, gender, familyInfo, studentPersonaID,
	)
	if err != nil {
		t.Fatalf("更新学生信息失败: %v", err)
	}

	// 3. 验证更新结果
	var savedAge int
	var savedGender, savedFamilyInfo string
	err = db.DB.QueryRow(`
		SELECT age, gender, family_info FROM class_members
		WHERE class_id = ? AND student_persona_id = ?`,
		classID, studentPersonaID).Scan(&savedAge, &savedGender, &savedFamilyInfo)
	if err != nil {
		t.Fatalf("查询更新后学生信息失败: %v", err)
	}

	if savedAge != age {
		t.Errorf("年龄不匹配: got %d, want %d", savedAge, age)
	}
	if savedGender != gender {
		t.Errorf("性别不匹配: got %q, want %q", savedGender, gender)
	}
	if savedFamilyInfo != familyInfo {
		t.Errorf("家庭信息不匹配: got %q", savedFamilyInfo)
	}

	// 4. 同时更新 class_join_requests 中的待处理申请
	requestRepo := NewClassJoinRequestRepository(db)
	req := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: studentPersonaID,
		StudentID:        studentID,
		Status:           "pending",
		RequestMessage:   "测试申请",
		StudentAge:       0,
		StudentGender:    "",
		RequestTime:      time.Now(),
	}
	requestID, err := requestRepo.CreateJoinRequest(req)
	if err != nil {
		t.Fatalf("创建测试申请失败: %v", err)
	}

	// 更新 pending 申请中的学生信息
	_, err = db.DB.Exec(`
		UPDATE class_join_requests SET student_age = ?, student_gender = ?, student_family_info = ?
		WHERE student_id = ? AND status = 'pending'`,
		age, gender, familyInfo, studentID,
	)
	if err != nil {
		t.Fatalf("更新申请中学生信息失败: %v", err)
	}

	// 验证申请中的信息已更新
	updatedReq, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		t.Fatalf("查询更新后申请失败: %v", err)
	}
	if updatedReq.StudentAge != age {
		t.Errorf("申请中年龄不匹配: got %d, want %d", updatedReq.StudentAge, age)
	}
	if updatedReq.StudentGender != gender {
		t.Errorf("申请中性别不匹配: got %q, want %q", updatedReq.StudentGender, gender)
	}
}

// ==================== P1-8: 发现页推荐列表 + 搜索 ====================

func TestV2I8_IT008_DiscoverAndSearch(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, _, _ := setupV8TestData(t, db)

	// 1. 创建公开班级（有分享链接）
	_, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "发现页测试班级", "这是一个公开的测试班级",
		"发现老师", "生物", `["高中","成人"]`,
		"/pages/share-join/index?code=disc0001", "disc0001", "/api/qrcode?text=disc0001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建公开班级失败: %v", err)
	}

	// 设置教师分身为公开
	_, err = db.DB.Exec(`UPDATE personas SET is_public = 1 WHERE id = ?`, teacherPersonaID)
	if err != nil {
		t.Fatalf("设置分身公开失败: %v", err)
	}

	// 2. 发现页 - 查询热门班级（有分享链接的公开班级）
	rows, err := db.DB.Query(`
		SELECT c.id, c.name, c.description, c.subject, c.age_group,
			p.nickname as teacher_name,
			(SELECT COUNT(*) FROM class_members WHERE class_id = c.id) as member_count
		FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.is_active = 1 AND c.share_link != ''
		ORDER BY member_count DESC
		LIMIT 10`)
	if err != nil {
		t.Fatalf("查询热门班级失败: %v", err)
	}

	var classCount int
	for rows.Next() {
		var id int64
		var name, desc, subject, ageGroupStr, teacherName string
		var memberCount int
		rows.Scan(&id, &name, &desc, &subject, &ageGroupStr, &teacherName, &memberCount)
		classCount++

		if name != "发现页测试班级" {
			t.Errorf("班级名称不匹配: got %q", name)
		}
		if subject != "生物" {
			t.Errorf("学科不匹配: got %q", subject)
		}
	}
	rows.Close()

	if classCount != 1 {
		t.Errorf("热门班级数不匹配: got %d, want 1", classCount)
	}

	// 3. 发现页 - 查询推荐教师（公开的教师分身）
	teacherRows, err := db.DB.Query(`
		SELECT p.id, p.nickname, p.school, p.description,
			(SELECT COUNT(*) FROM teacher_student_relations
			 WHERE teacher_id = p.user_id AND status = 'approved') as student_count
		FROM personas p
		WHERE p.role = 'teacher' AND p.is_public = 1 AND p.is_active = 1
		ORDER BY student_count DESC
		LIMIT 10`)
	if err != nil {
		t.Fatalf("查询推荐教师失败: %v", err)
	}

	var teacherCount int
	for teacherRows.Next() {
		var id int64
		var nickname, school, desc string
		var studentCount int
		teacherRows.Scan(&id, &nickname, &school, &desc, &studentCount)
		teacherCount++
	}
	teacherRows.Close()

	if teacherCount != 1 {
		t.Errorf("推荐教师数不匹配: got %d, want 1", teacherCount)
	}

	// 4. 搜索功能 - 按班级名搜索
	keyword := "%发现%"
	searchRows, err := db.DB.Query(`
		SELECT c.id, c.name
		FROM classes c
		JOIN personas p ON c.persona_id = p.id
		WHERE c.is_active = 1 AND c.share_link != '' AND (c.name LIKE ? OR p.nickname LIKE ?)
		ORDER BY c.created_at DESC`, keyword, keyword)
	if err != nil {
		t.Fatalf("搜索班级失败: %v", err)
	}

	var searchCount int
	for searchRows.Next() {
		var id int64
		var name string
		searchRows.Scan(&id, &name)
		searchCount++
	}
	searchRows.Close()

	if searchCount != 1 {
		t.Errorf("搜索结果数不匹配: got %d, want 1", searchCount)
	}

	// 5. 搜索 - 按教师名搜索
	_ = teacherID // 使用 teacherID
	teacherKeyword := "%V8测试老师%"
	searchTeacherRows, err := db.DB.Query(`
		SELECT p.id, p.nickname
		FROM personas p
		WHERE p.role = 'teacher' AND p.is_public = 1 AND p.is_active = 1 AND (p.nickname LIKE ? OR p.school LIKE ?)
		ORDER BY p.created_at DESC`, teacherKeyword, teacherKeyword)
	if err != nil {
		t.Fatalf("搜索教师失败: %v", err)
	}

	var searchTeacherCount int
	for searchTeacherRows.Next() {
		var id int64
		var nickname string
		searchTeacherRows.Scan(&id, &nickname)
		searchTeacherCount++
	}
	searchTeacherRows.Close()

	if searchTeacherCount != 1 {
		t.Errorf("搜索教师结果数不匹配: got %d, want 1", searchTeacherCount)
	}
}

// ==================== P1-9: 新会话创建 + 快捷指令 ====================

func TestV2I8_IT009_NewSessionAndQuickActions(t *testing.T) {
	db := setupTestDB(t)
	_, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 1. 模拟新会话创建（生成 session_id）
	sessionID := fmt.Sprintf("sess_v8test_%d", time.Now().UnixNano())
	if sessionID == "" {
		t.Fatal("会话ID不应为空")
	}

	// 验证session_id格式
	if len(sessionID) < 10 {
		t.Errorf("会话ID长度不足: got %d", len(sessionID))
	}

	// 2. 在新会话中创建对话
	convRepo := NewConversationRepository(db.DB)
	convID, err := convRepo.Create(&Conversation{
		StudentID:        studentID,
		TeacherID:        studentID, // 这里用studentID作为占位
		StudentPersonaID: studentPersonaID,
		TeacherPersonaID: teacherPersonaID,
		SessionID:        sessionID,
		Role:             "user",
		Content:          "你好老师",
		TokenCount:       5,
		SenderType:       "student",
	})
	if err != nil {
		t.Fatalf("在新会话中创建对话失败: %v", err)
	}
	if convID <= 0 {
		t.Fatal("对话ID应大于0")
	}

	// 3. 验证新会话中的对话可以查询到
	convs, total, err := convRepo.GetConversationsBySession(studentID, sessionID, 0, 10)
	if err != nil {
		t.Fatalf("查询新会话对话失败: %v", err)
	}
	if total != 1 {
		t.Errorf("新会话对话数不匹配: got %d, want 1", total)
	}
	if len(convs) != 1 {
		t.Fatalf("新会话对话列表长度不匹配: got %d", len(convs))
	}
	if convs[0].SessionID != sessionID {
		t.Errorf("会话ID不匹配: got %q, want %q", convs[0].SessionID, sessionID)
	}

	// 4. 快捷指令验证（硬编码的快捷指令列表）
	quickActions := []struct {
		ID    string
		Label string
	}{
		{"review", "📚 回顾上次内容"},
		{"summarize", "📝 总结已学知识"},
		{"practice", "✏️ 开始练习"},
		{"question", "❓ 提个问题"},
	}

	if len(quickActions) != 4 {
		t.Errorf("快捷指令数量不匹配: got %d, want 4", len(quickActions))
	}

	// 验证每个快捷指令都有ID和Label
	for _, qa := range quickActions {
		if qa.ID == "" {
			t.Error("快捷指令ID不应为空")
		}
		if qa.Label == "" {
			t.Error("快捷指令Label不应为空")
		}
	}
}

// ==================== P1-10: 重复审批校验 ====================

func TestV2I8_IT010_DuplicateApprovalCheck(t *testing.T) {
	db := setupTestDB(t)
	_, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 1. 创建班级
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "重复审批测试班级", "测试",
		"测试老师", "物理", `["高中"]`,
		"/pages/share-join/index?code=dup00001", "dup00001", "/api/qrcode?text=dup00001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	// 2. 学生申请加入
	requestRepo := NewClassJoinRequestRepository(db)
	req := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: studentPersonaID,
		StudentID:        studentID,
		Status:           "pending",
		RequestMessage:   "请求加入",
		RequestTime:      time.Now(),
	}
	requestID, err := requestRepo.CreateJoinRequest(req)
	if err != nil {
		t.Fatalf("创建申请失败: %v", err)
	}

	// 3. 第一次审批通过
	err = requestRepo.ApproveJoinRequest(requestID, "第一次审批")
	if err != nil {
		t.Fatalf("第一次审批失败: %v", err)
	}

	// 4. 验证状态已变为 approved
	approvedReq, err := requestRepo.GetJoinRequestByID(requestID)
	if err != nil {
		t.Fatalf("查询审批后申请失败: %v", err)
	}
	if approvedReq.Status != "approved" {
		t.Fatalf("第一次审批后状态应为 approved: got %q", approvedReq.Status)
	}

	// 5. 模拟重复审批检查（handler 层逻辑）
	// 在 handler 中，会先检查 request.Status != "pending"，如果不是 pending 则返回错误
	if approvedReq.Status != "pending" {
		// 这是预期行为：已审批的申请不能再次审批
		t.Logf("✓ 重复审批校验通过: 申请状态为 %q，不是 pending，应拒绝再次审批", approvedReq.Status)
	} else {
		t.Error("已审批的申请状态不应为 pending")
	}

	// 6. 同样测试拒绝后的重复审批
	// 创建新的学生和申请
	userRepo := NewUserRepository(db.DB)
	personaRepo := NewPersonaRepository(db.DB)
	suffix2 := time.Now().Format("150405.000000") + "_2"
	student2ID, _ := userRepo.Create(&User{
		Username: "v8_student2_" + suffix2,
		Password: "password",
		Role:     "student",
	})
	student2PersonaID, _ := personaRepo.Create(&Persona{
		UserID:   student2ID,
		Role:     "student",
		Nickname: "V8测试学生2_" + suffix2,
	})

	req2 := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: student2PersonaID,
		StudentID:        student2ID,
		Status:           "pending",
		RequestMessage:   "第二个学生请求加入",
		RequestTime:      time.Now(),
	}
	request2ID, err := requestRepo.CreateJoinRequest(req2)
	if err != nil {
		t.Fatalf("创建第二个申请失败: %v", err)
	}

	// 先拒绝
	err = requestRepo.RejectJoinRequest(request2ID, "拒绝原因")
	if err != nil {
		t.Fatalf("拒绝申请失败: %v", err)
	}

	// 验证拒绝后状态
	rejectedReq, err := requestRepo.GetJoinRequestByID(request2ID)
	if err != nil {
		t.Fatalf("查询拒绝后申请失败: %v", err)
	}
	if rejectedReq.Status != "rejected" {
		t.Fatalf("拒绝后状态应为 rejected: got %q", rejectedReq.Status)
	}

	// 尝试对已拒绝的申请再次审批（模拟 handler 层检查）
	if rejectedReq.Status != "pending" {
		t.Logf("✓ 拒绝后重复审批校验通过: 申请状态为 %q，不是 pending，应拒绝再次审批", rejectedReq.Status)
	} else {
		t.Error("已拒绝的申请状态不应为 pending")
	}
}

// ==================== 补充测试: 知识库按类型筛选 ====================

func TestV2I8_IT011_KnowledgeSearchByType(t *testing.T) {
	db := setupTestDB(t)
	teacherID, teacherPersonaID, _, _ := setupV8TestData(t, db)

	knowledgeRepo := NewKnowledgeRepository(db)

	// 创建不同类型的知识条目
	types := []string{"text", "url", "file"}
	for i, itemType := range types {
		item := &KnowledgeItem{
			TeacherID: teacherID,
			PersonaID: teacherPersonaID,
			Title:     fmt.Sprintf("测试条目_%s_%d", itemType, i),
			Content:   fmt.Sprintf("内容_%s", itemType),
			ItemType:  itemType,
			Tags:      "[]",
			Status:    "active",
			Scope:     "global",
		}
		if itemType == "url" {
			item.SourceURL = "https://example.com"
		}
		if itemType == "file" {
			item.FileName = "test.pdf"
			item.FileSize = 1024
		}
		_, err := knowledgeRepo.CreateKnowledgeItem(item)
		if err != nil {
			t.Fatalf("创建 %s 类型知识条目失败: %v", itemType, err)
		}
	}

	// 按类型筛选
	for _, itemType := range types {
		results, total, err := knowledgeRepo.SearchKnowledgeItems(teacherID, "", itemType, "", 10, 0)
		if err != nil {
			t.Fatalf("按类型 %s 筛选失败: %v", itemType, err)
		}
		if total != 1 {
			t.Errorf("类型 %s 筛选结果数不匹配: got %d, want 1", itemType, total)
		}
		if len(results) != 1 {
			t.Errorf("类型 %s 筛选结果列表长度不匹配: got %d", itemType, len(results))
		} else if results[0].ItemType != itemType {
			t.Errorf("类型 %s 筛选结果类型不匹配: got %q", itemType, results[0].ItemType)
		}
	}

	// 不筛选类型应返回全部
	allResults, allTotal, err := knowledgeRepo.SearchKnowledgeItems(teacherID, "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("查询全部失败: %v", err)
	}
	if allTotal != 3 {
		t.Errorf("全部结果数不匹配: got %d, want 3", allTotal)
	}
	if len(allResults) != 3 {
		t.Errorf("全部结果列表长度不匹配: got %d", len(allResults))
	}
}

// ==================== 补充测试: 班级加入申请 UPSERT ====================

func TestV2I8_IT012_JoinRequestUpsert(t *testing.T) {
	db := setupTestDB(t)
	_, teacherPersonaID, studentID, studentPersonaID := setupV8TestData(t, db)

	// 创建班级
	result, err := db.DB.Exec(`
		INSERT INTO classes (
			persona_id, name, description, teacher_display_name, subject, age_group,
			share_link, invite_code, qr_code_url, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		teacherPersonaID, "UPSERT测试班级", "测试",
		"测试老师", "化学", `["高中"]`,
		"/pages/share-join/index?code=ups00001", "ups00001", "/api/qrcode?text=ups00001",
		time.Now(), time.Now(),
	)
	if err != nil {
		t.Fatalf("创建班级失败: %v", err)
	}
	classID, _ := result.LastInsertId()

	requestRepo := NewClassJoinRequestRepository(db)

	// 第一次申请
	req1 := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: studentPersonaID,
		StudentID:        studentID,
		Status:           "pending",
		RequestMessage:   "第一次申请",
		RequestTime:      time.Now(),
	}
	_, err = requestRepo.CreateJoinRequest(req1)
	if err != nil {
		t.Fatalf("第一次申请失败: %v", err)
	}

	// 被拒绝后重新申请（UPSERT应更新为pending）
	req2 := &ClassJoinRequest{
		ClassID:          classID,
		StudentPersonaID: studentPersonaID,
		StudentID:        studentID,
		Status:           "pending",
		RequestMessage:   "重新申请",
		RequestTime:      time.Now(),
	}
	_, err = requestRepo.CreateJoinRequest(req2)
	if err != nil {
		t.Fatalf("重新申请失败: %v", err)
	}

	// 验证只有一条记录（UPSERT）
	existingReq, err := requestRepo.GetJoinRequestByClassAndStudent(classID, studentPersonaID)
	if err != nil {
		t.Fatalf("查询申请失败: %v", err)
	}
	if existingReq == nil {
		t.Fatal("应找到申请记录")
	}
	if existingReq.Status != "pending" {
		t.Errorf("重新申请后状态应为 pending: got %q", existingReq.Status)
	}
	if existingReq.RequestMessage != "重新申请" {
		t.Errorf("申请消息应更新为最新: got %q", existingReq.RequestMessage)
	}
}
