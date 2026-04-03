import { Component } from 'react'
import { View } from '@tarojs/components'
import CustomTabBar from '../components/CustomTabBar'
import './index.scss'

export default class CustomTabBarWrapper extends Component {
  render() {
    return (
      <View>
        <CustomTabBar />
      </View>
    )
  }
}
