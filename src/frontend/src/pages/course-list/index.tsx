import { useState, useEffect, useCallback } from 'react'
import Taro, { useReachBottom, usePullDownRefresh } from '@tarojs/taro'
import { View, Text } from '@tarojs/components'
import { getCourses, deleteCourse, CourseInfo } from '../../api/course'
import { getClasses } from '../../api/class'
import './index.scss'

interface ClassInfoV8 {
  id: number
  name: string
  subject?: string
}

/** 课程列表页 */
export default function CourseListPage() {
  const [courses, setCourses] = useState<CourseInfo[]>([])
  const [classes, setClasses] = useState<ClassInfoV8[]>([])
  const [selectedClassId, setSelectedClassId] = useState<number>(0)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchClasses()
  }, [])

  useEffect(() => {
    if (selectedClassId) {
      fetchCourses(true)
    }
  }, [selectedClassId])

  /** 获取班级列表 */
  const fetchClasses = async () => {
    try {
      const res = await getClasses()
      const classList = Array.isArray(res.data) ? res.data : (res.data as any).items || []
      setClasses(classList as ClassInfoV8[])
      
      if (classList.length > 0) {
        setSelectedClassId((classList[0] as ClassInfoV8).id)
      }
    } catch (error) {
      console.error('获取班级列表失败:', error)
    }
  }

  /** 获取课程列表 */
  const fetchCourses = async (refresh = false) => {
    if (loading) return
    
    setLoading(true)
    try {
      const currentPage = refresh ? 1 : page
      const res = await getCourses(selectedClassId, currentPage, 20)
      
      const items = res.data.items || res.data as any
      const newCourses = Array.isArray(items) ? items : []
      
      if (refresh) {
        setCourses(newCourses)
        setPage(1)
      } else {
        setCourses([...courses, ...newCourses])
      }
      
      setHasMore(newCourses.length >= 20)
      setPage(currentPage + 1)
    } catch (error) {
      console.error('获取课程列表失败:', error)
      Taro.showToast({ title: '加载失败', icon: 'none' })
    } finally {
      setLoading(false)
      Taro.stopPullDownRefresh()
    }
  }

  /** 下拉刷新 */
  usePullDownRefresh(() => {
    fetchCourses(true)
  })

  /** 上拉加载更多 */
  useReachBottom(() => {
    if (hasMore && !loading) {
      fetchCourses()
    }
  })

  /** 新建课程 */
  const handleCreate = () => {
    Taro.navigateTo({
      url: `/pages/course-publish/index?classId=${selectedClassId}`,
    })
  }

  /** 编辑课程 */
  const handleEdit = (id: number) => {
    Taro.navigateTo({
      url: `/pages/course-publish/index?id=${id}`,
    })
  }

  /** 删除课程 */
  const handleDelete = (id: number) => {
    Taro.showModal({
      title: '确认删除',
      content: '删除后无法恢复，是否继续？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await deleteCourse(id)
            Taro.showToast({ title: '删除成功', icon: 'success' })
            fetchCourses(true)
          } catch (error) {
            console.error('删除失败:', error)
          }
        }
      },
    })
  }

  /** 推送课程 */
  const handlePush = async (id: number) => {
    try {
      Taro.showLoading({ title: '推送中...' })
      // 这里调用推送API
      Taro.hideLoading()
      Taro.showToast({ title: '推送成功', icon: 'success' })
    } catch (error) {
      Taro.hideLoading()
      console.error('推送失败:', error)
    }
  }

  /** 格式化时间 */
  const formatTime = (time: string) => {
    const date = new Date(time)
    return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`
  }

  return (
    <View className='course-list'>
      {/* 班级筛选 */}
      <View className='course-list__tabs'>
        {classes.map((cls) => (
          <View
            key={cls.id}
            className={`course-list__tab ${selectedClassId === cls.id ? 'course-list__tab--active' : ''}`}
            onClick={() => setSelectedClassId(cls.id)}
          >
            <Text>{cls.name}</Text>
          </View>
        ))}
      </View>

      {/* 课程列表 */}
      <View className='course-list__content'>
        {courses.length === 0 && !loading ? (
          <View className='course-list__empty'>
            <Text className='course-list__empty-text'>暂无课程</Text>
            <View className='course-list__empty-btn' onClick={handleCreate}>
              <Text>发布课程</Text>
            </View>
          </View>
        ) : (
          <>
            {courses.map((course) => (
              <View key={course.id} className='course-list__item'>
                <View className='course-list__item-header'>
                  <Text className='course-list__item-title'>{course.title}</Text>
                  <Text className='course-list__item-time'>{formatTime(course.created_at)}</Text>
                </View>
                <Text className='course-list__item-content'>{course.content}</Text>
                <View className='course-list__item-footer'>
                  <View className='course-list__item-actions'>
                    <View className='course-list__action' onClick={() => handleEdit(course.id)}>
                      <Text>编辑</Text>
                    </View>
                    <View className='course-list__action' onClick={() => handleDelete(course.id)}>
                      <Text>删除</Text>
                    </View>
                    {!course.pushed && (
                      <View className='course-list__action course-list__action--primary' onClick={() => handlePush(course.id)}>
                        <Text>推送</Text>
                      </View>
                    )}
                  </View>
                  {course.pushed && (
                    <Text className='course-list__pushed'>已推送</Text>
                  )}
                </View>
              </View>
            ))}
            
            {loading && (
              <View className='course-list__loading'>
                <Text>加载中...</Text>
              </View>
            )}
            
            {!hasMore && courses.length > 0 && (
              <View className='course-list__no-more'>
                <Text>没有更多了</Text>
              </View>
            )}
          </>
        )}
      </View>

      {/* 新建按钮 */}
      <View className='course-list__fab' onClick={handleCreate}>
        <Text className='course-list__fab-icon'>+</Text>
      </View>
    </View>
  )
}
