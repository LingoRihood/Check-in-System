import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import LoginView from '@/views/LoginView.vue'
import PointsView from '@/views/PointsView.vue'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/', name: 'home', component: HomeView },
    { path: '/points', name: 'points', component: PointsView }
  ],
  scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.hydrated) {
    auth.hydrate()
    if (auth.isLoggedIn) {
      try {
        await auth.fetchMe()
      } catch {
        auth.logout()
      }
    }
  }

  if (to.meta.public) return true
  if (auth.isLoggedIn) return true
  return { path: '/login', query: { redirect: to.fullPath } }
})

export default router
