import { http } from './http'

export type CheckinCalendarDetail = {
  checkedInDays: number[]
  retroCheckedInDays: number[]
  isCheckedInToday: boolean
  remainRetroTimes: number
  consecutiveDays: number
}

export type CheckinCalendarRes = {
  year: number
  month: number
  detail: CheckinCalendarDetail
}

function unwrap<T>(payload: any): T {
  if (payload && typeof payload === 'object' && 'data' in payload) return payload.data as T
  return payload as T
}

/**
 * 每日签到
 * 后端：POST /api/v1/checkins
 * 成功一般返回 { code:0, message:'', data:{} }（由 http.ts 拦截器统一处理）
 */
export async function apiCheckinDaily(): Promise<void> {
  await http.post('/checkins')
}

/** 补签：POST /checkins/retroactive */
export async function apiCheckinRetroactive(date: string): Promise<void> {
  await http.post('/checkins/retroactive', { date })
}

/** 获取某月签到日历：GET /checkins/calendar?yearMonth=YYYY-MM */
export async function apiCheckinCalendar(yearMonth: string): Promise<CheckinCalendarRes> {
  const res = await http.get('/checkins/calendar', { params: { yearMonth } })
  return unwrap<CheckinCalendarRes>(res.data)
}
