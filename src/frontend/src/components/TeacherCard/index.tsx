import { View, Text, Button } from '@tarojs/components'
import './index.scss'

interface TeacherCardProps {
  id: number
  nickname: string
  username: string
  documentCount: number
  onChat: (teacherId: number) => void
}

export default function TeacherCard(props: TeacherCardProps) {
  const { id, nickname, documentCount, onChat } = props

  /** 获取昵称首字母（支持中英文） */
  const getInitial = (name: string): string => {
    if (!name) return '?'
    return name.charAt(0).toUpperCase()
  }

  return (
    <View className='teacher-card'>
      {/* 头像占位：圆形背景 + 首字母 */}
      <View className='teacher-card__avatar'>
        <Text className='teacher-card__avatar-text'>{getInitial(nickname)}</Text>
      </View>

      {/* 教师信息 */}
      <View className='teacher-card__info'>
        <Text className='teacher-card__name'>{nickname}</Text>
        <Text className='teacher-card__doc-count'>{documentCount} 篇文档</Text>
      </View>

      {/* 开始对话按钮 */}
      <View className='teacher-card__action'>
        <Button className='teacher-card__btn' onClick={() => onChat(id)}>
          开始对话
        </Button>
      </View>
    </View>
  )
}
