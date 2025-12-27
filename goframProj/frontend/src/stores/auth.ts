import { defineStore } from 'pinia'
import { apiLogin, apiMe, apiRegister, type MeRes } from '@/api/auth'
import { clearAuth, loadAuth, saveAuth } from '@/utils/storage'
import { showFailToast, showSuccessToast } from 'vant'
import { usePointsStore } from './points'

type User = MeRes

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: '' as string,
    refreshToken: '' as string,
    user: null as User | null,
    hydrated: false
  }),
  getters: {
    isLoggedIn: (s) => !!s.accessToken && !!s.refreshToken
  },
  actions: {
    hydrate() {
      const saved = loadAuth()
      if (saved) {
        this.accessToken = saved.accessToken
        this.refreshToken = saved.refreshToken
        this.user = saved.user || null
      }
      this.hydrated = true
    },
    persist() {
      saveAuth({
        accessToken: this.accessToken,
        refreshToken: this.refreshToken,
        user: this.user || undefined
      })
    },
    async login(username: string, password: string) {
      try {
        const res = await apiLogin({ username, password })
        this.accessToken = res.accessToken
        this.refreshToken = res.refreshToken
        this.persist()

        await this.fetchMe()

        usePointsStore().initForUser(this.user?.username || username)

        showSuccessToast('登录成功')
      } catch (e: any) {
        showFailToast(e?.response?.data?.message || e?.message || '登录失败')
        throw e
      }
    },
    async register(username: string, email: string, password: string, confirmPassword: string) {
      try {
        await apiRegister({ username, email, password, confirmPassword })
        showSuccessToast('注册成功，请登录')
      } catch (e: any) {
        showFailToast(e?.response?.data?.message || e?.message || '注册失败')
        throw e
      }
    },
    async fetchMe() {
      const me = await apiMe()
      this.user = me
      this.persist()
      if (me?.username) usePointsStore().initForUser(me.username)
      return me
    },
    logout() {
      this.accessToken = ''
      this.refreshToken = ''
      this.user = null
      clearAuth()
      usePointsStore().reset()
    }
  }
})
