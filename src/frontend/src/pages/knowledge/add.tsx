import { useState } from 'react'
import { View, Text, Input, Textarea } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { addDocument } from '@/api/knowledge'
import TagInput from '@/components/TagInput'
import './add.scss'

export default function KnowledgeAdd() {
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)

  /** 提交文档 */
  const handleSubmit = async () => {
    // 校验标题
    const trimmedTitle = title.trim()
    if (!trimmedTitle || trimmedTitle.length > 200) {
      Taro.showToast({ title: '标题为必填项，长度 1-200 字符', icon: 'none' })
      return
    }
    // 校验内容
    const trimmedContent = content.trim()
    if (!trimmedContent || trimmedContent.length > 100000) {
      Taro.showToast({ title: '内容为必填项，长度 1-100000 字符', icon: 'none' })
      return
    }

    setSubmitting(true)
    try {
      await addDocument(trimmedTitle, trimmedContent, tags)
      Taro.showToast({ title: '添加成功', icon: 'success' })
      setTimeout(() => {
        Taro.navigateBack()
      }, 1500)
    } catch (error) {
      console.error('添加文档失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <View className='knowledge-add-page'>
      {/* 标题输入 */}
      <View className='knowledge-add-page__section'>
        <Text className='knowledge-add-page__label'>标题</Text>
        <Input
          className='knowledge-add-page__title-input'
          placeholder='请输入文档标题'
          value={title}
          maxlength={200}
          onInput={(e) => setTitle(e.detail.value)}
        />
      </View>

      {/* 内容输入 */}
      <View className='knowledge-add-page__section'>
        <Text className='knowledge-add-page__label'>内容</Text>
        <View className='knowledge-add-page__textarea-wrap'>
          <Textarea
            className='knowledge-add-page__textarea'
            placeholder='请输入文档内容...'
            value={content}
            maxlength={100000}
            onInput={(e) => setContent(e.detail.value)}
          />
          <Text className='knowledge-add-page__word-count'>{content.length} 字</Text>
        </View>
      </View>

      {/* 标签输入 */}
      <View className='knowledge-add-page__section'>
        <Text className='knowledge-add-page__label'>标签</Text>
        <TagInput tags={tags} onChange={setTags} maxTags={5} />
      </View>

      {/* 提交按钮 */}
      <View
        className={`knowledge-add-page__submit ${submitting ? 'knowledge-add-page__submit--disabled' : ''}`}
        onClick={submitting ? undefined : handleSubmit}
      >
        <Text className='knowledge-add-page__submit-text'>
          {submitting ? '提交中...' : '提交'}
        </Text>
      </View>
    </View>
  )
}
