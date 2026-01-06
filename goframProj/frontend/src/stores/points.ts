import { defineStore } from 'pinia'
import dayjs from 'dayjs'
import { showFailToast, showSuccessToast } from 'vant'
import {
  apiCheckinDaily,
  apiCheckinRetroactive,
  apiCheckinCalendar,
  type CheckinCalendarDetail
} from '@/api/checkin'
import { apiPointsSummary, apiPointsRecords, type BackendPointsRecord } from '@/api/points'

/**
 * ✅ 重要原则（按你的要求）：
 * - “签到/补签/漏签”的展示只允许来自后端真实数据（DB/Redis 计算结果）
 * - 不再使用 localStorage / 本地缓存去“推断”或“记住”签到状态
 * - 如果某月日历还没拉到：UI 不显示漏签/可补签点（避免误判）
 */

export type DayInfo = { signed: boolean; makeup?: boolean }
export type MonthState = {
  days: Record<number, DayInfo>
  /** 本月已补签次数（来自后端 retroCheckedInDays） */
  makeupUsed: number
  /** 预留：后端暂未返回“已发放奖励”，前端不再本地记账 */
  bonusAwarded: Record<string, boolean>
}

function monthKey(d: dayjs.Dayjs) {
  return d.format('YYYY-MM')
}

function mapBackendRecord(r: BackendPointsRecord, idx: number) {
  const t = r.transactionTime || ''
  const type = Number(r.transactionType) || 0
  const delta = Number(r.pointsChange) || 0

  return {
    id: `${t}-${type}-${delta}-${idx}`,
    date: t ? dayjs(t).format('YYYY-MM-DD') : '',
    title: r.description || '积分变动',
    points: delta
  }
}

export const usePointsStore = defineStore('points', {
  state: () => ({
    username: '' as string,

    /** ✅ 后端积分数据（真实） */
    backendTotal: 0 as number,
    backendRecords: [] as Array<{ id: string; date: string; title: string; points: number }>,
    backendLoading: false as boolean,

    /** ✅ 后端签到日历（真实）：yearMonth -> detail */
    calendar: {} as Record<string, CheckinCalendarDetail | undefined>,
    calendarLoading: {} as Record<string, boolean>,

    /** UI 状态 */
    checkinLoading: false as boolean
  }),

  getters: {
    today(): string {
      return dayjs().format('YYYY-MM-DD')
    },
    currentMonthKey(): string {
      return dayjs().format('YYYY-MM')
    },

    totalPoints(): number {
      return this.backendTotal || 0
    },

    monthPoints(): number {
      const mk = this.currentMonthKey
      return this.backendRecords
        .filter(r => r.date.startsWith(mk))
        .reduce((acc, r) => acc + (Number(r.points) || 0), 0)
    },

    /** ✅ 连续签到天数：完全以当前月后端返回为准 */
    streakDays(): number {
      const cur = this.calendar[this.currentMonthKey]
      if (cur && typeof cur.consecutiveDays === 'number') return Math.max(0, cur.consecutiveDays)
      return 0
    },

    /** ✅ 剩余补签次数：完全以当前月后端返回为准 */
    makeupLeft(): number {
      const cur = this.calendar[this.currentMonthKey]
      if (cur && typeof cur.remainRetroTimes === 'number') return Math.max(0, cur.remainRetroTimes)
      return 0
    },

    /** 某月日历是否已经从后端加载完成（成功拿到 detail） */
    isCalendarReady: (s) => (yearMonth: string) => {
      return !!s.calendar[yearMonth]
    },
    isCalendarLoading: (s) => (yearMonth: string) => {
      return !!s.calendarLoading[yearMonth]
    }
  },

  actions: {
    reset() {
      this.username = ''
      this.backendTotal = 0
      this.backendRecords = []
      this.backendLoading = false
      this.calendar = {}
      this.calendarLoading = {}
      this.checkinLoading = false
    },

    async initForUser(username: string) {
      this.username = username
      await Promise.allSettled([
        this.refreshBackendPoints(),
        this.refreshCalendar(this.currentMonthKey, true)
      ])
    },

    /**
     * 从后端同步积分（总积分 + 流水）
     *
     * ✅ 注意：你的 apiPointsRecords 使用 limit + offset（不是 page）
     * 返回结构是 { total, hasMore, list }
     */
    async refreshBackendPoints(options?: { limit?: number; maxPages?: number }) {
      const limit = options?.limit ?? 50
      const maxPages = options?.maxPages ?? 10

      if (this.backendLoading) return
      this.backendLoading = true
      try {
        const s = await apiPointsSummary()
        // 你的 PointsSummaryRes 是 { total: number }
        this.backendTotal = Number((s as any)?.total ?? 0)

        // 拉流水（分页：offset）
        const all: BackendPointsRecord[] = []
        for (let page = 1; page <= maxPages; page++) {
          const offset = (page - 1) * limit
          const res = await apiPointsRecords({ limit, offset })

          const list = (res as any)?.list || []
          if (!Array.isArray(list) || list.length === 0) break

          all.push(...list)

          // 优先用后端的 hasMore 控制
          const hasMore = Boolean((res as any)?.hasMore)
          if (!hasMore) break

          // 兜底：如果返回不足 limit，也认为没有更多
          if (list.length < limit) break
        }

        // ✅ 关键：显式传 idx
        this.backendRecords = all.map((r, idx) => mapBackendRecord(r, idx))
      } catch (e) {
        console.error('[points] refreshBackendPoints failed', e)
      } finally {
        this.backendLoading = false
      }
    },

    /**
     * ✅ 从后端获取某月签到日历（真实）
     * - 不使用 localStorage 做任何“兜底”
     * - 如果请求失败：不改动该月 calendar（避免用旧数据误标）
     */
    async refreshCalendar(yearMonth: string, force = false) {
      if (!force && this.calendar[yearMonth]) return
      if (this.calendarLoading[yearMonth]) return
      this.calendarLoading[yearMonth] = true
      try {
        const res = await apiCheckinCalendar(yearMonth)
        this.calendar[yearMonth] = res.detail
      } catch (e) {
        console.error('[checkin] refreshCalendar failed', yearMonth, e)
        delete this.calendar[yearMonth]
      } finally {
        this.calendarLoading[yearMonth] = false
      }
    },

    /**
     * 将后端日历 detail 转换为现有 MonthState 结构（供组件复用）
     * - retroCheckedInDays 也算“已签到”（绿色）
     * - makeupUsed = retroCheckedInDays.length
     */
    monthStateFromCalendar(yearMonth: string): MonthState | null {
      const detail = this.calendar[yearMonth]
      if (!detail) return null
      const date = dayjs(yearMonth + '-01')
      const daysInMonth = date.daysInMonth()

      const checked = new Set<number>(detail.checkedInDays || [])
      const retro = new Set<number>(detail.retroCheckedInDays || [])

      const days: Record<number, DayInfo> = {}
      for (let i = 1; i <= daysInMonth; i++) {
        const signed = checked.has(i) || retro.has(i)
        days[i] = signed ? { signed: true, makeup: retro.has(i) } : { signed: false }
      }

      return {
        days,
        makeupUsed: (detail.retroCheckedInDays || []).length,
        bonusAwarded: {}
      }
    },

    getMonthState(yearMonth?: string): MonthState | null {
      const key = yearMonth || this.currentMonthKey
      return this.monthStateFromCalendar(key)
    },

    /**
     * ✅ 判断某一天是否已签到：只允许基于后端日历
     */
    isSigned(dateStr: string): boolean {
      const d = dayjs(dateStr)
      const mk = monthKey(d)
      const detail = this.calendar[mk]
      if (!detail) return false
      const day = d.date()
      return (detail.checkedInDays || []).includes(day) || (detail.retroCheckedInDays || []).includes(day)
    },

    async checkinToday() {
      if (this.checkinLoading) return
      this.checkinLoading = true
      try {
        await apiCheckinDaily()
        showSuccessToast('签到成功')
        await Promise.allSettled([
          this.refreshBackendPoints(),
          this.refreshCalendar(this.currentMonthKey, true)
        ])
      } catch (e: any) {
        const msg = e?.response?.data?.message || e?.response?.data?.msg || e?.message || '签到失败'
        showFailToast(msg)
        await Promise.allSettled([
          this.refreshBackendPoints(),
          this.refreshCalendar(this.currentMonthKey, true)
        ])
      } finally {
        this.checkinLoading = false
      }
    },

    async makeup(dateStr: string) {
      if (this.checkinLoading) return
      this.checkinLoading = true
      try {
        await apiCheckinRetroactive(dateStr)
        showSuccessToast('补签成功')
        const mk = monthKey(dayjs(dateStr))
        await Promise.allSettled([
          this.refreshBackendPoints(),
          this.refreshCalendar(mk, true)
        ])
      } catch (e: any) {
        const msg = e?.response?.data?.message || e?.response?.data?.msg || e?.message || '补签失败'
        showFailToast(msg)
      } finally {
        this.checkinLoading = false
      }
    },

    getRecordsByMonth(key: string) {
      return this.backendRecords.filter(r => r.date.startsWith(key))
    }
  }
})
