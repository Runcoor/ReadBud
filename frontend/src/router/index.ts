import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/Login.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    name: 'Workbench',
    component: () => import('@/pages/Workbench.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/task/:id',
    name: 'TaskDetail',
    component: () => import('@/pages/TaskDetail.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard placeholder — will be implemented with auth store
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('readbud_token')
  if (to.meta.requiresAuth && !token) {
    next({ name: 'Login' })
  } else if (to.name === 'Login' && token) {
    next({ name: 'Workbench' })
  } else {
    next()
  }
})

export default router
