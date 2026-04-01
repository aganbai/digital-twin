import { View, Text, Button } from '@tarojs/components'
import './index.scss'

/** 授权状态类型 */
export type AuthStatus = 'approved' | 'pending' | 'none'

interface TeacherCardProps {
  id: number
  nickname: string
  username: string
  school?: string
  description?: string
  documentCount: number
  /** 授权状态 */
  authStatus?: AuthStatus
  /** 点击开始对话 */
  onChat?: (teacherId: number) => void
  /** 点击申请使用 */
  onApply?: (teacherId: number) => void
}

export default function TeacherCard(props: TeacherCardProps) {
  const { id, nickname, school, description, documentCount, authStatus = 'none', onChat, onApply } = props

  /** 获取昵称首字母（支持中英文） */
  const getInitial = (name: string): string => {
    if (!name) return '?'
    return name.charAt(0).toUpperCase()
  }

  /** 渲染操作按钮 */
  const renderAction = () => {
    switch (authStatus) {
      case 'approved':
        return (
          <Button className='teacher-card__btn teacher-card__btn--primary' onClick={() => onChat?.(id)}>
            开始对话
          </Button>
        )
      case 'pending':
        return (
          <View className='teacher-card__status teacher-card__status--pending'>
            <Text className='teacher-card__status-text'>审批中 ⏳</Text>
          </View>
        )
      default:
        return (
          <Button className='teacher-card__btn teacher-card__btn--outline' onClick={() => onApply?.(id)}>
            申请使用
          </Button>
        )
    }
  }

  return (
    <View className='teacher-card'>
      {/* 头像占位：圆形背景 + 首字母 */}
      <View className='teacher-card__avatar'>
        <Text className='teacher-card__avatar-text'>{getInitial(nickname)}</Text>
      </View>

      {/* 教师信息 */}
      <View className='teacher-card__info'>
        <View className='teacher-card__name-row'>
          <Text className='teacher-card__name'>{nickname}</Text>
          {authStatus === 'approved' && (
            <Text className='teacher-card__auth-badge'>✅</Text>
          )}
        </View>
        {school && <Text className='teacher-card__school'>{school}</Text>}
        {description && <Text className='teacher-card__desc'>{description}</Text>}
        <Text className='teacher-card__doc-count'>{documentCount} 篇文档</Text>
      </View>

      {/* 操作区域 */}
      <View className='teacher-card__action'>
        {renderAction()}
      </View>
    </View>
  )
}
