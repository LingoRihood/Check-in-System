import axios, { AxiosError, type AxiosInstance } from 'axios'
import { loadAuth, saveAuth, clearAuth } from '@/utils/storage'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 15000
})

http.interceptors.request.use((config) => {
  const auth = loadAuth()
  if (auth?.accessToken) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${auth.accessToken}`
  }
  return config
})

let refreshing = false
let pending: Array<(token: string | null) => void> = []

function resolvePending(token: string | null) {
  pending.forEach((fn) => fn(token))
  pending = []
}

http.interceptors.response.use(
  (res) => res,
  async (err: AxiosError) => {
    const status = err.response?.status
    const original = err.config as any

    if (status === 401 && !original?._retry) {
      original._retry = true
      const auth = loadAuth()
      if (!auth?.refreshToken) {
        clearAuth()
        return Promise.reject(err)
      }

      if (refreshing) {
        return new Promise((resolve, reject) => {
          pending.push((token) => {
            if (!token) return reject(err)
            original.headers = original.headers || {}
            original.headers.Authorization = `Bearer ${token}`
            resolve(http(original))
          })
        })
      }

      refreshing = true
      try {
        const resp = await axios.post(
          `${baseURL}/auth/refresh`,
          { refreshToken: auth.refreshToken },
          { timeout: 15000 }
        )
        const data = (resp.data as any)?.data ?? resp.data
        const newAccess = data?.accessToken
        const newRefresh = data?.refreshToken

        if (newAccess && newRefresh) {
          saveAuth({ ...auth, accessToken: newAccess, refreshToken: newRefresh })
          resolvePending(newAccess)
          original.headers = original.headers || {}
          original.headers.Authorization = `Bearer ${newAccess}`
          return http(original)
        }

        clearAuth()
        resolvePending(null)
        return Promise.reject(err)
      } catch (e) {
        clearAuth()
        resolvePending(null)
        return Promise.reject(e)
      } finally {
        refreshing = false
      }
    }

    return Promise.reject(err)
  }
)
