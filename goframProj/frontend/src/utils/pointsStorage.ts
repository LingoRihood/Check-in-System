export type PointsRecordType = 'checkin' | 'makeup_cost' | 'bonus' | 'spend' | 'earn'

export type PointsRecord = {
  id: string
  date: string // YYYY-MM-DD
  title: string
  type: PointsRecordType
  points: number
}

export type DayInfo = { signed: boolean; makeup?: boolean }

export type MonthState = {
  days: Record<number, DayInfo>
  makeupUsed: number
  bonusAwarded: Record<string, boolean> // '3'|'7'|'15'|'full'
}

export type UserPointsState = {
  version: 1
  records: PointsRecord[]
  months: Record<string, MonthState> // YYYY-MM
}

const prefix = 'checkin_h5_points_v1::'

export function loadUserPoints(username: string): UserPointsState | null {
  try {
    const raw = localStorage.getItem(prefix + username)
    return raw ? (JSON.parse(raw) as UserPointsState) : null
  } catch {
    return null
  }
}

export function saveUserPoints(username: string, data: UserPointsState) {
  localStorage.setItem(prefix + username, JSON.stringify(data))
}

export function clearUserPoints(username: string) {
  localStorage.removeItem(prefix + username)
}
