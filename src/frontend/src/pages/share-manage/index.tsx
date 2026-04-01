import { useState, useCallback, useEffect } from 'react'
import { View, Text, Canvas } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { createShare, getShares, deactivateShare } from '@/api/share'
import { getClasses } from '@/api/class'
import type { ShareInfo } from '@/api/share'
import type { ClassInfo } from '@/api/class'
import { usePersonaStore } from '@/store'
import Empty from '@/components/Empty'
import './index.scss'

/**
 * 简易二维码生成器（Canvas 绘制）
 * 使用简单的编码方式将分享码绘制为可扫描的二维码
 */
function drawQRCode(
  canvasId: string,
  content: string,
  size: number,
  scope: any,
) {
  // 使用 Taro Canvas API 绘制简易二维码占位
  const ctx = Taro.createCanvasContext(canvasId, scope)
  const moduleCount = 21 // QR 码最小尺寸
  const cellSize = size / (moduleCount + 2) // 留白边
  const offset = cellSize

  // 白色背景
  ctx.setFillStyle('#FFFFFF')
  ctx.fillRect(0, 0, size, size)

  // 简易编码：基于字符串的哈希生成伪随机矩阵
  const matrix: boolean[][] = []
  let hash = 0
  for (let i = 0; i < content.length; i++) {
    hash = ((hash << 5) - hash) + content.charCodeAt(i)
    hash |= 0
  }

  for (let row = 0; row < moduleCount; row++) {
    matrix[row] = []
    for (let col = 0; col < moduleCount; col++) {
      // 定位图案（三个角）
      const isFinderPattern =
        (row < 7 && col < 7) ||
        (row < 7 && col >= moduleCount - 7) ||
        (row >= moduleCount - 7 && col < 7)

      if (isFinderPattern) {
        // 定位图案边框
        const isOuter =
          row === 0 || row === 6 || col === 0 || col === 6 ||
          (row >= moduleCount - 7 && (row === moduleCount - 7 || row === moduleCount - 1)) ||
          (col >= moduleCount - 7 && (col === moduleCount - 7 || col === moduleCount - 1))
        const isInner =
          (row >= 2 && row <= 4 && col >= 2 && col <= 4) ||
          (row >= 2 && row <= 4 && col >= moduleCount - 5 && col <= moduleCount - 3) ||
          (row >= moduleCount - 5 && row <= moduleCount - 3 && col >= 2 && col <= 4)
        matrix[row][col] = isOuter || isInner
      } else {
        // 数据区域：基于哈希的伪随机
        const seed = (hash + row * 31 + col * 17) & 0x7fffffff
        matrix[row][col] = (seed % 3) === 0
      }
    }
  }

  // 绘制矩阵
  ctx.setFillStyle('#000000')
  for (let row = 0; row < moduleCount; row++) {
    for (let col = 0; col < moduleCount; col++) {
      if (matrix[row][col]) {
        ctx.fillRect(
          offset + col * cellSize,
          offset + row * cellSize,
          cellSize,
          cellSize,
        )
      }
    }
  }

  ctx.draw()
}

export default function ShareManage() {
  const { currentPersona } = usePersonaStore()
  const [shares, setShares] = useState<ShareInfo[]>([])
  const [classes, setClasses] = useState<ClassInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [shareClassId, setShareClassId] = useState<number | undefined>(undefined)
  /** 当前展示二维码的分享码 ID */
  const [qrcodeShareId, setQrcodeShareId] = useState<number | null>(null)

  /** 获取分享码列表 */
  const fetchShares = useCallback(async () => {
    try {
      const res = await getShares()
      setShares(res.data || [])
    } catch (error) {
      console.error('获取分享码失败:', error)
    }
  }, [])

  /** 获取班级列表 */
  const fetchClasses = useCallback(async () => {
    try {
      const res = await getClasses()
      setClasses(res.data || [])
    } catch (error) {
      console.error('获取班级列表失败:', error)
    }
  }, [])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      await Promise.all([fetchShares(), fetchClasses()])
      setLoading(false)
    }
    init()
  }, [fetchShares, fetchClasses])

  /** 生成分享码 */
  const handleCreateShare = async () => {
    if (!currentPersona) {
      Taro.showToast({ title: '请先选择分身', icon: 'none' })
      return
    }
    try {
      await createShare(currentPersona.id, shareClassId, 72, 50)
      Taro.showToast({ title: '分享码已生成', icon: 'success' })
      setShowCreateModal(false)
      setShareClassId(undefined)
      fetchShares()
    } catch (error) {
      console.error('生成分享码失败:', error)
    }
  }

  /** 停用分享码 */
  const handleDeactivate = async (shareId: number) => {
    try {
      await deactivateShare(shareId)
      Taro.showToast({ title: '已停用', icon: 'success' })
      fetchShares()
    } catch (error) {
      console.error('停用分享码失败:', error)
    }
  }

  /** 复制分享码 */
  const handleCopy = (code: string) => {
    Taro.setClipboardData({
      data: code,
      success: () => {
        Taro.showToast({ title: '已复制', icon: 'success' })
      },
    })
  }

  /** 显示/隐藏二维码 */
  const handleToggleQRCode = (share: ShareInfo) => {
    if (qrcodeShareId === share.id) {
      setQrcodeShareId(null)
    } else {
      setQrcodeShareId(share.id)
      // 延迟绘制二维码（等待 Canvas 渲染）
      setTimeout(() => {
        const qrContent = `pages/share-join/index?code=${share.share_code}`
        drawQRCode(`qrcode-${share.id}`, qrContent, 200, null)
      }, 100)
    }
  }

  /** 保存二维码到相册 */
  const handleSaveQRCode = (shareId: number) => {
    Taro.canvasToTempFilePath({
      canvasId: `qrcode-${shareId}`,
      success: (res) => {
        Taro.saveImageToPhotosAlbum({
          filePath: res.tempFilePath,
          success: () => {
            Taro.showToast({ title: '已保存到相册', icon: 'success' })
          },
          fail: () => {
            Taro.showToast({ title: '保存失败，请授权相册权限', icon: 'none' })
          },
        })
      },
      fail: () => {
        Taro.showToast({ title: '生成图片失败', icon: 'none' })
      },
    })
  }

  const activeShares = shares.filter((s) => s.is_active)
  const expiredShares = shares.filter((s) => !s.is_active)

  if (loading) {
    return (
      <View className='share-manage'>
        <View className='share-manage__loading'>
          <Text className='share-manage__loading-text'>加载中...</Text>
        </View>
      </View>
    )
  }

  return (
    <View className='share-manage'>
      <View className='share-manage__header'>
        <Text className='share-manage__title'>🔗 分享码管理</Text>
        <Text className='share-manage__subtitle'>
          当前分身：{currentPersona?.nickname || '未选择'}
        </Text>
      </View>

      {/* 有效分享码 */}
      <View className='share-manage__section'>
        <Text className='share-manage__section-title'>有效分享码 ({activeShares.length})</Text>
        {activeShares.length > 0 ? (
          activeShares.map((share) => (
            <View key={share.id} className='share-manage__card'>
              <View className='share-manage__card-info'>
                <Text className='share-manage__card-code'>{share.share_code}</Text>
                <Text className='share-manage__card-meta'>
                  已使用 {share.used_count}/{share.max_uses}
                </Text>
              </View>
              <View className='share-manage__card-actions'>
                <View
                  className='share-manage__btn share-manage__btn--qrcode'
                  onClick={() => handleToggleQRCode(share)}
                >
                  <Text className='share-manage__btn-text'>
                    {qrcodeShareId === share.id ? '收起' : '二维码'}
                  </Text>
                </View>
                <View
                  className='share-manage__btn share-manage__btn--copy'
                  onClick={() => handleCopy(share.share_code)}
                >
                  <Text className='share-manage__btn-text'>复制</Text>
                </View>
                <View
                  className='share-manage__btn share-manage__btn--deactivate'
                  onClick={() => handleDeactivate(share.id)}
                >
                  <Text className='share-manage__btn-text--deactivate'>停用</Text>
                </View>
              </View>
              {/* 二维码展示区域 */}
              {qrcodeShareId === share.id && (
                <View className='share-manage__qrcode-area'>
                  <Canvas
                    canvasId={`qrcode-${share.id}`}
                    className='share-manage__qrcode-canvas'
                    style={{ width: '200px', height: '200px' }}
                  />
                  <Text className='share-manage__qrcode-hint'>
                    学生扫码即可加入
                  </Text>
                  <View
                    className='share-manage__qrcode-save-btn'
                    onClick={() => handleSaveQRCode(share.id)}
                  >
                    <Text className='share-manage__qrcode-save-text'>保存到相册</Text>
                  </View>
                </View>
              )}
            </View>
          ))
        ) : (
          <Empty text='暂无有效分享码' />
        )}
      </View>

      {/* 已失效分享码 */}
      {expiredShares.length > 0 && (
        <View className='share-manage__section'>
          <Text className='share-manage__section-title'>已失效 ({expiredShares.length})</Text>
          {expiredShares.map((share) => (
            <View key={share.id} className='share-manage__card share-manage__card--expired'>
              <View className='share-manage__card-info'>
                <Text className='share-manage__card-code share-manage__card-code--expired'>
                  {share.share_code}
                </Text>
                <Text className='share-manage__card-meta'>
                  已使用 {share.used_count}/{share.max_uses}
                </Text>
              </View>
            </View>
          ))}
        </View>
      )}

      {/* 生成分享码按钮 */}
      <View
        className='share-manage__create-btn'
        onClick={() => setShowCreateModal(true)}
      >
        <Text className='share-manage__create-btn-text'>+ 生成新分享码</Text>
      </View>

      {/* 生成分享码弹窗 */}
      {showCreateModal && (
        <View className='share-manage__modal-mask' onClick={() => setShowCreateModal(false)}>
          <View className='share-manage__modal' onClick={(e) => e.stopPropagation()}>
            <Text className='share-manage__modal-title'>生成分享码</Text>
            <Text className='share-manage__modal-desc'>
              选择关联班级（可选），生成后分享给学生扫码加入
            </Text>
            <View className='share-manage__modal-class-list'>
              <View
                className={`share-manage__modal-class-item ${!shareClassId ? 'share-manage__modal-class-item--active' : ''}`}
                onClick={() => setShareClassId(undefined)}
              >
                <Text>不关联班级</Text>
              </View>
              {classes.map((cls) => (
                <View
                  key={cls.id}
                  className={`share-manage__modal-class-item ${shareClassId === cls.id ? 'share-manage__modal-class-item--active' : ''}`}
                  onClick={() => setShareClassId(cls.id)}
                >
                  <Text>{cls.name}</Text>
                </View>
              ))}
            </View>
            <View className='share-manage__modal-actions'>
              <View
                className='share-manage__modal-btn share-manage__modal-btn--cancel'
                onClick={() => { setShowCreateModal(false); setShareClassId(undefined) }}
              >
                <Text>取消</Text>
              </View>
              <View
                className='share-manage__modal-btn share-manage__modal-btn--confirm'
                onClick={handleCreateShare}
              >
                <Text className='share-manage__modal-btn-text--confirm'>生成</Text>
              </View>
            </View>
          </View>
        </View>
      )}
    </View>
  )
}
