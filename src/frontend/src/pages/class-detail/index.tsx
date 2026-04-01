import { useCallback, useEffect, useState } from 'react'
import { View, Text } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getClassMembers, toggleClass, ClassMember, ClassMemberPageData } from '@/api/class'
import { getRelations, toggleRelation, RelationItemTeacher } from '@/api/relation'
import { usePersonaStore } from '@/store/personaStore'
import './index.scss'

/** 学生卡片数据（合并成员信息与关系状态） */
interface StudentCardData extends ClassMember {
  relation_id?: number
  is_active?: boolean
}

export default function ClassDetail() {
  const params = Taro.getCurrentInstance().router?.params || {}
  const classId = Number(params.class_id) || 0
  const className = decodeURIComponent(params.class_name || '班级')

  const [members, setMembers] = useState<StudentCardData[]>([])
  const [loading, setLoading] = useState(false)
  const [classActive, setClassActive] = useState(true)
  // 启停操作 loading 状态，防止重复点击
  const [togglingClass, setTogglingClass] = useState(false)
  const [togglingStudentId, setTogglingStudentId] = useState<number | null>(null)

  // 获取当前教师分身 ID
  const currentPersona = usePersonaStore((s) => s.currentPersona)
  const teacherPersonaId = currentPersona?.id

  /** 获取班级成员列表 + 关系状态 */
  const fetchMembers = useCallback(async () => {
    if (!classId) return
    setLoading(true)
    try {
      // 并行获取成员列表和关系列表
      const [membersRes, relationsRes] = await Promise.all([
        getClassMembers(classId),
        getRelations({ page: 1, page_size: 200 }),
      ])

      // 解析成员
      const data = membersRes.data
      let memberList: ClassMember[] = []
      if (Array.isArray(data)) {
        memberList = data
      } else {
        memberList = (data as ClassMemberPageData).items || []
      }

      // 构建 relation 映射（student_persona_id -> relation）
      const relations = (relationsRes.data.items || []) as RelationItemTeacher[]
      const relationMap = new Map<number, RelationItemTeacher>()
      relations.forEach((r) => {
        if (r.student_persona_id) {
          relationMap.set(r.student_persona_id, r)
        }
        // 也用 student_id 做映射兜底
        relationMap.set(r.student_id, r)
      })

      // 合并数据
      const merged: StudentCardData[] = memberList.map((m) => {
        const rel = relationMap.get(m.student_persona_id)
        return {
          ...m,
          relation_id: rel?.id,
          is_active: rel?.is_active ?? true,
        }
      })

      setMembers(merged)
    } catch (error) {
      console.error('获取班级成员失败:', error)
    } finally {
      setLoading(false)
    }
  }, [classId])

  useEffect(() => {
    // 设置页面标题
    Taro.setNavigationBarTitle({ title: className })
    // 从路由参数中获取班级激活状态
    const isActiveParam = params.is_active
    if (isActiveParam !== undefined) {
      setClassActive(isActiveParam !== '0' && isActiveParam !== 'false')
    }
    fetchMembers()
  }, [className, fetchMembers])

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    await fetchMembers()
    Taro.stopPullDownRefresh()
  })

  /** 查看学生详情 */
  const handleViewStudent = (member: StudentCardData) => {
    Taro.navigateTo({
      url: `/pages/student-detail/index?student_id=${member.student_persona_id}&student_name=${encodeURIComponent(member.student_nickname)}`,
    })
  }

  /** 查看对话记录 */
  const handleViewChat = (member: StudentCardData) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_persona_id=${teacherPersonaId}&student_persona_id=${member.student_persona_id}`,
    })
  }

  /** 写评语 */
  const handleWriteComment = (member: StudentCardData) => {
    Taro.navigateTo({
      url: `/pages/student-detail/index?student_id=${member.student_persona_id}&student_name=${encodeURIComponent(member.student_nickname)}&tab=comment`,
    })
  }

  /** 设置风格 */
  const handleSetStyle = (member: StudentCardData) => {
    Taro.navigateTo({
      url: `/pages/student-detail/index?student_id=${member.student_persona_id}&student_name=${encodeURIComponent(member.student_nickname)}&tab=style`,
    })
  }

  /** 执行学生启停 API 调用 */
  const doToggleStudent = async (member: StudentCardData, newActive: boolean) => {
    setTogglingStudentId(member.id)
    try {
      await toggleRelation(member.relation_id!, newActive)
      Taro.showToast({ title: newActive ? '已开启' : '已关闭', icon: 'success' })
      setMembers((prev) =>
        prev.map((m) =>
          m.id === member.id ? { ...m, is_active: newActive } : m,
        ),
      )
    } catch (error) {
      console.error('启停学生失败:', error)
    } finally {
      setTogglingStudentId(null)
    }
  }

  /** 启停学生 */
  const handleToggleStudent = (member: StudentCardData) => {
    if (!member.relation_id) {
      Taro.showToast({ title: '无法操作，缺少关系信息', icon: 'none' })
      return
    }
    // 防止重复点击
    if (togglingStudentId !== null) return

    const newActive = !member.is_active

    // 开启操作直接调用，无需二次确认
    if (newActive) {
      doToggleStudent(member, true)
      return
    }

    // 关闭操作需要二次确认
    Taro.showModal({
      title: '关闭确认',
      content: `确认关闭学生"${member.student_nickname}"的访问权限？\n关闭后，该学生将无法与你发起新对话。\n已有的对话记录和数据不会被删除。`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          doToggleStudent(member, false)
        }
      },
    })
  }

  /** 执行班级启停 API 调用 */
  const doToggleClass = async (newActive: boolean) => {
    setTogglingClass(true)
    try {
      await toggleClass(classId, newActive)
      Taro.showToast({ title: newActive ? '班级已开启' : '班级已关闭', icon: 'success' })
      setClassActive(newActive)
    } catch (error) {
      console.error('启停班级失败:', error)
    } finally {
      setTogglingClass(false)
    }
  }

  /** 启停班级 */
  const handleToggleClass = () => {
    // 防止重复点击
    if (togglingClass) return

    const newActive = !classActive

    // 开启操作直接调用，无需二次确认
    if (newActive) {
      doToggleClass(true)
      return
    }

    // 关闭操作需要二次确认
    Taro.showModal({
      title: '关闭确认',
      content: `确认关闭班级"${className}"？\n关闭后，该班级下 ${members.length} 名学生将无法发起新对话。\n已有的对话记录和数据不会被删除。`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          doToggleClass(false)
        }
      },
    })
  }

  /** 编辑班级 */
  const handleEditClass = () => {
    Taro.showToast({ title: '班级编辑功能开发中', icon: 'none' })
  }

  return (
    <View className='class-detail-page'>
      {/* 班级信息头部 */}
      <View className='class-detail-page__header'>
        <View className='class-detail-page__header-info'>
          <Text className='class-detail-page__name'>{className}</Text>
          <Text className='class-detail-page__member-count'>{members.length} 名学生</Text>
        </View>
        <View className='class-detail-page__header-actions'>
          <View
            className={`class-detail-page__toggle-btn ${classActive ? 'class-detail-page__toggle-btn--active' : 'class-detail-page__toggle-btn--inactive'} ${togglingClass ? 'class-detail-page__toggle-btn--loading' : ''}`}
            onClick={handleToggleClass}
          >
            <Text className={`class-detail-page__toggle-btn-text ${classActive ? 'class-detail-page__toggle-btn-text--active' : 'class-detail-page__toggle-btn-text--inactive'}`}>
              {togglingClass ? '操作中...' : (classActive ? '已启用' : '已停用')}
            </Text>
          </View>
          <View className='class-detail-page__edit-btn' onClick={handleEditClass}>
            <Text className='class-detail-page__edit-btn-text'>编辑</Text>
          </View>
        </View>
      </View>

      {/* 学生列表 */}
      <View className='class-detail-page__section'>
        <Text className='class-detail-page__section-title'>学生列表</Text>

        {loading ? (
          <View className='class-detail-page__loading'>
            <Text className='class-detail-page__loading-text'>加载中...</Text>
          </View>
        ) : members.length > 0 ? (
          <View className='class-detail-page__student-list'>
            {members.map((member) => (
              <View
                key={member.id}
                className='class-detail-page__student-item'
              >
                <View
                  className='class-detail-page__student-info'
                  onClick={() => handleViewStudent(member)}
                >
                  <View className='class-detail-page__student-avatar'>
                    <Text className='class-detail-page__student-avatar-text'>
                      {member.student_nickname.charAt(0)}
                    </Text>
                  </View>
                  <View className='class-detail-page__student-detail'>
                    <Text className='class-detail-page__student-name'>
                      {member.student_nickname}
                    </Text>
                    <Text className='class-detail-page__student-time'>
                      加入时间：{member.joined_at?.split('T')[0] || '-'}
                    </Text>
                  </View>
                </View>
                {/* 快捷操作按钮 */}
                <View className='class-detail-page__quick-actions'>
                  <View className='class-detail-page__quick-btn' onClick={() => handleViewChat(member)}>
                    <Text className='class-detail-page__quick-btn-text'>对话记录</Text>
                  </View>
                  <View className='class-detail-page__quick-btn' onClick={() => handleWriteComment(member)}>
                    <Text className='class-detail-page__quick-btn-text'>写评语</Text>
                  </View>
                  <View className='class-detail-page__quick-btn' onClick={() => handleSetStyle(member)}>
                    <Text className='class-detail-page__quick-btn-text'>设置风格</Text>
                  </View>
                </View>
                <View className='class-detail-page__student-actions'>
                  <View
                    className={`class-detail-page__toggle-btn ${member.is_active ? 'class-detail-page__toggle-btn--active' : 'class-detail-page__toggle-btn--inactive'} ${togglingStudentId === member.id ? 'class-detail-page__toggle-btn--loading' : ''}`}
                    onClick={() => handleToggleStudent(member)}
                  >
                    <Text className={`class-detail-page__toggle-btn-text ${member.is_active ? 'class-detail-page__toggle-btn-text--active' : 'class-detail-page__toggle-btn-text--inactive'}`}>
                      {togglingStudentId === member.id ? '操作中...' : (member.is_active ? '已启用' : '已停用')}
                    </Text>
                  </View>
                  <Text
                    className='class-detail-page__student-arrow'
                    onClick={() => handleViewStudent(member)}
                  >›</Text>
                </View>
              </View>
            ))}
          </View>
        ) : (
          <View className='class-detail-page__empty'>
            <Text className='class-detail-page__empty-text'>暂无学生</Text>
          </View>
        )}
      </View>
    </View>
  )
}
