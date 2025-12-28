import { defineStore } from 'pinia'
import { apiLogin, apiMe, apiRegister, type MeRes } from '@/api/auth'
import { clearAuth, loadAuth, saveAuth } from '@/utils/storage'
import { showFailToast, showSuccessToast } from 'vant'
import { usePointsStore } from './points'

type User = MeRes

function pickTokens(res: any): { accessToken: string; refreshToken: string } | null {
  // 兼容：res / res.data / res.data.data
  const data = res?.data?.data ?? res?.data ?? res

  const accessToken = data?.accessToken ?? data?.token ?? ''
  const refreshToken = data?.refreshToken ?? data?.refresh_token ?? ''

  if (!accessToken || !refreshToken) return null
  return { accessToken, refreshToken }
}

function getNiceErrorMessage(e: unknown, fallback: string) {
  // 后端消息：只在确实有值时才用
  const resp = (e as any)?.response
  const data = resp?.data

  const backendMsg = data?.message ?? data?.msg
  if (typeof backendMsg === 'string' && backendMsg.trim().length > 0) {
    return backendMsg.trim()
  }

  // 本地异常 message
  const rawMsg = (e as any)?.message
  const msg = typeof rawMsg === 'string' ? rawMsg.trim() : ''

  // 过滤掉 TypeError 这类不该给用户看的
  if (msg && !msg.includes('Cannot read properties')) {
    return msg
  }

  return fallback
}

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

        const tokens = pickTokens(res)
        if (!tokens) {
          const backendMsg = (res as any)?.message || (res as any)?.msg
          showFailToast(
            typeof backendMsg === 'string' && backendMsg.trim()
              ? backendMsg.trim()
              : '用户名或密码错误'
          )
          return
        }


        this.accessToken = tokens.accessToken
        this.refreshToken = tokens.refreshToken
        this.persist()

        await this.fetchMe()

        usePointsStore().initForUser(this.user?.username || username)

        showSuccessToast('登录成功')
      } catch (e: any) {
        // ❗不要把 TypeError 原文给用户看
        showFailToast(getNiceErrorMessage(e, '登录失败，请稍后重试'))
        // 失败不要污染 token
        this.accessToken = ''
        this.refreshToken = ''
        this.user = null
        clearAuth()
        throw e
      }
    },

    /**
     * ✅ 关键改动：
     * - 注册成功：返回 true（页面再切回登录 tab）
     * - 注册失败：toast + 返回 false（页面不要跳）
     * - 不再 throw（避免页面逻辑没写好导致 finally 里跳转）
     */
    async register(username: string, email: string, password: string, confirmPassword: string): Promise<boolean> {
      // 本地先校验（避免请求）
      if (password !== confirmPassword) {
        showFailToast('两次密码需保持一致')
        return false
      }

      try {
        await apiRegister({ username, email, password, confirmPassword })
        showSuccessToast('注册成功，请登录')
        return true
      } catch (e: any) {
        showFailToast(getNiceErrorMessage(e, '注册失败，请稍后重试'))
        return false
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
