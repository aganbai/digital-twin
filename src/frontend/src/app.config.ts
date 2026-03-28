export default defineAppConfig({
  pages: [
    'pages/login/index',
    'pages/role-select/index',
    'pages/home/index',
    'pages/chat/index',
    'pages/history/index',
    'pages/knowledge/index',
    'pages/knowledge/add',
    'pages/memories/index',
    'pages/profile/index',
  ],
  tabBar: {
    color: '#999999',
    selectedColor: '#4F46E5',
    backgroundColor: '#FFFFFF',
    borderStyle: 'white',
    list: [
      {
        pagePath: 'pages/home/index',
        text: '首页',
        iconPath: '',
        selectedIconPath: '',
      },
      {
        pagePath: 'pages/history/index',
        text: '历史',
        iconPath: '',
        selectedIconPath: '',
      },
      {
        pagePath: 'pages/profile/index',
        text: '我的',
        iconPath: '',
        selectedIconPath: '',
      },
    ],
  },
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#4F46E5',
    navigationBarTitleText: 'AI 数字分身',
    navigationBarTextStyle: 'white',
  },
})
