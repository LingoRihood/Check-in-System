import { http } from './http'

export type LoginReq = { username: string; password: string }
export type LoginRes = { accessToken: string; refreshToken: string }
export type RegisterReq = { username: string; email: string; password: string; confirmPassword: string }
export type RegisterRes = { userID: number; username: string }
export type MeRes = { username: string; email: string; avatar: string }

function unwrap<T>(payload: any): T {
  if (payload && typeof payload === 'object' && 'data' in payload) return payload.data as T
  return payload as T
}

export async function apiLogin(req: LoginReq): Promise<LoginRes> {
  const res = await http.post('/auth/login', req)
  return unwrap<LoginRes>(res.data)
}
export async function apiRegister(req: RegisterReq): Promise<RegisterRes> {
  const res = await http.post('/users', req)
  return unwrap<RegisterRes>(res.data)
}
export async function apiMe(): Promise<MeRes> {
  const res = await http.get('/users/me')
  return unwrap<MeRes>(res.data)
}
