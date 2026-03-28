import { Component, PropsWithChildren } from 'react'
import Taro from '@tarojs/taro'
import { getToken } from './utils/storage'
import './app.scss'

class App extends Component<PropsWithChildren> {
  componentDidMount() {
    this.checkAuth()
  }

  componentDidShow() {}

  componentDidHide() {}

  /** 路由守卫：检查登录态，无 token 跳转登录页 */
  checkAuth() {
    const token = getToken()
    const currentPage = Taro.getCurrentInstance()?.router?.path || ''

    // 白名单页面不需要登录
    const whiteList = ['/pages/login/index', 'pages/login/index']
    const isWhiteListed = whiteList.some((path) => currentPage.includes(path))

    if (!token && !isWhiteListed) {
      Taro.redirectTo({ url: '/pages/login/index' })
    }
  }

  render() {
    // this.props.children 是将要会渲染的页面
    return this.props.children
  }
}

export default App
