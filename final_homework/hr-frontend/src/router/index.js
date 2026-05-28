import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', component: () => import('../views/Login.vue') },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    redirect: '/jobs',
    meta: { requiresAuth: true },
    children: [
      { path: 'jobs', component: () => import('../views/Jobs.vue') },
      { path: 'candidates', component: () => import('../views/Candidates.vue') },
      { path: 'ai-chat', component: () => import('../views/AIChat.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    return '/login'
  }
  if (to.path === '/login' && auth.isLoggedIn) {
    return '/jobs'
  }
})

export default router
