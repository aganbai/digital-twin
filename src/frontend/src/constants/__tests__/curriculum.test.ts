/**
 * Curriculum Constants 单元测试
 * 模块: FE-IT13-002 - 小程序-班级编辑页教材配置区域
 */

import {
  GRADE_LEVEL_OPTIONS,
  GRADE_MAP,
  K12_SUBJECTS,
  TEXTBOOK_VERSIONS,
  ADULT_LIFE_CATEGORIES,
  ADULT_PROFESSIONAL_CATEGORIES,
  getGradesByLevel,
  needsGradeSelection,
  needsCustomTextbook,
  getSubjectsByLevel,
  EMPTY_CURRICULUM_CONFIG,
  type GradeLevel,
} from '../curriculum'

describe('Curriculum Constants', () => {
  describe('学段选项', () => {
    it('GRADE_LEVEL_OPTIONS 应包含所有学段', () => {
      expect(GRADE_LEVEL_OPTIONS).toHaveLength(8)

      // 验证每个学段都有 value 和 label
      GRADE_LEVEL_OPTIONS.forEach(option => {
        expect(option).toHaveProperty('value')
        expect(option).toHaveProperty('label')
        expect(typeof option.value).toBe('string')
        expect(typeof option.label).toBe('string')
      })
    })

    it('应包含预期的学段值', () => {
      const values = GRADE_LEVEL_OPTIONS.map(o => o.value)
      expect(values).toContain('preschool')
      expect(values).toContain('primary_lower')
      expect(values).toContain('primary_upper')
      expect(values).toContain('junior')
      expect(values).toContain('senior')
      expect(values).toContain('university')
      expect(values).toContain('adult_life')
      expect(values).toContain('adult_professional')
    })
  })

  describe('年级映射 GRADE_MAP', () => {
    it('小学低年级应有3个年级', () => {
      expect(GRADE_MAP.primary_lower).toHaveLength(3)
      expect(GRADE_MAP.primary_lower).toContain('一年级')
      expect(GRADE_MAP.primary_lower).toContain('二年级')
      expect(GRADE_MAP.primary_lower).toContain('三年级')
    })

    it('小学高年级应有3个年级', () => {
      expect(GRADE_MAP.primary_upper).toHaveLength(3)
      expect(GRADE_MAP.primary_upper).toContain('四年级')
      expect(GRADE_MAP.primary_upper).toContain('五年级')
      expect(GRADE_MAP.primary_upper).toContain('六年级')
    })

    it('初中应有3个年级', () => {
      expect(GRADE_MAP.junior).toHaveLength(3)
      expect(GRADE_MAP.junior).toContain('七年级')
      expect(GRADE_MAP.junior).toContain('八年级')
      expect(GRADE_MAP.junior).toContain('九年级')
    })

    it('高中应有3个年级', () => {
      expect(GRADE_MAP.senior).toHaveLength(3)
      expect(GRADE_MAP.senior).toContain('高一')
      expect(GRADE_MAP.senior).toContain('高二')
      expect(GRADE_MAP.senior).toContain('高三')
    })

    it('大学及以上学段应有更多年级选项', () => {
      expect(GRADE_MAP.university).toContain('大一')
      expect(GRADE_MAP.university).toContain('大二')
      expect(GRADE_MAP.university).toContain('大三')
      expect(GRADE_MAP.university).toContain('大四')
      expect(GRADE_MAP.university).toContain('研究生')
      expect(GRADE_MAP.university).toContain('博士')
    })

    it('成人培训学段应无年级选项', () => {
      expect(GRADE_MAP.adult_life).toHaveLength(0)
      expect(GRADE_MAP.adult_professional).toHaveLength(0)
    })
  })

  describe('getGradesByLevel helper函数', () => {
    it('应正确返回对应学段的年级', () => {
      const grades = getGradesByLevel('primary_lower')
      expect(grades).toEqual(['一年级', '二年级', '三年级'])
    })

    it('应正确返回大学年级', () => {
      const grades = getGradesByLevel('university')
      expect(grades).toHaveLength(6)
    })

    it('对无效学段应返回空数组', () => {
      const grades = getGradesByLevel('invalid' as GradeLevel)
      expect(grades).toEqual([])
    })
  })

  describe('needsGradeSelection helper函数', () => {
    it('K12学段应需要年级选择', () => {
      expect(needsGradeSelection('preschool')).toBe(true)
      expect(needsGradeSelection('primary_lower')).toBe(true)
      expect(needsGradeSelection('primary_upper')).toBe(true)
      expect(needsGradeSelection('junior')).toBe(true)
      expect(needsGradeSelection('senior')).toBe(true)
    })

    it('大学及以上也应需要年级选择', () => {
      expect(needsGradeSelection('university')).toBe(true)
    })

    it('成人培训不需要年级选择', () => {
      expect(needsGradeSelection('adult_life')).toBe(false)
      expect(needsGradeSelection('adult_professional')).toBe(false)
    })
  })

  describe('needsCustomTextbook helper函数', () => {
    it('大学及以上应需要自定义教材', () => {
      expect(needsCustomTextbook('university')).toBe(true)
    })

    it('成人培训应需要自定义教材', () => {
      expect(needsCustomTextbook('adult_life')).toBe(true)
      expect(needsCustomTextbook('adult_professional')).toBe(true)
    })

    it('K12学段不需要自定义教材', () => {
      expect(needsCustomTextbook('primary_lower')).toBe(false)
      expect(needsCustomTextbook('junior')).toBe(false)
      expect(needsCustomTextbook('senior')).toBe(false)
    })
  })

  describe('学科列表', () => {
    it('K12学科应包含所有基础学科', () => {
      expect(K12_SUBJECTS).toContain('语文')
      expect(K12_SUBJECTS).toContain('数学')
      expect(K12_SUBJECTS).toContain('英语')
      expect(K12_SUBJECTS).toContain('物理')
      expect(K12_SUBJECTS).toContain('化学')
      expect(K12_SUBJECTS).toContain('生物')
    })

    it('应包含艺术和体育学科', () => {
      expect(K12_SUBJECTS).toContain('音乐')
      expect(K12_SUBJECTS).toContain('美术')
      expect(K12_SUBJECTS).toContain('体育')
    })
  })

  describe('成人课程类别', () => {
    it('生活技能类别应包含生活相关课程', () => {
      expect(ADULT_LIFE_CATEGORIES).toContain('中餐')
      expect(ADULT_LIFE_CATEGORIES).toContain('西餐')
      expect(ADULT_LIFE_CATEGORIES).toContain('烘焙')
      expect(ADULT_LIFE_CATEGORIES).toContain('瑜伽')
    })

    it('职业培训类别应包含职业技能课程', () => {
      expect(ADULT_PROFESSIONAL_CATEGORIES).toContain('编程')
      expect(ADULT_PROFESSIONAL_CATEGORIES).toContain('设计')
      expect(ADULT_PROFESSIONAL_CATEGORIES).toContain('会计')
      expect(ADULT_PROFESSIONAL_CATEGORIES).toContain('医学')
    })
  })

  describe('getSubjectsByLevel helper函数', () => {
    it('K12学段应返回K12学科', () => {
      const subjects = getSubjectsByLevel('primary_lower')
      expect(subjects).toEqual(K12_SUBJECTS)
    })

    it('成人生活应返回生活技能类别', () => {
      const subjects = getSubjectsByLevel('adult_life')
      expect(subjects).toEqual(ADULT_LIFE_CATEGORIES)
    })

    it('成人职业培训应返回职业培训类别', () => {
      const subjects = getSubjectsByLevel('adult_professional')
      expect(subjects).toEqual(ADULT_PROFESSIONAL_CATEGORIES)
    })

    it('大学应返回K12学科', () => {
      const subjects = getSubjectsByLevel('university')
      expect(subjects).toEqual(K12_SUBJECTS)
    })
  })

  describe('教材版本选项', () => {
    it('应包含常见教材版本', () => {
      expect(TEXTBOOK_VERSIONS).toContain('人教版')
      expect(TEXTBOOK_VERSIONS).toContain('北师大版')
      expect(TEXTBOOK_VERSIONS).toContain('苏教版')
      expect(TEXTBOOK_VERSIONS).toContain('沪教版')
      expect(TEXTBOOK_VERSIONS).toContain('部编版')
    })
  })

  describe('空配置默认值', () => {
    it('EMPTY_CURRICULUM_CONFIG 应有正确的初始值', () => {
      expect(EMPTY_CURRICULUM_CONFIG.grade_level).toBeUndefined()
      expect(EMPTY_CURRICULUM_CONFIG.grade).toBeUndefined()
      expect(EMPTY_CURRICULUM_CONFIG.subjects).toEqual([])
      expect(EMPTY_CURRICULUM_CONFIG.textbook_versions).toEqual([])
      expect(EMPTY_CURRICULUM_CONFIG.custom_textbooks).toEqual([])
      expect(EMPTY_CURRICULUM_CONFIG.current_progress).toBe('')
    })
  })
})
