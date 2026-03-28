import { View, Text } from '@tarojs/components'
import './index.scss'

interface EmptyProps {
  text?: string
}

export default function Empty(props: EmptyProps) {
  const { text = '暂无数据' } = props

  return (
    <View className='empty'>
      <View className='empty__icon'>
        <Text className='empty__icon-text'>∅</Text>
      </View>
      <Text className='empty__text'>{text}</Text>
    </View>
  )
}
