import { useState, useEffect, useCallback } from 'react'
import { View, Text, Input, Textarea, Picker } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { submitAssignment, submitAssignmentWithFile } from '@/api/assignment'
import { getRelations, RelationItemStudent } from '@/api/relation'
import './index.scss'

export default function SubmitAssignment() {
  const [teachers, setTeachers] = useState<RelationItemStudent[]>([])
  const [selectedTeacherIndex, setSelectedTeacherIndex] = useState(-1)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [filePath, setFilePath] = useState('')
  const [fileName, setFileName] = useState('')
  const [submitting, setSubmitting] = useState(false)

  /** 获取已授权教师列表 */
  const fetchTeachers = useCallback(async () => {
    try {
      const res = await getRelations({ status: 'approved', page: 1, page_size: 100 })
      const items = (res.data.items || []) as RelationItemStudent[]
      setTeachers(items)
    } catch (error) {
      console.error('获取教师列表失败:', error)
    }
  }, [])

  useEffect(() => {
    fetchTeachers()
  }, [fetchTeachers])

  /** 选择文件 */
  const handleChooseFile = () => {
    Taro.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: ['pdf', 'docx', 'txt', 'md'],
      success: (res) => {
        if (res.tempFiles && res.tempFiles.length > 0) {
          setFilePath(res.tempFiles[0].path)
          setFileName(res.tempFiles[0].name)
        }
      },
    })
  }

  /** 移除文件 */
  const handleRemoveFile = () => {
    setFilePath('')
    setFileName('')
  }

  /** 提交作业 */
  const handleSubmit = async () => {
    if (selectedTeacherIndex < 0) {
      Taro.showToast({ title: '请选择教师', icon: 'none' })
      return
    }
    if (!title.trim()) {
      Taro.showToast({ title: '请输入标题', icon: 'none' })
      return
    }
    if (!content.trim() && !filePath) {
      Taro.showToast({ title: '请输入内容或上传附件', icon: 'none' })
      return
    }

    const teacher = teachers[selectedTeacherIndex]
    if (!teacher) return

    setSubmitting(true)
    try {
      if (filePath) {
        // 含文件上传
        await submitAssignmentWithFile(
          filePath,
          teacher.teacher_id,
          title.trim(),
          content.trim() || undefined,
        )
      } else {
        // 纯文本
        await submitAssignment({
          teacher_id: teacher.teacher_id,
          title: title.trim(),
          content: content.trim(),
        })
      }
      Taro.showToast({ title: '提交成功', icon: 'success' })
      setTimeout(() => Taro.navigateBack(), 1500)
    } catch (error) {
      console.error('提交作业失败:', error)
    } finally {
      setSubmitting(false)
    }
  }

  const teacherNames = teachers.map((t) => `${t.teacher_nickname} · ${t.teacher_school || ''}`)

  return (
    <View className='submit-assignment-page'>
      {/* 选择教师 */}
      <View className='submit-assignment-page__section'>
        <Text className='submit-assignment-page__label'>教师</Text>
        <Picker
          mode='selector'
          range={teacherNames}
          value={selectedTeacherIndex}
          onChange={(e) => setSelectedTeacherIndex(Number(e.detail.value))}
        >
          <View className='submit-assignment-page__picker'>
            <Text className='submit-assignment-page__picker-text'>
              {selectedTeacherIndex >= 0
                ? teacherNames[selectedTeacherIndex]
                : '选择教师 ▼'}
            </Text>
          </View>
        </Picker>
      </View>

      {/* 标题 */}
      <View className='submit-assignment-page__section'>
        <Text className='submit-assignment-page__label'>标题</Text>
        <Input
          className='submit-assignment-page__input'
          placeholder='请输入作业标题'
          value={title}
          maxlength={200}
          onInput={(e) => setTitle(e.detail.value)}
        />
      </View>

      {/* 内容 */}
      <View className='submit-assignment-page__section'>
        <Text className='submit-assignment-page__label'>内容</Text>
        <Textarea
          className='submit-assignment-page__textarea'
          placeholder='请输入作业内容...'
          value={content}
          maxlength={10000}
          onInput={(e) => setContent(e.detail.value)}
        />
      </View>

      {/* 附件 */}
      <View className='submit-assignment-page__section'>
        <Text className='submit-assignment-page__label'>附件（可选）</Text>
        {filePath ? (
          <View className='submit-assignment-page__file-info'>
            <Text className='submit-assignment-page__file-name'>📄 {fileName}</Text>
            <Text className='submit-assignment-page__file-remove' onClick={handleRemoveFile}>
              ✕
            </Text>
          </View>
        ) : (
          <View className='submit-assignment-page__file-btn' onClick={handleChooseFile}>
            <Text className='submit-assignment-page__file-btn-text'>+ 上传文件</Text>
          </View>
        )}
      </View>

      {/* 提交按钮 */}
      <View
        className={`submit-assignment-page__submit ${submitting ? 'submit-assignment-page__submit--disabled' : ''}`}
        onClick={submitting ? undefined : handleSubmit}
      >
        <Text className='submit-assignment-page__submit-text'>
          {submitting ? '提交中...' : '提交作业'}
        </Text>
      </View>
    </View>
  )
}
