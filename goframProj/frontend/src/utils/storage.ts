const KEY = 'checkin_h5_auth_v1'

export type StoredAuth = {
  accessToken: string
  refreshToken: string
  user?: { username: string; email: string; avatar: string }
}

export function loadAuth(): StoredAuth | null {
  try {
    const raw = localStorage.getItem(KEY)
    return raw ? (JSON.parse(raw) as StoredAuth) : null
  } catch {
    return null
  }
}

export function saveAuth(data: StoredAuth) {
  localStorage.setItem(KEY, JSON.stringify(data))
}

export function clearAuth() {
  localStorage.removeItem(KEY)
}
