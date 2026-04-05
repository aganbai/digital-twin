import { useState, useCallback } from 'react'
import { View, Text, ScrollView, Input } from '@tarojs/components'
import Taro, { useDidShow, usePullDownRefresh } from '@tarojs/taro'
import { useUserStore } from '@/store'
import {
  getStudentChatList,
  getTeacherChatList,
  createNewSession,
  getQuickActions,
  pinChat,
  unpinChat,
} from '@/api/chat-list'
import type {
  StudentTeacherChatItem,
  TeacherChatClassItem,
  TeacherChatStudent,
  QuickActionItem,
} from '@/api/chat-list'
import { getSessionsV9, SessionInfo } from '@/api/session'
import { formatTime, truncateText } from '@/utils/format'
import { ROLES } from '@/utils/constants'
import Empty from '@/components/Empty'
import './index.scss'

/** 默认每个班级显示的学生数量 */
const DEFAULT_SHOW_COUNT = 5
/** 最大置顶数量 */
const MAX_PIN_COUNT = 10

export default function ChatList() {
  const { userInfo } = useUserStore()
  const isTeacher = userInfo?.role === ROLES.TEACHER

  // ========== 通用状态 ==========
  const [loading, setLoading] = useState(false)

  // ========== 学生端状态 ==========
  const [teachers, setTeachers] = useState<StudentTeacherChatItem[]>([])
  /** 每个老师的历史会话列表（迭代9新增） */
  const [teacherSessions, setTeacherSessions] = useState<Record<number, SessionInfo[]>>({})
  /** 每个老师的历史会话展开状态 */
  const [expandedTeachers, setExpandedTeachers] = useState<Record<number, boolean>>({})
  /** 新会话弹层 */
  const [showNewSession, setShowNewSession] = useState(false)
  /** 新会话选中的老师 */
  const [selectedTeacher, setSelectedTeacher] = useState<StudentTeacherChatItem | null>(null)
  /** 快捷指令列表 */
  const [quickActions, setQuickActions] = useState<QuickActionItem[]>([])
  /** 新会话创建中 */
  const [creatingSession, setCreatingSession] = useState(false)

  // ========== 教师端状态 ==========
  const [classes, setClasses] = useState<TeacherChatClassItem[]>([])
  /** 每个班级的展开状态（key 为 class_id，value 为是否展开） */
  const [expandedClasses, setExpandedClasses] = useState<Record<number, boolean>>({})

  /** 加载学生端聊天列表 */
  const fetchStudentList = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getStudentChatList()
      // 置顶的排在前面
      const sorted = (res.data.teachers || []).sort((a, b) => {
        if (a.is_pinned && !b.is_pinned) return -1
        if (!a.is_pinned && b.is_pinned) return 1
        return 0
      })
      setTeachers(sorted)
    } catch (error) {
      console.error('获取学生聊天列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 加载教师端聊天列表 */
  const fetchTeacherList = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getTeacherChatList()
      // 置顶的班级排在前面
      const sorted = (res.data.classes || []).sort((a, b) => {
        if (a.is_pinned && !b.is_pinned) return -1
        if (!a.is_pinned && b.is_pinned) return 1
        return 0
      })
      // 每个班级内，置顶学生排前面
      sorted.forEach((cls) => {
        if (cls.students) {
          cls.students.sort((a, b) => {
            if (a.is_pinned && !b.is_pinned) return -1
            if (!a.is_pinned && b.is_pinned) return 1
            return 0
          })
        }
      })
      setClasses(sorted)
    } catch (error) {
      console.error('获取教师聊天列表失败:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  /** 页面显示时加载数据 */
  useDidShow(() => {
    if (isTeacher) {
      fetchTeacherList()
    } else {
      fetchStudentList()
    }
  })

  /** 下拉刷新 */
  usePullDownRefresh(async () => {
    if (isTeacher) {
      await fetchTeacherList()
    } else {
      await fetchStudentList()
    }
    Taro.stopPullDownRefresh()
  })

  // ========== 学生端操作 ==========

  /** 点击老师进入聊天详情 */
  const handleTeacherClick = (teacher: StudentTeacherChatItem) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${teacher.teacher_persona_id}&teacher_name=${encodeURIComponent(teacher.teacher_nickname)}`,
    })
  }

  /** 获取老师的历史会话列表（迭代9新增） */
  const fetchTeacherHistorySessions = useCallback(async (teacherPersonaId: number) => {
    try {
      const res = await getSessionsV9(teacherPersonaId, 1, 20)
      setTeacherSessions((prev) => ({
        ...prev,
        [teacherPersonaId]: res.data.items || [],
      }))
    } catch (error) {
      console.error('获取历史会话失败:', error)
    }
  }, [])

  /** 切换老师历史会话展开/收起（迭代9新增） */
  const toggleTeacherExpand = useCallback((teacherPersonaId: number) => {
    setExpandedTeachers((prev) => {
      const newExpanded = !prev[teacherPersonaId]
      // 如果是展开，同时加载历史会话
      if (newExpanded && !teacherSessions[teacherPersonaId]) {
        fetchTeacherHistorySessions(teacherPersonaId)
      }
      return {
        ...prev,
        [teacherPersonaId]: newExpanded,
      }
    })
  }, [teacherSessions, fetchTeacherHistorySessions])

  /** 点击历史会话进入聊天详情（迭代9新增） */
  const handleSessionClick = (session: SessionInfo, teacher: StudentTeacherChatItem) => {
    Taro.navigateTo({
      url: `/pages/chat/index?teacher_id=${teacher.teacher_persona_id}&teacher_name=${encodeURIComponent(teacher.teacher_nickname)}&session_id=${session.session_id}`,
    })
  }

  /** 长按老师 → 置顶/取消置顶 */
  const handleTeacherLongPress = (teacher: StudentTeacherChatItem) => {
    const actions = teacher.is_pinned
      ? ['取消置顶', '取消']
      : ['置顶该老师', '取消']

    Taro.showActionSheet({
      itemList: actions.slice(0, -1),
      success: async (res) => {
        if (res.tapIndex === 0) {
          try {
            if (teacher.is_pinned) {
              await unpinChat('teacher', teacher.teacher_persona_id)
              Taro.showToast({ title: '已取消置顶', icon: 'success' })
            } else {
              // 检查置顶数量限制
              const pinnedCount = teachers.filter((t) => t.is_pinned).length
              if (pinnedCount >= MAX_PIN_COUNT) {
                Taro.showToast({ title: `最多置顶${MAX_PIN_COUNT}个`, icon: 'none' })
                return
              }
              await pinChat({ target_type: 'teacher', target_id: teacher.teacher_persona_id })
              Taro.showToast({ title: '已置顶', icon: 'success' })
            }
            fetchStudentList()
          } catch (error) {
            console.error('置顶操作失败:', error)
          }
        }
      },
    })
  }

  /** 打开新会话弹层 */
  const handleOpenNewSession = (teacher: StudentTeacherChatItem) => {
    setSelectedTeacher(teacher)
    setShowNewSession(true)
    // 加载快捷指令
    getQuickActions(teacher.teacher_persona_id)
      .then((res) => {
        setQuickActions(res.data as unknown as QuickActionItem[] || [])
      })
      .catch(() => {
        setQuickActions([])
      })
  }

  /** 创建新会话 */
  const handleCreateNewSession = async (initialMessage?: string) => {
    if (!selectedTeacher || creatingSession) return
    setCreatingSession(true)
    try {
      const res = await createNewSession({
        teacher_persona_id: selectedTeacher.teacher_persona_id,
        initial_message: initialMessage,
      })
      setShowNewSession(false)
      // 跳转到聊天页面，带上新的 session_id
      Taro.navigateTo({
        url: `/pages/chat/index?teacher_id=${selectedTeacher.teacher_persona_id}&teacher_name=${encodeURIComponent(selectedTeacher.teacher_nickname)}&session_id=${res.data.session_id}`,
      })
    } catch (error) {
      console.error('创建新会话失败:', error)
      Taro.showToast({ title: '创建失败，请重试', icon: 'none' })
    } finally {
      setCreatingSession(false)
    }
  }

  // ========== 教师端操作 ==========

  /** 切换班级展开/收起 */
  const toggleClassExpand = (classId: number) => {
    setExpandedClasses((prev) => ({
      ...prev,
      [classId]: !prev[classId],
    }))
  }

  /** 点击学生进入聊天详情 */
  const handleStudentClick = (student: TeacherChatStudent) => {
    Taro.navigateTo({
      url: `/pages/student-chat-history/index?student_persona_id=${student.student_persona_id}&student_name=${encodeURIComponent(student.student_nickname)}`,
    })
  }

  /** 长按班级 → 置顶/取消置顶 */
  const handleClassLongPress = (cls: TeacherChatClassItem) => {
    const actions = cls.is_pinned
      ? ['取消置顶']
      : ['置顶该班级']

    Taro.showActionSheet({
      itemList: actions,
      success: async (res) => {
        if (res.tapIndex === 0) {
          try {
            if (cls.is_pinned) {
              await unpinChat('class', cls.class_id)
              Taro.showToast({ title: '已取消置顶', icon: 'success' })
            } else {
              const pinnedCount = classes.filter((c) => c.is_pinned).length
              if (pinnedCount >= MAX_PIN_COUNT) {
                Taro.showToast({ title: `最多置顶${MAX_PIN_COUNT}个`, icon: 'none' })
                return
              }
              await pinChat({ target_type: 'class', target_id: cls.class_id })
              Taro.showToast({ title: '已置顶', icon: 'success' })
            }
            fetchTeacherList()
          } catch (error) {
            console.error('置顶操作失败:', error)
          }
        }
      },
    })
  }

  /** 长按学生 → 置顶/取消置顶 */
  const handleStudentLongPress = (student: TeacherChatStudent) => {
    const actions = student.is_pinned
      ? ['取消置顶']
      : ['置顶该学生']

    Taro.showActionSheet({
      itemList: actions,
      success: async (res) => {
        if (res.tapIndex === 0) {
          try {
            if (student.is_pinned) {
              await unpinChat('student', student.student_persona_id)
              Taro.showToast({ title: '已取消置顶', icon: 'success' })
            } else {
              const allPinnedStudents = classes.reduce(
                (count, cls) => count + (cls.students?.filter((s) => s.is_pinned).length || 0),
                0,
              )
              if (allPinnedStudents >= MAX_PIN_COUNT) {
                Taro.showToast({ title: `最多置顶${MAX_PIN_COUNT}个学生`, icon: 'none' })
                return
              }
              await pinChat({ target_type: 'student', target_id: student.student_persona_id })
              Taro.showToast({ title: '已置顶', icon: 'success' })
            }
            fetchTeacherList()
          } catch (error) {
            console.error('置顶操作失败:', error)
          }
        }
      },
    })
  }

  // ========== 渲染：学生端 ==========
  const renderStudentView = () => (
    <View className='chat-list__body'>
      {loading && teachers.length === 0 ? (
        <View className='chat-list__loading'>
          <Text className='chat-list__loading-text'>加载中...</Text>
        </View>
      ) : teachers.length > 0 ? (
        <ScrollView className='chat-list__scroll' scrollY enhanced showScrollbar={false}>
          {teachers.map((teacher) => {
            const isExpanded = expandedTeachers[teacher.teacher_persona_id]
            const sessions = teacherSessions[teacher.teacher_persona_id] || []
            
            return (
              <View key={teacher.teacher_persona_id} className='chat-list__teacher-group'>
                {/* 老师卡片（最新会话） */}
                <View
                  className={`chat-list__item ${teacher.is_pinned ? 'chat-list__item--pinned' : ''}`}
                  onClick={() => handleTeacherClick(teacher)}
                  onLongPress={() => handleTeacherLongPress(teacher)}
                >
                  {/* 头像 */}
                  <View className='chat-list__avatar'>
                    <Text className='chat-list__avatar-text'>
                      {teacher.teacher_nickname.charAt(0).toUpperCase()}
                    </Text>
                  </View>

                  {/* 信息区 */}
                  <View className='chat-list__info'>
                    <View className='chat-list__info-top'>
                      <View className='chat-list__name-row'>
                        {teacher.is_pinned && (
                          <Text className='chat-list__pin-icon'>📌</Text>
                        )}
                        <Text className='chat-list__name'>{teacher.teacher_nickname}</Text>
                        {teacher.subject && (
                          <Text className='chat-list__subject'>{teacher.subject}</Text>
                        )}
                      </View>
                      {teacher.last_message_time && (
                        <Text className='chat-list__time'>
                          {formatTime(teacher.last_message_time)}
                        </Text>
                      )}
                    </View>
                    <View className='chat-list__info-bottom'>
                      <Text className='chat-list__message'>
                        {teacher.last_message
                          ? truncateText(teacher.last_message, 30)
                          : '暂无消息'}
                      </Text>
                      {teacher.unread_count > 0 && (
                        <View className='chat-list__badge'>
                          <Text className='chat-list__badge-text'>
                            {teacher.unread_count > 99 ? '99+' : teacher.unread_count}
                          </Text>
                        </View>
                      )}
                    </View>
                  </View>
                </View>

                {/* 历史会话展开按钮 */}
                <View
                  className='chat-list__expand-sessions'
                  onClick={() => toggleTeacherExpand(teacher.teacher_persona_id)}
                >
                  <Text className='chat-list__expand-text'>
                    {isExpanded ? '收起历史会话' : '展开历史会话'}
                  </Text>
                  <Text className={`chat-list__expand-arrow ${isExpanded ? 'chat-list__expand-arrow--up' : ''}`}>
                    ▼
                  </Text>
                </View>

                {/* 历史会话列表 */}
                {isExpanded && sessions.length > 0 && (
                  <View className='chat-list__sessions'>
                    {sessions.map((session) => (
                      <View
                        key={session.session_id}
                        className='chat-list__session-item'
                        onClick={() => handleSessionClick(session, teacher)}
                      >
                        <View className='chat-list__session-icon'>💬</View>
                        <View className='chat-list__session-info'>
                          <Text className='chat-list__session-title'>
                            {session.title || truncateText(session.last_message, 25)}
                          </Text>
                          <Text className='chat-list__session-meta'>
                            {session.message_count} 条消息 · {formatTime(session.updated_at)}
                          </Text>
                        </View>
                      </View>
                    ))}
                  </View>
                )}

                {isExpanded && sessions.length === 0 && (
                  <View className='chat-list__sessions-empty'>
                    <Text>暂无历史会话</Text>
                  </View>
                )}
              </View>
            )
          })}
        </ScrollView>
      ) : (
        <View className='chat-list__empty'>
          <Empty text='暂无聊天记录' />
          <Text className='chat-list__empty-hint'>去发现页面找一位老师开始对话吧</Text>
        </View>
      )}

      {/* 底部输入栏（仿微信） */}
      {teachers.length > 0 && (
        <View className='chat-list__bottom-bar safe-area-bottom'>
          <View className='chat-list__bottom-wrapper'>
            {/* 语音按钮（占位） */}
            <View className='chat-list__bar-btn'>
              <Text className='chat-list__bar-btn-text'>🎤</Text>
            </View>
            {/* 输入框（点击后选择老师进入聊天） */}
            <View
              className='chat-list__bar-input'
              onClick={() => {
                if (teachers.length === 1) {
                  handleTeacherClick(teachers[0])
                } else {
                  Taro.showActionSheet({
                    itemList: teachers.map((t) => t.teacher_nickname),
                    success: (res) => {
                      handleTeacherClick(teachers[res.tapIndex])
                    },
                  })
                }
              }}
            >
              <Text className='chat-list__bar-input-placeholder'>选择老师开始对话...</Text>
            </View>
            {/* 表情按钮 */}
            <View className='chat-list__bar-btn'>
              <Text className='chat-list__bar-btn-text'>😊</Text>
            </View>
            {/* ➕ 新会话按钮 */}
            <View
              className='chat-list__bar-btn chat-list__bar-btn--plus'
              onClick={() => {
                if (teachers.length === 1) {
                  handleOpenNewSession(teachers[0])
                } else {
                  Taro.showActionSheet({
                    itemList: teachers.map((t) => `${t.teacher_nickname}${t.subject ? ` (${t.subject})` : ''}`),
                    success: (res) => {
                      handleOpenNewSession(teachers[res.tapIndex])
                    },
                  })
                }
              }}
            >
              <Text className='chat-list__bar-btn-text'>➕</Text>
            </View>
          </View>
        </View>
      )}

      {/* 新会话弹层 */}
      {showNewSession && selectedTeacher && (
        <View className='chat-list__modal-mask' onClick={() => setShowNewSession(false)}>
          <View className='chat-list__modal' onClick={(e) => e.stopPropagation()}>
            <View className='chat-list__modal-header'>
              <Text className='chat-list__modal-title'>
                开启与 {selectedTeacher.teacher_nickname} 的新会话
              </Text>
              <View className='chat-list__modal-close' onClick={() => setShowNewSession(false)}>
                <Text className='chat-list__modal-close-text'>✕</Text>
              </View>
            </View>

            {/* 隔离线提示 */}
            <View className='chat-list__modal-divider'>
              <View className='chat-list__modal-divider-line' />
              <Text className='chat-list__modal-divider-text'>新的对话将从这里开始</Text>
              <View className='chat-list__modal-divider-line' />
            </View>

            {/* 快捷指令 */}
            <View className='chat-list__quick-actions'>
              <Text className='chat-list__quick-actions-title'>快捷开始：</Text>
              {quickActions.length > 0 ? (
                quickActions.map((action) => (
                  <View
                    key={action.id}
                    className='chat-list__quick-action-item'
                    onClick={() => handleCreateNewSession(action.action)}
                  >
                    <Text className='chat-list__quick-action-text'>{action.label}</Text>
                  </View>
                ))
              ) : (
                <Text className='chat-list__quick-actions-empty'>加载中...</Text>
              )}
            </View>

            {/* 空白开始 */}
            <View
              className='chat-list__modal-start-btn'
              onClick={() => handleCreateNewSession()}
            >
              <Text className='chat-list__modal-start-text'>
                {creatingSession ? '创建中...' : '💬 直接开始对话'}
              </Text>
            </View>
          </View>
        </View>
      )}
    </View>
  )

  // ========== 渲染：教师端 ==========
  const renderTeacherView = () => (
    <View className='chat-list__body'>
      {loading && classes.length === 0 ? (
        <View className='chat-list__loading'>
          <Text className='chat-list__loading-text'>加载中...</Text>
        </View>
      ) : classes.length > 0 ? (
        <ScrollView className='chat-list__scroll' scrollY enhanced showScrollbar={false}>
          {classes.map((cls) => {
            const isExpanded = expandedClasses[cls.class_id]
            const displayStudents = isExpanded
              ? cls.students || []
              : (cls.students || []).slice(0, DEFAULT_SHOW_COUNT)
            const hasMore = (cls.students || []).length > DEFAULT_SHOW_COUNT

            return (
              <View
                key={cls.class_id}
                className={`chat-list__class ${cls.is_pinned ? 'chat-list__class--pinned' : ''}`}
              >
                {/* 班级头部 */}
                <View
                  className='chat-list__class-header'
                  onLongPress={() => handleClassLongPress(cls)}
                >
                  <View className='chat-list__class-info'>
                    {cls.is_pinned && (
                      <Text className='chat-list__pin-icon'>📌</Text>
                    )}
                    <Text className='chat-list__class-name'>{cls.class_name}</Text>
                    {cls.subject && (
                      <Text className='chat-list__class-subject'>{cls.subject}</Text>
                    )}
                  </View>
                  <Text className='chat-list__class-count'>
                    {(cls.students || []).length} 名学生
                  </Text>
                </View>

                {/* 学生列表 */}
                <View className='chat-list__students'>
                  {displayStudents.map((student) => (
                    <View
                      key={student.student_persona_id}
                      className={`chat-list__student ${student.is_pinned ? 'chat-list__student--pinned' : ''}`}
                      onClick={() => handleStudentClick(student)}
                      onLongPress={() => handleStudentLongPress(student)}
                    >
                      {/* 学生头像 */}
                      <View className='chat-list__student-avatar'>
                        <Text className='chat-list__student-avatar-text'>
                          {student.student_nickname.charAt(0).toUpperCase()}
                        </Text>
                      </View>

                      {/* 学生信息 */}
                      <View className='chat-list__student-info'>
                        <View className='chat-list__student-top'>
                          <View className='chat-list__student-name-row'>
                            {student.is_pinned && (
                              <Text className='chat-list__pin-icon chat-list__pin-icon--small'>📌</Text>
                            )}
                            <Text className='chat-list__student-name'>
                              {student.student_nickname}
                            </Text>
                          </View>
                          {student.last_message_time && (
                            <Text className='chat-list__student-time'>
                              {formatTime(student.last_message_time)}
                            </Text>
                          )}
                        </View>
                        <View className='chat-list__student-bottom'>
                          <Text className='chat-list__student-message'>
                            {student.last_message
                              ? truncateText(student.last_message, 25)
                              : '暂无消息'}
                          </Text>
                          {student.unread_count > 0 && (
                            <View className='chat-list__badge'>
                              <Text className='chat-list__badge-text'>
                                {student.unread_count > 99 ? '99+' : student.unread_count}
                              </Text>
                            </View>
                          )}
                        </View>
                      </View>
                    </View>
                  ))}

                  {/* 查看更多 / 收起 */}
                  {hasMore && (
                    <View
                      className='chat-list__expand-btn'
                      onClick={() => toggleClassExpand(cls.class_id)}
                    >
                      <Text className='chat-list__expand-text'>
                        {isExpanded
                          ? '收起'
                          : `查看更多 (${(cls.students || []).length - DEFAULT_SHOW_COUNT})`}
                      </Text>
                      <Text className='chat-list__expand-arrow'>
                        {isExpanded ? '▲' : '▼'}
                      </Text>
                    </View>
                  )}
                </View>
              </View>
            )
          })}
        </ScrollView>
      ) : (
        <View className='chat-list__empty'>
          <Empty text='暂无班级聊天记录' />
          <Text className='chat-list__empty-hint'>创建班级并添加学生后即可查看</Text>
        </View>
      )}
    </View>
  )

  return (
    <View className='chat-list-page'>
      {/* 页面标题 */}
      <View className='chat-list__header'>
        <Text className='chat-list__title'>
          {isTeacher ? '💬 学生消息' : '💬 聊天列表'}
        </Text>
        <Text className='chat-list__subtitle'>
          {isTeacher
            ? `共 ${classes.length} 个班级`
            : `共 ${teachers.length} 位老师`}
        </Text>
      </View>

      {/* 根据角色渲染不同内容 */}
      {isTeacher ? renderTeacherView() : renderStudentView()}
    </View>
  )
}
