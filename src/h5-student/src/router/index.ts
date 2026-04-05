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
      path: '/',
      component: () => import('@/views/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/chat' },
        { path: 'chat', name: 'Chat', component: () => import('@/views/Chat.vue'), meta: { title: '对话' } },
        { path: 'history', name: 'History', component: () => import('@/views/History.vue'), meta: { title: '历史记录' } },
        { path: 'discover', name: 'Discover', component: () => import('@/views/Discover.vue'), meta: { title: '发现' } },
        { path: 'profile', name: 'Profile', component: () => import('@/views/Profile.vue'), meta: { title: '我的' } },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/NotFound.vue') },
  ],
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth !== false && !isLoggedIn()) {
    next('/login')
  } else {
    next()
  }
})

export default router
