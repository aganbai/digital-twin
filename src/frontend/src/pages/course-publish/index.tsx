import { useState, useEffect } from 'react'
import Taro, { useRouter } from '@tarojs/taro'
import { View, Text, Textarea, Input, Picker, Switch } from '@tarojs/components'
import { createCourse, updateCourse } from '../../api/course'
import { getClasses, getClassInfo } from '../../api/class'
import './index.scss'

interface ClassInfoV8 {
  id: number
  name: string
  subject?: string
}

/** 课程发布页 */
export default function CoursePublishPage() {
  const router = useRouter()
  const editId = router.params.id // 编辑模式下的课程ID

  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [classId, setClassId] = useState<number>(0)
  const [pushToStudents, setPushToStudents] = useState(false)
  const [classes, setClasses] = useState<ClassInfoV8[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchClasses()
    if (editId) {
      // 编辑模式：加载课程数据
      fetchCourseData(parseInt(editId))
    }
  }, [editId])

  /** 获取班级列表 */
  const fetchClasses = async () => {
    try {
      const res = await getClasses()
      // 兼容不同响应格式
      const classList = Array.isArray(res.data) ? res.data : (res.data as any).items || []
      setClasses(classList as ClassInfoV8[])
      
      // 默认选择第一个班级
      if (classList.length > 0 && !classId) {
        setClassId((classList[0] as ClassInfoV8).id)
      }
    } catch (error) {
      console.error('获取班级列表失败:', error)
    }
  }

  /** 获取课程数据（编辑模式） */
  const fetchCourseData = async (id: number) => {
    try {
      // 这里需要单独的课程详情API，暂不实现
      Taro.showToast({ title: '加载课程数据', icon: 'none' })
    } catch (error) {
      console.error('获取课程数据失败:', error)
    }
  }

  /** 提交课程 */
  const handleSubmit = async () => {
    if (!title.trim()) {
      Taro.showToast({ title: '请输入课程标题', icon: 'none' })
      return
    }
    if (!content.trim()) {
      Taro.showToast({ title: '请输入课程内容', icon: 'none' })
      return
    }
    if (!classId) {
      Taro.showToast({ title: '请选择班级', icon: 'none' })
      return
    }

    setLoading(true)
    try {
      if (editId) {
        // 编辑模式
        await updateCourse(parseInt(editId), { title, content })
        Taro.showToast({ title: '更新成功', icon: 'success' })
        setTimeout(() => {
          Taro.navigateBack()
        }, 1500)
      } else {
        // 新建模式
        const res = await createCourse({
          title,
          content,
          class_id: classId,
          push_to_students: pushToStudents,
        })
        Taro.showToast({ title: '发布成功', icon: 'success' })
        // 跳转到课程列表页
        setTimeout(() => {
          Taro.redirectTo({ url: '/pages/course-list/index' })
        }, 1500)
      }
    } catch (error) {
      console.error('提交失败:', error)
    } finally {
      setLoading(false)
    }
  }

  return (
    <View className='course-publish'>
      <View className='course-publish__form'>
        {/* 标题 */}
        <View className='course-publish__field'>
          <Text className='course-publish__label'>课程标题</Text>
          <Input
            className='course-publish__input'
            value={title}
            onInput={(e) => setTitle(e.detail.value)}
            placeholder='请输入课程标题'
            maxLength={100}
          />
        </View>

        {/* 内容 */}
        <View className='course-publish__field'>
          <Text className='course-publish__label'>课程内容</Text>
          <Textarea
            className='course-publish__textarea'
            value={content}
            onInput={(e) => setContent(e.detail.value)}
            placeholder='请输入课程内容'
            maxlength={5000}
          />
          <Text className='course-publish__count'>{content.length}/5000</Text>
        </View>

        {/* 选择班级 */}
        <View className='course-publish__field'>
          <Text className='course-publish__label'>选择班级</Text>
          <Picker
            mode='selector'
            range={classes}
            rangeKey='name'
            value={classes.findIndex(c => c.id === classId)}
            onChange={(e) => setClassId(classes[parseInt(e.detail.value)].id)}
          >
            <View className='course-publish__picker'>
              <Text className='course-publish__picker-value'>
                {classes.find(c => c.id === classId)?.name || '请选择班级'}
              </Text>
              <Text className='course-publish__picker-arrow'>▼</Text>
            </View>
          </Picker>
        </View>

        {/* 推送选项 */}
        {!editId && (
          <View className='course-publish__field course-publish__field--row'>
            <View className='course-publish__switch-label'>
              <Text className='course-publish__label'>推送给学生</Text>
              <Text className='course-publish__hint'>发布后自动推送通知给学生</Text>
            </View>
            <Switch
              checked={pushToStudents}
              onChange={(e) => setPushToStudents(e.detail.value)}
              color='#4F46E5'
            />
          </View>
        )}

        {/* 提交按钮 */}
        <View className='course-publish__submit' onClick={handleSubmit}>
          <Text className='course-publish__submit-text'>
            {loading ? '提交中...' : editId ? '保存修改' : '发布课程'}
          </Text>
        </View>
      </View>
    </View>
  )
}
