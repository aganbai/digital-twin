import { useState, useCallback, useEffect } from 'react'
import { View, Text, Input } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { getRelations, approveRelation, rejectRelation, inviteStudent, toggleRelation, RelationItemTeacher } from '@/api/relation'
import { getClasses, deleteClass, getClassMembers, removeClassMember, toggleClass } from '@/api/class'
import type { ClassInfo, ClassMember } from '@/api/class'
import { usePersonaStore } from '@/store'
import Empty from '@/components/Empty'
import './index.scss'

/** Tab 类型 - R4: 合并为4个Tab */
type TabType = 'all' | 'byClass' | 'pending' | 'classSettings'

export default function TeacherStudents() {
  const [activeTab, setActiveTab] = useState<TabType>('all')
  const { currentPersona } = usePersonaStore()

  // 学生管理
  const [relations, setRelations] = useState<RelationItemTeacher[]>([])
  const [loading, setLoading] = useState(false)
  const [showInviteModal, setShowInviteModal] = useState(false)
  const [inviteId, setInviteId] = useState('')

  // 班级管理
  const [classes, setClasses] = useState<ClassInfo[]>([])
  const [classLoading, setClassLoading] = useState(false)
  const [expandedClassId, setExpandedClassId] = useState<number | null>(null)
  const [classMembers, setClassMembers] = useState<Record<number, ClassMember[]>>({})

  // 启停操作 loading 状态，防止重复点击
  const [togglingRelationId, setTogglingRelationId] = useState<number | null>(null)
  const [togglingClassId, setTogglingClassId] = useState<number | null>(null)

  /** 获取关系列表 */
  const fetchRelations = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getRelations({ page: 1, page_size: 100 })
      setRelations((res.data.items || []) as RelationItemTeacher[])
    } catch (error) {
      console.error('获取学生列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 获取班级列表 */
  const fetchClasses = useCallback(async () => {
    setClassLoading(true)
    try {
      const res = await getClasses()
      setClasses(res.data || [])
    } catch (error) {
      console.error('获取班级列表失败:', error)
    } finally {
      setClassLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchRelations()
    fetchClasses()
  }, [fetchRelations, fetchClasses])

  usePullDownRefresh(async () => {
    await Promise.all([fetchRelations(), fetchClasses()])
    Taro.stopPullDownRefresh()
  })

  /** 待审批列表 */
  const pendingList = relations.filter((r) => r.status === 'pending')
  /** 已授权列表 */
  const approvedList = relations.filter((r) => r.status === 'approved')

  /** 同意申请 */
  const handleApprove = async (id: number) => {
    try {
      await approveRelation(id)
      Taro.showToast({ title: '已同意', icon: 'success' })
      setRelations((prev) =>
        prev.map((r) => (r.id === id ? { ...r, status: 'approved' as const } : r)),
      )
    } catch (error) {
      console.error('审批失败:', error)
    }
  }

  /** 拒绝申请 */
  const handleReject = async (id: number) => {
    try {
      await rejectRelation(id)
      Taro.showToast({ title: '已拒绝', icon: 'success' })
      setRelations((prev) => prev.filter((r) => r.id !== id))
    } catch (error) {
      console.error('拒绝失败:', error)
    }
  }

  /** 查看学生详情 */
  const handleViewDetail = (studentId: number, studentName: string, studentPersonaId?: number) => {
    Taro.navigateTo({
      url: `/pages/student-detail/index?student_id=${studentId}&student_name=${encodeURIComponent(studentName)}${studentPersonaId ? '&student_persona_id=' + studentPersonaId : ''}`,
    })
  }

  /** 查看学生对话记录 */
  const handleViewChatHistory = (studentPersonaId: number, studentName: string) => {
    if (!studentPersonaId) {
      Taro.showToast({ title: '暂无分身信息', icon: 'none' })
      return
    }
    Taro.navigateTo({
      url: `/pages/student-chat-history/index?student_persona_id=${studentPersonaId}&student_name=${encodeURIComponent(studentName)}`,
    })
  }

  /** 邀请学生 */
  const handleInvite = async () => {
    const id = parseInt(inviteId.trim())
    if (!id || isNaN(id)) {
      Taro.showToast({ title: '请输入有效的学生ID', icon: 'none' })
      return
    }
    try {
      await inviteStudent(id)
      Taro.showToast({ title: '邀请成功', icon: 'success' })
      setShowInviteModal(false)
      setInviteId('')
      fetchRelations()
    } catch (error) {
      console.error('邀请失败:', error)
    }
  }

  /** 跳转创建班级 */
  const handleCreateClass = () => {
    Taro.navigateTo({ url: '/pages/class-create/index' })
  }

  /** 展开/收起班级成员 */
  const handleToggleClass = async (classId: number) => {
    if (expandedClassId === classId) {
      setExpandedClassId(null)
      return
    }
    setExpandedClassId(classId)
    if (!classMembers[classId]) {
      try {
        const res = await getClassMembers(classId)
        const members = res.data?.items || res.data || []
        setClassMembers((prev) => ({ ...prev, [classId]: members as ClassMember[] }))
      } catch (error) {
        console.error('获取班级成员失败:', error)
      }
    }
  }

  /** 删除班级 */
  const handleDeleteClass = (cls: ClassInfo) => {
    Taro.showModal({
      title: '删除确认',
      content: `确定要删除班级「${cls.name}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          try {
            await deleteClass(cls.id)
            Taro.showToast({ title: '删除成功', icon: 'success' })
            fetchClasses()
          } catch (error) {
            console.error('删除班级失败:', error)
          }
        }
      },
    })
  }

  /** 移除班级成员 */
  const handleRemoveMember = (classId: number, member: ClassMember) => {
    Taro.showModal({
      title: '移除确认',
      content: `确定要移除「${member.student_nickname}」吗？`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          try {
            await removeClassMember(classId, member.id)
            Taro.showToast({ title: '已移除', icon: 'success' })
            setClassMembers((prev) => ({
              ...prev,
              [classId]: (prev[classId] || []).filter((m) => m.id !== member.id),
            }))
          } catch (error) {
            console.error('移除成员失败:', error)
          }
        }
      },
    })
  }

  /** 执行学生启停 API 调用 */
  const doToggleStudent = async (item: RelationItemTeacher, newActive: boolean) => {
    setTogglingRelationId(item.id)
    try {
      await toggleRelation(item.id, newActive)
      Taro.showToast({ title: newActive ? '已开启' : '已关闭', icon: 'success' })
      setRelations((prev) =>
        prev.map((r) => (r.id === item.id ? { ...r, is_active: newActive } : r)),
      )
    } catch (error) {
      console.error('启停学生失败:', error)
    } finally {
      setTogglingRelationId(null)
    }
  }

  /** 启停学生访问权限 */
  const handleToggleStudent = (item: RelationItemTeacher) => {
    if (togglingRelationId !== null) return
    const newActive = !(item.is_active ?? true)
    if (newActive) {
      doToggleStudent(item, true)
      return
    }
    Taro.showModal({
      title: '关闭确认',
      content: `确认关闭学生"${item.student_nickname}"的访问权限？\n关闭后，该学生将无法与你发起新对话。\n已有的对话记录和数据不会被删除。`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          doToggleStudent(item, false)
        }
      },
    })
  }

  /** 执行班级启停 API 调用 */
  const doToggleClassStatus = async (cls: ClassInfo, newActive: boolean) => {
    setTogglingClassId(cls.id)
    try {
      await toggleClass(cls.id, newActive)
      Taro.showToast({ title: newActive ? '班级已开启' : '班级已关闭', icon: 'success' })
      setClasses((prev) =>
        prev.map((c) => (c.id === cls.id ? { ...c, is_active: newActive } : c)),
      )
    } catch (error) {
      console.error('启停班级失败:', error)
    } finally {
      setTogglingClassId(null)
    }
  }

  /** 启停班级 */
  const handleToggleClassStatus = (cls: ClassInfo) => {
    if (togglingClassId !== null) return
    const newActive = !(cls.is_active ?? true)
    if (newActive) {
      doToggleClassStatus(cls, true)
      return
    }
    Taro.showModal({
      title: '关闭确认',
      content: `确认关闭班级"${cls.name}"？\n关闭后，该班级下所有学生将无法发起新对话。\n已有的对话记录和数据不会被删除。`,
      confirmColor: '#FF4D4F',
      success: async (result) => {
        if (result.confirm) {
          doToggleClassStatus(cls, false)
        }
      },
    })
  }

  /** 渲染学生卡片（复用） */
  const renderStudentCard = (item: RelationItemTeacher) => (
    <View key={item.id} className='teacher-students-page__card'>
      <View className='teacher-students-page__card-info'>
        <Text className='teacher-students-page__card-name'>
          {item.student_nickname}
        </Text>
        <Text className='teacher-students-page__card-desc'>
          {item.status === 'approved' ? '已授权 ✅' : '申请使用'}
        </Text>
      </View>
      <View className='teacher-students-page__card-actions'>
        {item.status === 'approved' && (
          <>
            <View
              className={`teacher-students-page__toggle-btn ${(item.is_active ?? true) ? 'teacher-students-page__toggle-btn--active' : 'teacher-students-page__toggle-btn--inactive'} ${togglingRelationId === item.id ? 'teacher-students-page__toggle-btn--loading' : ''}`}
              onClick={() => handleToggleStudent(item)}
            >
              <Text className={`teacher-students-page__toggle-btn-text ${(item.is_active ?? true) ? 'teacher-students-page__toggle-btn-text--active' : 'teacher-students-page__toggle-btn-text--inactive'}`}>
                {togglingRelationId === item.id ? '操作中...' : ((item.is_active ?? true) ? '已启用' : '已停用')}
              </Text>
            </View>
            {item.student_persona_id && (
              <View
                className='teacher-students-page__link teacher-students-page__link--chat'
                onClick={() => handleViewChatHistory(item.student_persona_id!, item.student_nickname)}
              >
                <Text className='teacher-students-page__link-text--chat'>对话记录</Text>
              </View>
            )}
          </>
        )}
        <View
          className='teacher-students-page__link'
          onClick={() => handleViewDetail(item.student_id, item.student_nickname, item.student_persona_id)}
        >
          <Text className='teacher-students-page__link-text'>详情 →</Text>
        </View>
      </View>
    </View>
  )

  return (
    <View className='teacher-students-page'>
      {/* R4: 4个Tab切换 */}
      <View className='teacher-students-page__tabs'>
        <View
          className={`teacher-students-page__tab ${activeTab === 'all' ? 'teacher-students-page__tab--active' : ''}`}
          onClick={() => setActiveTab('all')}
        >
          <Text className={`teacher-students-page__tab-text ${activeTab === 'all' ? 'teacher-students-page__tab-text--active' : ''}`}>
            全部学生
          </Text>
        </View>
        <View
          className={`teacher-students-page__tab ${activeTab === 'byClass' ? 'teacher-students-page__tab--active' : ''}`}
          onClick={() => setActiveTab('byClass')}
        >
          <Text className={`teacher-students-page__tab-text ${activeTab === 'byClass' ? 'teacher-students-page__tab-text--active' : ''}`}>
            按班级
          </Text>
        </View>
        <View
          className={`teacher-students-page__tab ${activeTab === 'pending' ? 'teacher-students-page__tab--active' : ''}`}
          onClick={() => setActiveTab('pending')}
        >
          <Text className={`teacher-students-page__tab-text ${activeTab === 'pending' ? 'teacher-students-page__tab-text--active' : ''}`}>
            待审批{pendingList.length > 0 ? `(${pendingList.length})` : ''}
          </Text>
        </View>
        <View
          className={`teacher-students-page__tab ${activeTab === 'classSettings' ? 'teacher-students-page__tab--active' : ''}`}
          onClick={() => setActiveTab('classSettings')}
        >
          <Text className={`teacher-students-page__tab-text ${activeTab === 'classSettings' ? 'teacher-students-page__tab-text--active' : ''}`}>
            班级设置
          </Text>
        </View>
      </View>

      {/* Tab 1: 全部学生 */}
      {activeTab === 'all' && (
        <>
          {loading ? (
            <View className='teacher-students-page__loading'>
              <Text>加载中...</Text>
            </View>
          ) : approvedList.length > 0 ? (
            <View className='teacher-students-page__section'>
              <Text className='teacher-students-page__section-title'>
                已授权学生 ({approvedList.length})
              </Text>
              {approvedList.map((item) => renderStudentCard(item))}
            </View>
          ) : (
            <Empty text='暂无学生' />
          )}

          {/* 底部操作 */}
          <View className='teacher-students-page__bottom-actions'>
            <View
              className='teacher-students-page__bottom-btn teacher-students-page__bottom-btn--secondary'
              onClick={() => setShowInviteModal(true)}
            >
              <Text className='teacher-students-page__bottom-btn-text--secondary'>邀请学生</Text>
            </View>
          </View>
        </>
      )}

      {/* Tab 2: 按班级 */}
      {activeTab === 'byClass' && (
        <>
          {classLoading ? (
            <View className='teacher-students-page__loading'>
              <Text>加载中...</Text>
            </View>
          ) : (
            <>
              {classes.map((cls) => (
                <View key={cls.id} className='teacher-students-page__class-card'>
                  <View
                    className='teacher-students-page__class-header'
                    onClick={() => handleToggleClass(cls.id)}
                  >
                    <View className='teacher-students-page__class-info'>
                      <Text className='teacher-students-page__card-name'>{cls.name}</Text>
                      <Text className='teacher-students-page__card-desc'>
                        {cls.member_count} 名学生
                        {cls.description ? ` · ${cls.description}` : ''}
                      </Text>
                    </View>
                    <Text className='teacher-students-page__class-arrow'>
                      {expandedClassId === cls.id ? '▲' : '▼'}
                    </Text>
                  </View>

                  {expandedClassId === cls.id && (
                    <View className='teacher-students-page__class-members'>
                      {(classMembers[cls.id] || []).length > 0 ? (
                        (classMembers[cls.id] || []).map((member) => (
                          <View key={member.id} className='teacher-students-page__member-item'>
                            <Text className='teacher-students-page__member-name'>
                              {member.student_nickname}
                            </Text>
                            <View
                              className='teacher-students-page__member-remove'
                              onClick={() => handleRemoveMember(cls.id, member)}
                            >
                              <Text className='teacher-students-page__member-remove-text'>移除</Text>
                            </View>
                          </View>
                        ))
                      ) : (
                        <Text className='teacher-students-page__empty-text'>暂无成员</Text>
                      )}
                    </View>
                  )}
                </View>
              ))}

              {/* 未分班学生：从已授权学生中排除所有班级成员 */}
              {(() => {
                const allClassMemberIds = new Set(
                  Object.values(classMembers).flat().map((m) => m.student_id),
                )
                const unassignedList = approvedList.filter(
                  (item) => !allClassMemberIds.has(item.student_id),
                )
                return unassignedList.length > 0 ? (
                  <View className='teacher-students-page__section'>
                    <Text className='teacher-students-page__section-title'>未分班学生 ({unassignedList.length})</Text>
                    {unassignedList.map((item) => (
                      <View key={item.id} className='teacher-students-page__card'>
                        <View className='teacher-students-page__card-info'>
                          <Text className='teacher-students-page__card-name'>{item.student_nickname}</Text>
                        </View>
                        <View
                          className='teacher-students-page__link'
                          onClick={() => handleViewDetail(item.student_id, item.student_nickname, item.student_persona_id)}
                        >
                          <Text className='teacher-students-page__link-text'>详情 →</Text>
                        </View>
                      </View>
                    ))}
                  </View>
                ) : null
              })()}
            </>
          )}
        </>
      )}

      {/* Tab 3: 待审批 */}
      {activeTab === 'pending' && (
        <>
          {loading ? (
            <View className='teacher-students-page__loading'>
              <Text>加载中...</Text>
            </View>
          ) : pendingList.length > 0 ? (
            <View className='teacher-students-page__section'>
              {pendingList.map((item) => (
                <View key={item.id} className='teacher-students-page__card'>
                  <View className='teacher-students-page__card-info'>
                    <Text className='teacher-students-page__card-name'>
                      {item.student_nickname}
                    </Text>
                    <Text className='teacher-students-page__card-desc'>申请使用</Text>
                  </View>
                  <View className='teacher-students-page__card-actions'>
                    <View
                      className='teacher-students-page__action-btn teacher-students-page__action-btn--approve'
                      onClick={() => handleApprove(item.id)}
                    >
                      <Text className='teacher-students-page__action-btn-text'>同意</Text>
                    </View>
                    <View
                      className='teacher-students-page__action-btn teacher-students-page__action-btn--reject'
                      onClick={() => handleReject(item.id)}
                    >
                      <Text className='teacher-students-page__action-btn-text--reject'>拒绝</Text>
                    </View>
                  </View>
                </View>
              ))}
            </View>
          ) : (
            <Empty text='暂无待审批申请' />
          )}
        </>
      )}

      {/* Tab 4: 班级设置 */}
      {activeTab === 'classSettings' && (
        <>
          {classLoading ? (
            <View className='teacher-students-page__loading'>
              <Text>加载中...</Text>
            </View>
          ) : classes.length > 0 ? (
            <View className='teacher-students-page__section'>
              {classes.map((cls) => (
                <View key={cls.id} className='teacher-students-page__card'>
                  <View className='teacher-students-page__card-info'>
                    <Text className='teacher-students-page__card-name'>{cls.name}</Text>
                    <Text className='teacher-students-page__card-desc'>
                      {cls.member_count} 名学生
                      {cls.description ? ` · ${cls.description}` : ''}
                    </Text>
                  </View>
                  <View className='teacher-students-page__card-actions'>
                    <View
                      className={`teacher-students-page__toggle-btn ${(cls.is_active ?? true) ? 'teacher-students-page__toggle-btn--active' : 'teacher-students-page__toggle-btn--inactive'} ${togglingClassId === cls.id ? 'teacher-students-page__toggle-btn--loading' : ''}`}
                      onClick={() => handleToggleClassStatus(cls)}
                    >
                      <Text className={`teacher-students-page__toggle-btn-text ${(cls.is_active ?? true) ? 'teacher-students-page__toggle-btn-text--active' : 'teacher-students-page__toggle-btn-text--inactive'}`}>
                        {togglingClassId === cls.id ? '操作中...' : ((cls.is_active ?? true) ? '已启用' : '已停用')}
                      </Text>
                    </View>
                    <View
                      className='teacher-students-page__class-action-btn'
                      onClick={() => handleDeleteClass(cls)}
                    >
                      <Text className='teacher-students-page__class-action-text--danger'>删除</Text>
                    </View>
                  </View>
                </View>
              ))}
            </View>
          ) : (
            <Empty text='暂无班级' />
          )}

          {/* 创建班级按钮 */}
          <View
            className='teacher-students-page__invite-btn'
            onClick={handleCreateClass}
          >
            <Text className='teacher-students-page__invite-btn-text'>+ 创建班级</Text>
          </View>
        </>
      )}

      {/* 邀请学生弹窗 */}
      {showInviteModal && (
        <View className='teacher-students-page__modal-mask' onClick={() => setShowInviteModal(false)}>
          <View className='teacher-students-page__modal' onClick={(e) => e.stopPropagation()}>
            <Text className='teacher-students-page__modal-title'>邀请学生</Text>
            <Input
              className='teacher-students-page__modal-input'
              placeholder='请输入学生ID'
              value={inviteId}
              type='number'
              onInput={(e) => setInviteId(e.detail.value)}
            />
            <View className='teacher-students-page__modal-actions'>
              <View
                className='teacher-students-page__modal-btn teacher-students-page__modal-btn--cancel'
                onClick={() => { setShowInviteModal(false); setInviteId('') }}
              >
                <Text>取消</Text>
              </View>
              <View
                className='teacher-students-page__modal-btn teacher-students-page__modal-btn--confirm'
                onClick={handleInvite}
              >
                <Text className='teacher-students-page__modal-btn-text--confirm'>确认邀请</Text>
              </View>
            </View>
          </View>
        </View>
      )}
    </View>
  )
}
