import { useState } from 'react'
import { View, Text, Input } from '@tarojs/components'
import './index.scss'

interface TagInputProps {
  tags: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
  maxTags?: number // 默认 5
}

export default function TagInput(props: TagInputProps) {
  const { tags, onChange, placeholder = '输入标签后按回车添加', maxTags = 5 } = props
  const [inputValue, setInputValue] = useState('')

  /** 确认添加标签 */
  const handleConfirm = () => {
    const value = inputValue.trim()
    if (!value) return
    // 去重
    if (tags.includes(value)) {
      setInputValue('')
      return
    }
    // 数量限制
    if (tags.length >= maxTags) return
    onChange([...tags, value])
    setInputValue('')
  }

  /** 删除标签 */
  const handleDelete = (index: number) => {
    const newTags = tags.filter((_, i) => i !== index)
    onChange(newTags)
  }

  return (
    <View className='tag-input'>
      <View className='tag-input__tags'>
        {tags.map((tag, index) => (
          <View key={tag} className='tag-input__tag'>
            <Text className='tag-input__tag-text'>{tag}</Text>
            <Text className='tag-input__tag-close' onClick={() => handleDelete(index)}>×</Text>
          </View>
        ))}
      </View>
      {tags.length < maxTags && (
        <Input
          className='tag-input__input'
          value={inputValue}
          placeholder={placeholder}
          onInput={(e) => setInputValue(e.detail.value)}
          onConfirm={handleConfirm}
        />
      )}
    </View>
  )
}
