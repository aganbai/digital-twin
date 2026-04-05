import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '@/utils/auth'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/role-select',
      name: 'RoleSelect',
      component: () => import('@/views/RoleSelect.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      component: () => import('@/views/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/chat-list' },
        { path: 'home', name: 'Home', component: () => import('@/views/Home.vue'), meta: { title: '工作台' } },
        { path: 'chat-list', name: 'ChatList', component: () => import('@/views/ChatList.vue'), meta: { title: '消息列表' } },
        { path: 'chat/:id', name: 'Chat', component: () => import('@/views/Chat.vue'), meta: { title: '对话' } },
        { path: 'classes', name: 'Classes', component: () => import('@/views/Classes.vue'), meta: { title: '班级管理' } },
        { path: 'class/:id', name: 'ClassDetail', component: () => import('@/views/ClassDetail.vue'), meta: { title: '班级详情' } },
        { path: 'knowledge', name: 'Knowledge', component: () => import('@/views/Knowledge.vue'), meta: { title: '知识库' } },
        { path: 'courses', name: 'Courses', component: () => import('@/views/Courses.vue'), meta: { title: '课程管理' } },
        { path: 'personas', name: 'Personas', component: () => import('@/views/Personas.vue'), meta: { title: '分身管理' } },
        { path: 'profile', name: 'Profile', component: () => import('@/views/Profile.vue'), meta: { title: '个人中心' } },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/NotFound.vue') },
  ],
})

// 路由守卫
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth !== false && !isLoggedIn()) {
    next('/login')
  } else {
    next()
  }
})

export default router
