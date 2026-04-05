import { useState } from 'react'
import Taro from '@tarojs/taro'
import { View, Text } from '@tarojs/components'
import './index.scss'

interface PlusPanelProps {
  /** 是否显示 */
  visible: boolean
  /** 关闭回调 */
  onClose: () => void
  /** 选择文件回调 */
  onFileSelect?: (files: string[]) => void
  /** 选择相册回调 */
  onImageSelect?: (images: string[]) => void
  /** 拍摄回调 */
  onCameraCapture?: (image: string) => void
}

/** 功能项配置 */
const ACTION_ITEMS = [
  {
    id: 'file',
    icon: '📁',
    label: '文件',
    description: '选择本地文件',
  },
  {
    id: 'image',
    icon: '📷',
    label: '相册',
    description: '选择图片或视频',
  },
  {
    id: 'camera',
    icon: '📸',
    label: '拍摄',
    description: '拍照或录像',
  },
]

/**
 * +号多功能面板
 * 支持文件、相册、拍摄功能
 */
export default function PlusPanel({
  visible,
  onClose,
  onFileSelect,
  onImageSelect,
  onCameraCapture,
}: PlusPanelProps) {
  if (!visible) return null

  /** 选择文件 */
  const handleFileSelect = async () => {
    try {
      // 微信小程序选择文件
      const res = await Taro.chooseMessageFile({
        count: 5,
        type: 'file',
      })
      
      const filePaths = res.tempFiles.map(f => f.path)
      onFileSelect?.(filePaths)
      onClose()
    } catch (error) {
      console.error('选择文件失败:', error)
      Taro.showToast({ title: '选择文件失败', icon: 'none' })
    }
  }

  /** 选择相册 */
  const handleImageSelect = async () => {
    try {
      const res = await Taro.chooseMedia({
        count: 9,
        mediaType: ['image', 'video'],
        sourceType: ['album'],
      })
      
      const mediaPaths = res.tempFiles.map(f => f.tempFilePath)
      onImageSelect?.(mediaPaths)
      onClose()
    } catch (error) {
      console.error('选择相册失败:', error)
      Taro.showToast({ title: '选择相册失败', icon: 'none' })
    }
  }

  /** 拍摄 */
  const handleCameraCapture = async () => {
    try {
      const res = await Taro.chooseMedia({
        count: 1,
        mediaType: ['image', 'video'],
        sourceType: ['camera'],
      })
      
      const mediaPath = res.tempFiles[0].tempFilePath
      onCameraCapture?.(mediaPath)
      onClose()
    } catch (error) {
      console.error('拍摄失败:', error)
      Taro.showToast({ title: '拍摄失败', icon: 'none' })
    }
  }

  /** 点击功能项 */
  const handleAction = (id: string) => {
    switch (id) {
      case 'file':
        handleFileSelect()
        break
      case 'image':
        handleImageSelect()
        break
      case 'camera':
        handleCameraCapture()
        break
    }
  }

  return (
    <View className='plus-panel' onClick={onClose}>
      <View className='plus-panel__content' onClick={(e) => e.stopPropagation()}>
        <View className='plus-panel__header'>
          <Text className='plus-panel__title'>选择功能</Text>
        </View>
        <View className='plus-panel__actions'>
          {ACTION_ITEMS.map((item) => (
            <View
              key={item.id}
              className='plus-panel__action'
              onClick={() => handleAction(item.id)}
            >
              <View className='plus-panel__action-icon'>
                <Text>{item.icon}</Text>
              </View>
              <View className='plus-panel__action-info'>
                <Text className='plus-panel__action-label'>{item.label}</Text>
                <Text className='plus-panel__action-desc'>{item.description}</Text>
              </View>
            </View>
          ))}
        </View>
        <View className='plus-panel__footer'>
          <View className='plus-panel__cancel' onClick={onClose}>
            <Text className='plus-panel__cancel-text'>取消</Text>
          </View>
        </View>
      </View>
    </View>
  )
}
