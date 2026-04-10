/**
 * Class API 单元测试
 * 模块: FE-IT13-002 - 小程序-班级编辑页教材配置区域
 */

import {
  createClassV11,
  updateClassV11,
  getClassDetail,
  createClass,
  updateClass,
  deleteClass,
  type CurriculumConfig,
  type CreateClassV11Params,
  type UpdateClassV11Params,
} from '@/api/class'
import { server, mockClassDetail, mockClassWithCurriculum } from '@/__mocks__/server'

describe('Class API', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('createClassV11 - 创建班级 V11', () => {
    it('应正确导出 createClassV11 函数', () => {
      expect(typeof createClassV11).toBe('function')
    })

    it('CreateClassV11Params 类型应正确', () => {
      // 验证类型定义 - 完整配置
      const fullParams: CreateClassV11Params = {
        name: '三年级数学班',
        description: '培优班',
        persona_nickname: '王老师',
        persona_school: '实验小学',
        persona_description: '资深教师',
        is_public: true,
        curriculum_config: {
          grade_level: 'primary_lower',
          grade: '三年级',
          subjects: ['数学'],
          textbook_versions: ['人教版'],
          custom_textbooks: [],
          current_progress: '第三单元',
        },
      }

      expect(fullParams.name).toBe('三年级数学班')
      expect(fullParams.curriculum_config?.grade_level).toBe('primary_lower')

      // 最小配置
      const minParams: CreateClassV11Params = {
        name: '基础班级',
        persona_nickname: '王老师',
        persona_school: '实验小学',
        persona_description: '教师',
      }

      expect(minParams.is_public).toBeUndefined()
    })
  })

  describe('updateClassV11 - 更新班级 V11', () => {
    it('应正确导出 updateClassV11 函数', () => {
      expect(typeof updateClassV11).toBe('function')
    })

    it('UpdateClassV11Params 类型应允许部分更新', () => {
      // 只更新名称
      const nameOnly: UpdateClassV11Params = {
        name: '新名称',
      }
      expect(nameOnly.name).toBe('新名称')

      // 只更新配置
      const configOnly: UpdateClassV11Params = {
        curriculum_config: {
          grade_level: 'university',
          grade: '大一',
          subjects: ['高等数学'],
          custom_textbooks: ['教材A'],
        },
      }
      expect(configOnly.curriculum_config?.grade_level).toBe('university')

      // 同时更新多个字段
      const fullUpdate: UpdateClassV11Params = {
        name: '新名称',
        description: '新描述',
        is_public: false,
        curriculum_config: {
          grade_level: 'junior',
          subjects: ['语文', '数学'],
        },
      }
      expect(fullUpdate).toBeDefined()
    })
  })

  describe('getClassDetail - 获取班级详情', () => {
    it('应正确导出 getClassDetail 函数', () => {
      expect(typeof getClassDetail).toBe('function')
    })
  })

  describe('其他班级 API', () => {
    it('应正确导出所有班级管理函数', () => {
      expect(typeof createClass).toBe('function')
      expect(typeof updateClass).toBe('function')
      expect(typeof deleteClass).toBe('function')
    })
  })

  describe('CurriculumConfig 类型', () => {
    it('CurriculumConfig 类型应支持所有字段', () => {
      // 完整配置
      const fullConfig: CurriculumConfig = {
        id: 1,
        grade_level: 'primary_lower',
        grade: '三年级',
        subjects: ['数学', '语文'],
        textbook_versions: ['人教版', '北师大版'],
        custom_textbooks: ['教辅A'],
        current_progress: '第三单元',
      }

      expect(fullConfig.id).toBe(1)
      expect(fullConfig.grade_level).toBe('primary_lower')
      expect(fullConfig.subjects).toHaveLength(2)

      // 最小配置
      const minConfig: CurriculumConfig = {
        grade_level: 'junior',
        subjects: ['语文'],
      }

      expect(minConfig.grade_level).toBe('junior')
      expect(minConfig.grade).toBeUndefined()
      expect(minConfig.id).toBeUndefined()
    })
  })
})
