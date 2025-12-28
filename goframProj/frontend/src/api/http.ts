import axios, { AxiosError, type AxiosInstance, type AxiosResponse } from 'axios'
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

/** ✅ GoFrame 默认响应结构：{ code, message, data } */
function isGFResp(x: any): x is { code: number; message: string; data: any } {
  return (
    x &&
    typeof x === 'object' &&
    typeof x.code === 'number' &&
    typeof x.message === 'string' &&
    'data' in x
  )
}

http.interceptors.response.use(
  (res: AxiosResponse) => {
    // ✅ 关键：HTTP 200 也可能是业务失败，需要用 code 判断
    const payload = res.data
    if (isGFResp(payload) && payload.code !== 0) {
      const err: any = new Error(payload.message || '请求失败')
      // 让上层还能用 e.response.data.message 取到后端提示
      err.response = { ...res, data: payload }
      err.gfCode = payload.code
      throw err
    }
    return res
  },
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

        // 刷新接口也可能是 GoFrame 包装
        const payload = resp.data as any
        const data = (payload?.data ?? payload) as any
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
