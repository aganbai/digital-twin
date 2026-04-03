import { Component, PropsWithChildren } from 'react'
import Taro from '@tarojs/taro'
import { getToken } from './utils/storage'
import './app.scss'

class App extends Component<PropsWithChildren> {
  componentDidMount() {
    // 延迟检查登录态，确保路由已初始化
    setTimeout(() => {
      this.checkAuth()
    }, 100)
  }

  componentDidShow() {
    // 每次小程序切到前台时检查登录态
    this.checkAuth()
  }

  componentDidHide() {}

  /** 路由守卫：检查登录态，无 token 跳转登录页 */
  checkAuth() {
    const token = getToken()
    const currentPage = Taro.getCurrentInstance()?.router?.path || ''

    // 白名单页面不需要登录
    const whiteList = ['/pages/login/index', 'pages/login/index', '/pages/role-select/index', 'pages/role-select/index']
    const isWhiteListed = whiteList.some((path) => currentPage.includes(path))

    // 如果当前页面未确定或已在白名单中，不做跳转
    if (!currentPage || isWhiteListed) {
      return
    }

    if (!token) {
      Taro.redirectTo({ url: '/pages/login/index' })
    }
  }

  render() {
    // this.props.children 是将要会渲染的页面
    return this.props.children
  }
}

export default App
