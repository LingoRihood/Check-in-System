import { http } from './http'

/**
 * 每日签到
 * 后端：POST /api/v1/checkins
 * 成功一般返回 { code:0, message:'', data:{} }（由 http.ts 拦截器统一处理）
 */
export async function apiCheckinDaily(): Promise<void> {
  await http.post('/checkins')
}
