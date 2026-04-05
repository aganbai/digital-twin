import { useState, useEffect } from 'react'
import Taro from '@tarojs/taro'
import { View, Text, Input, Button } from '@tarojs/components'
import { getClassDetail, ClassDetailForStudent } from '../../api/session'
import { getStudentProfile, updateStudentEvaluation, StudentProfileForTeacher } from '../../api/session'
import { getUserInfo } from '../../utils/storage'
import './index.scss'

/** 用户角色类型 */
type UserRole = 'student' | 'teacher'

interface AvatarPopupProps {
  /** 是否显示 */
  visible: boolean
  /** 关闭回调 */
  onClose: () => void
  /** 用户角色 */
  userRole: UserRole
  /** 目标用户ID（学生点击老师时为teacher_persona_id，老师点击学生时为student_persona_id） */
  targetId?: number
  /** 班级ID（学生查看班级信息时需要） */
  classId?: number
}

/**
 * 头像点击弹窗组件
 * 学生点击老师头像 → 显示班级信息
 * 老师点击学生头像 → 显示学生信息并可修改评语
 */
export default function AvatarPopup({
  visible,
  onClose,
  userRole,
  targetId,
  classId,
}: AvatarPopupProps) {
  const [loading, setLoading] = useState(false)
  const [classInfo, setClassInfo] = useState<ClassDetailForStudent | null>(null)
  const [studentInfo, setStudentInfo] = useState<StudentProfileForTeacher | null>(null)
  const [editMode, setEditMode] = useState(false)
  const [evaluation, setEvaluation] = useState('')

  useEffect(() => {
    if (visible) {
      fetchData()
    }
  }, [visible, userRole, targetId, classId])

  /** 获取数据 */
  const fetchData = async () => {
    setLoading(true)
    try {
      if (userRole === 'student' && classId) {
        // 学生查看班级信息
        const res = await getClassDetail(classId)
        setClassInfo(res.data)
      } else if (userRole === 'teacher' && targetId) {
        // 老师查看学生信息
        const res = await getStudentProfile(targetId)
        setStudentInfo(res.data)
        setEvaluation(res.data.teacher_evaluation || '')
      }
    } catch (error) {
      console.error('获取信息失败:', error)
      Taro.showToast({ title: '获取信息失败', icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  /** 保存评语 */
  const handleSaveEvaluation = async () => {
    if (!targetId) return
    
    try {
      await updateStudentEvaluation(targetId, evaluation)
      Taro.showToast({ title: '保存成功', icon: 'success' })
      setEditMode(false)
      fetchData()
    } catch (error) {
      console.error('保存评语失败:', error)
      Taro.showToast({ title: '保存失败', icon: 'none' })
    }
  }

  if (!visible) return null

  return (
    <View className='avatar-popup' onClick={onClose}>
      <View className='avatar-popup__content' onClick={(e) => e.stopPropagation()}>
        <View className='avatar-popup__header'>
          <Text className='avatar-popup__title'>
            {userRole === 'student' ? '班级信息' : '学生信息'}
          </Text>
          <View className='avatar-popup__close' onClick={onClose}>
            <Text>✕</Text>
          </View>
        </View>

        {loading ? (
          <View className='avatar-popup__loading'>
            <Text>加载中...</Text>
          </View>
        ) : userRole === 'student' && classInfo ? (
          // 学生查看班级信息
          <View className='avatar-popup__body'>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>班级名称</Text>
              <Text className='avatar-popup__value'>{classInfo.name}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>学科</Text>
              <Text className='avatar-popup__value'>{classInfo.subject}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>老师昵称</Text>
              <Text className='avatar-popup__value'>{classInfo.teacher_name}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>班级简介</Text>
              <Text className='avatar-popup__value'>{classInfo.description || '暂无简介'}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>成员数</Text>
              <Text className='avatar-popup__value'>{classInfo.member_count} 人</Text>
            </View>
          </View>
        ) : userRole === 'teacher' && studentInfo ? (
          // 老师查看学生信息
          <View className='avatar-popup__body'>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>学生昵称</Text>
              <Text className='avatar-popup__value'>{studentInfo.nickname}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>年龄</Text>
              <Text className='avatar-popup__value'>{studentInfo.age || '未设置'}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>性别</Text>
              <Text className='avatar-popup__value'>{studentInfo.gender || '未设置'}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>家庭情况</Text>
              <Text className='avatar-popup__value'>{studentInfo.family_info || '未设置'}</Text>
            </View>
            <View className='avatar-popup__field'>
              <Text className='avatar-popup__label'>所属班级</Text>
              <Text className='avatar-popup__value'>{studentInfo.class_name || '未加入班级'}</Text>
            </View>
            <View className='avatar-popup__field avatar-popup__field--column'>
              <View className='avatar-popup__label-row'>
                <Text className='avatar-popup__label'>老师评语</Text>
                {!editMode && (
                  <View className='avatar-popup__edit-btn' onClick={() => setEditMode(true)}>
                    <Text>编辑</Text>
                  </View>
                )}
              </View>
              {editMode ? (
                <View className='avatar-popup__edit-area'>
                  <Input
                    className='avatar-popup__input'
                    value={evaluation}
                    onInput={(e) => setEvaluation(e.detail.value)}
                    placeholder='请输入评语'
                    multiline
                  />
                  <View className='avatar-popup__btns'>
                    <View className='avatar-popup__btn avatar-popup__btn--cancel' onClick={() => setEditMode(false)}>
                      <Text>取消</Text>
                    </View>
                    <View className='avatar-popup__btn avatar-popup__btn--confirm' onClick={handleSaveEvaluation}>
                      <Text>保存</Text>
                    </View>
                  </View>
                </View>
              ) : (
                <Text className='avatar-popup__value'>{studentInfo.teacher_evaluation || '暂无评语'}</Text>
              )}
            </View>
          </View>
        ) : (
          <View className='avatar-popup__empty'>
            <Text>暂无信息</Text>
          </View>
        )}
      </View>
    </View>
  )
}
