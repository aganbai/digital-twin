package knowledge

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 测试数据目录
const testDataDir = "../../data/prebuilt-knowledge-packs/knowledge_packs"

// 测试创建管理器
func TestNewPackManager(t *testing.T) {
	config := PackManagerConfig{
		BasePath:    testDataDir,
		EnableCache: true,
	}

	manager, err := NewPackManager(config)
	require.NoError(t, err, "创建包管理器失败")
	require.NotNil(t, manager, "管理器不能为nil")

	// 验证索引加载
	assert.Greater(t, manager.Count(), 0, "包数量应该大于0")

	// 验证统计信息
	stats := manager.GetStats()
	assert.True(t, stats["loaded"].(bool), "应该已加载")
	assert.Greater(t, stats["total_packs"].(int), 0, "总包数应该大于0")
}

// 测试获取所有包
func TestGetAllPacks(t *testing.T) {
	manager := createTestManager(t)

	packs := manager.GetAllPacks()
	assert.NotEmpty(t, packs, "应该返回包列表")

	// 验证第一个包的结构
	if len(packs) > 0 {
		pack := packs[0]
		assert.NotEmpty(t, pack.ID, "包ID不能为空")
		assert.NotEmpty(t, pack.Name, "包名称不能为空")
		assert.NotEmpty(t, pack.Subject, "学科不能为空")
		assert.NotEmpty(t, pack.SubjectName, "学科名称不能为空")
		assert.NotEmpty(t, pack.Level, "年级不能为空")
		assert.NotEmpty(t, pack.Status, "状态不能为空")
	}
}

// 测试获取特定包
func TestGetPack(t *testing.T) {
	manager := createTestManager(t)

	// 测试存在的包
	pack, err := manager.GetPack("math_小学")
	require.NoError(t, err, "获取数学小学包失败")
	require.NotNil(t, pack, "包不能为nil")

	assert.Equal(t, "math_小学", pack.ID, "包ID应该匹配")
	assert.Equal(t, "数学", pack.SubjectName, "学科名称应该匹配")
	assert.Equal(t, "小学", pack.Level, "年级应该匹配")

	// 测试不存在的包
	_, err = manager.GetPack("nonexistent_pack")
	assert.Error(t, err, "应该返回错误")
	assert.Contains(t, err.Error(), "不存在", "错误消息应该包含'不存在'")
}

// 测试按学科年级获取包
func TestGetPackBySubjectLevel(t *testing.T) {
	manager := createTestManager(t)

	pack, err := manager.GetPackBySubjectLevel("physics", "初中")
	require.NoError(t, err, "获取物理初中包失败")
	require.NotNil(t, pack, "包不能为nil")

	assert.Equal(t, "physics_初中", pack.ID, "包ID应该匹配")
	assert.Equal(t, "物理", pack.SubjectName, "学科名称应该匹配")
	assert.Equal(t, "初中", pack.Level, "年级应该匹配")
}

// 测试按学科查询
func TestGetPacksBySubject(t *testing.T) {
	manager := createTestManager(t)

	// 测试数学学科
	mathPacks := manager.GetPacksBySubject("math")
	require.NotEmpty(t, mathPacks, "数学包不应该为空")

	// 验证排序（应该按年级排序）
	expectedLevels := []string{"小学", "初中", "高中", "大学"}
	for i, pack := range mathPacks {
		assert.Equal(t, "math", pack.Subject, "学科应该都是数学")
		if i < len(expectedLevels) {
			assert.Equal(t, expectedLevels[i], pack.Level, "年级排序应该正确")
		}
	}

	// 测试不存在的学科
	nonexistentPacks := manager.GetPacksBySubject("nonexistent")
	assert.Empty(t, nonexistentPacks, "不存在的学科应该返回空列表")
}

// 测试按年级查询
func TestGetPacksByLevel(t *testing.T) {
	manager := createTestManager(t)

	// 测试小学年级
	primaryPacks := manager.GetPacksByLevel("小学")
	require.NotEmpty(t, primaryPacks, "小学包不应该为空")

	// 验证排序（应该按学科排序）
	expectedSubjects := []string{"math", "chinese", "english", "music", "art", "pe"}
	for i, pack := range primaryPacks {
		assert.Equal(t, "小学", pack.Level, "年级应该都是小学")
		if i < len(expectedSubjects) {
			assert.Equal(t, expectedSubjects[i], pack.Subject, "学科排序应该正确")
		}
	}
}

// 测试获取所有学科
func TestGetAllSubjects(t *testing.T) {
	manager := createTestManager(t)

	subjects := manager.GetAllSubjects()
	require.NotEmpty(t, subjects, "学科列表不应该为空")

	expectedSubjects := []string{
		"math", "physics", "chemistry", "biology",
		"chinese", "english", "history", "geography",
		"politics", "music", "art", "pe",
	}

	// 验证数量和顺序
	assert.Equal(t, len(expectedSubjects), len(subjects), "学科数量应该匹配")
	for i, subject := range subjects {
		assert.Equal(t, expectedSubjects[i], subject, "学科顺序应该正确")
	}
}

// 测试获取所有年级
func TestGetAllLevels(t *testing.T) {
	manager := createTestManager(t)

	levels := manager.GetAllLevels()
	require.NotEmpty(t, levels, "年级列表不应该为空")

	expectedLevels := []string{"小学", "初中", "高中", "大学"}

	// 验证数量和顺序
	assert.Equal(t, len(expectedLevels), len(levels), "年级数量应该匹配")
	for i, level := range levels {
		assert.Equal(t, expectedLevels[i], level, "年级顺序应该正确")
	}
}

// 测试学科名称转换
func TestGetSubjectName(t *testing.T) {
	manager := createTestManager(t)

	testCases := []struct {
		subject  string
		expected string
	}{
		{"math", "数学"},
		{"physics", "物理"},
		{"chemistry", "化学"},
		{"biology", "生物"},
		{"chinese", "语文"},
		{"english", "英语"},
		{"history", "历史"},
		{"geography", "地理"},
		{"politics", "政治"},
		{"music", "音乐"},
		{"art", "美术"},
		{"pe", "体育"},
		{"unknown", "unknown"},
	}

	for _, tc := range testCases {
		name := manager.GetSubjectName(tc.subject)
		assert.Equal(t, tc.expected, name, "学科名称转换应该正确")
	}
}

// 测试包存在检查
func TestHasPack(t *testing.T) {
	manager := createTestManager(t)

	assert.True(t, manager.HasPack("math_小学"), "数学小学包应该存在")
	assert.True(t, manager.HasPack("physics_初中"), "物理初中包应该存在")
	assert.True(t, manager.HasPack("chemistry_高中"), "化学高中包应该存在")

	assert.False(t, manager.HasPack("nonexistent_pack"), "不存在的包应该返回false")
	assert.False(t, manager.HasPack(""), "空字符串应该返回false")
}

// 测试重新加载
func TestReload(t *testing.T) {
	manager := createTestManager(t)

	originalCount := manager.Count()
	assert.Greater(t, originalCount, 0, "原始包数量应该大于0")

	// 重新加载
	err := manager.Reload()
	require.NoError(t, err, "重新加载失败")

	reloadedCount := manager.Count()
	assert.Equal(t, originalCount, reloadedCount, "重新加载后包数量应该不变")
}

// 测试验证包完整性
func TestValidatePack(t *testing.T) {
	manager := createTestManager(t)

	// 验证存在的包
	err := manager.ValidatePack("math_小学")
	assert.NoError(t, err, "验证数学小学包应该成功")

	// 验证不存在的包
	err = manager.ValidatePack("nonexistent_pack")
	assert.Error(t, err, "验证不存在的包应该失败")
}

// 测试验证所有包
func TestValidateAll(t *testing.T) {
	manager := createTestManager(t)

	results := manager.ValidateAll()
	require.NotNil(t, results, "验证结果不应该为nil")

	// 统计成功和失败的包
	successCount := 0
	for _, err := range results {
		if err == nil {
			successCount++
		}
	}

	assert.Greater(t, successCount, 0, "至少应该有成功的验证")
	assert.Equal(t, manager.Count(), len(results), "验证结果数量应该等于包数量")
}

// 测试搜索包
func TestSearchPacks(t *testing.T) {
	manager := createTestManager(t)

	// 搜索数学
	mathResults := manager.SearchPacks("数学")
	assert.NotEmpty(t, mathResults, "搜索'数学'应该返回结果")

	for _, pack := range mathResults {
		assert.Contains(t, pack.Name, "数学", "结果应该包含'数学'")
	}

	// 搜索小学
	primaryResults := manager.SearchPacks("小学")
	assert.NotEmpty(t, primaryResults, "搜索'小学'应该返回结果")

	for _, pack := range primaryResults {
		assert.Contains(t, pack.Level, "小学", "结果应该包含'小学'")
	}

	// 搜索空字符串
	emptyResults := manager.SearchPacks("")
	assert.Empty(t, emptyResults, "搜索空字符串应该返回空")

	// 搜索不存在的
	nonexistentResults := manager.SearchPacks("nonexistent")
	assert.Empty(t, nonexistentResults, "搜索不存在的内容应该返回空")
}

// 测试API响应格式
func TestToAPIResponse(t *testing.T) {
	manager := createTestManager(t)

	response := manager.ToAPIResponse()
	require.NotNil(t, response, "API响应不应该为nil")

	// 验证响应结构
	stats, ok := response["stats"].(gin.H)
	require.True(t, ok, "stats字段应该是gin.H类型")
	assert.True(t, stats["loaded"].(bool), "stats.loaded应该是true")

	subjects, ok := response["subjects"].([]string)
	require.True(t, ok, "subjects字段应该是[]string类型")
	assert.NotEmpty(t, subjects, "subjects不应该为空")

	packs, ok := response["packs"].([]PackMetadata)
	require.True(t, ok, "packs字段应该是[]PackMetadata类型")
	assert.NotEmpty(t, packs, "packs不应该为空")
}

// 测试获取包路径
func TestGetPackPath(t *testing.T) {
	manager := createTestManager(t)

	// 测试存在的包
	path, err := manager.GetPackPath("math_小学")
	require.NoError(t, err, "获取数学小学包路径失败")

	expectedPath := filepath.Join(testDataDir, "math_小学")
	assert.Equal(t, expectedPath, path, "包路径应该正确")

	// 验证路径是否存在
	_, err = os.Stat(path)
	assert.NoError(t, err, "包目录应该存在")

	// 测试不存在的包
	_, err = manager.GetPackPath("nonexistent_pack")
	assert.Error(t, err, "获取不存在的包路径应该失败")
}

// 测试推荐包
func TestRecommendPacks(t *testing.T) {
	manager := createTestManager(t)

	// 测试有匹配的情况
	userInfo := gin.H{
		"preferred_subject": "math",
		"grade_level":       "小学",
	}

	recommendations := manager.RecommendPacks(userInfo)
	require.NotEmpty(t, recommendations, "应该有推荐包")

	firstPack := recommendations[0]
	assert.Equal(t, "math_小学", firstPack.ID, "应该推荐数学小学包")

	// 测试无匹配但有同学科的情况
	userInfo2 := gin.H{
		"preferred_subject": "music",
		"grade_level":       "大学", // 音乐没有大学年级
	}

	recommendations2 := manager.RecommendPacks(userInfo2)
	require.NotEmpty(t, recommendations2, "应该有推荐包")

	for _, pack := range recommendations2 {
		assert.Equal(t, "music", pack.Subject, "应该推荐音乐学科包")
	}

	// 测试完全无匹配的情况
	userInfo3 := gin.H{
		"preferred_subject": "nonexistent",
		"grade_level":       "nonexistent",
	}

	recommendations3 := manager.RecommendPacks(userInfo3)
	assert.Empty(t, recommendations3, "应该返回空列表")
}

// 测试基础路径不存在的情况
func TestBasePathNotFound(t *testing.T) {
	config := PackManagerConfig{
		BasePath: "nonexistent/path",
	}

	_, err := NewPackManager(config)
	assert.Error(t, err, "应该返回错误")
	assert.Contains(t, err.Error(), "失败", "错误消息应该包含'失败'")
}

// 辅助函数：创建测试管理器
func createTestManager(t *testing.T) *PackManager {
	t.Helper()

	config := PackManagerConfig{
		BasePath:    testDataDir,
		EnableCache: true,
	}

	manager, err := NewPackManager(config)
	require.NoError(t, err, "创建测试管理器失败")
	require.NotNil(t, manager, "管理器不能为nil")

	return manager
}
