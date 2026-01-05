import { defineStore } from 'pinia'
import dayjs from 'dayjs'
import { showFailToast, showSuccessToast } from 'vant'
import { apiCheckinDaily, apiCheckinRetroactive } from '@/api/checkin'
import { apiPointsSummary, apiPointsRecords, type BackendPointsRecord } from '@/api/points'
import {
  clearUserPoints,
  loadUserPoints,
  saveUserPoints,
  type MonthState,
  type PointsRecord,
  type UserPointsState
} from '@/utils/pointsStorage'

const RULES = {
  daily: 1,
  milestones: [
    { days: 3, bonus: 5 },
    { days: 7, bonus: 10 },
    { days: 15, bonus: 20 }
  ],
  fullMonthBonus: 100,
  makeupCost: 100,
  maxMakeupPerMonth: 3
}

function uid() {
  return (crypto.randomUUID ? crypto.randomUUID() : String(Date.now()) + Math.random())
}

function monthKey(d: dayjs.Dayjs) {
  return d.format('YYYY-MM')
}

function ensureMonth(state: UserPointsState, key: string, date: dayjs.Dayjs): MonthState {
  if (state.months[key]) return state.months[key]
  const daysInMonth = date.daysInMonth()
  const days: Record<number, { signed: boolean; makeup?: boolean }> = {}
  for (let i = 1; i <= daysInMonth; i++) days[i] = { signed: false }
  const m: MonthState = {
    days,
    makeupUsed: 0,
    bonusAwarded: { '3': false, '7': false, '15': false, full: false }
  }
  state.months[key] = m
  return m
}

function sumPoints(records: PointsRecord[]) {
  return records.reduce((s, r) => s + r.points, 0)
}

function backendRecordToPointsRecord(r: BackendPointsRecord, fallbackId: string): PointsRecord {
  const date = (r.transactionTime || '').slice(0, 10) || dayjs().format('YYYY-MM-DD')
  // 后端 transactionType: 1=每日签到, 2=连续签到奖励, 3=补签(可能包含消耗/奖励)
  let type: PointsRecord['type'] = 'earn'
  if (r.transactionType === 1) type = 'checkin'
  else if (r.transactionType === 2) type = 'bonus'
  else if (r.transactionType === 3) type = r.pointsChange < 0 ? 'makeup_cost' : 'checkin'

  return {
    id: `${fallbackId}`,
    date,
    title: r.description || '积分变动',
    type,
    points: Number(r.pointsChange) || 0
  }
}


export const usePointsStore = defineStore('points', {
  state: () => ({
    username: '' as string,
    data: null as UserPointsState | null,
    checkinLoading: false as boolean,

    // 后端同步的积分数据（用于确保前后端一致）
    backendTotal: null as number | null,
    backendRecords: [] as PointsRecord[],
    backendLoading: false as boolean
  }),

  getters: {
    ready: (s) => !!s.username && !!s.data,
    today(): string {
      return dayjs().format('YYYY-MM-DD')
    },
    currentMonthKey(): string {
      return dayjs().format('YYYY-MM')
    },
    totalPoints(): number {
      if (this.backendTotal != null) return this.backendTotal
      if (!this.data) return 0
      return sumPoints(this.data.records)
    },
    monthPoints(): number {
      const key = this.currentMonthKey
      // 优先使用后端记录计算
      if (this.backendRecords.length) {
        return sumPoints(this.backendRecords.filter(r => r.date.startsWith(key)))
      }
      if (!this.data) return 0
      return sumPoints(this.data.records.filter(r => r.date.startsWith(key)))
    },
    streakDays(): number {
      if (!this.data) return 0
      const now = dayjs()
      const mk = monthKey(now)
      const m = ensureMonth(this.data, mk, now)
      let streak = 0
      for (let d = now.date(); d >= 1; d--) {
        if (m.days[d]?.signed) streak++
        else break
      }
      return streak
    },
    makeupLeft(): number {
      if (!this.data) return RULES.maxMakeupPerMonth
      const now = dayjs()
      const mk = monthKey(now)
      const m = ensureMonth(this.data, mk, now)
      return Math.max(0, RULES.maxMakeupPerMonth - (m.makeupUsed || 0))
    }
  },

  actions: {
    /**
     * 从后端同步总积分与积分明细（用于替代/校准本地存储，保证一致）
     */
    async refreshBackendPoints(options?: { limit?: number; maxPages?: number }) {
      const limit = options?.limit ?? 50
      const maxPages = options?.maxPages ?? 10
      if (this.backendLoading) return
      this.backendLoading = true
      try {
        // 1) 总积分
        const s = await apiPointsSummary()
        this.backendTotal = Number(s?.total) || 0

        // 2) 明细（分页拉取，最多 maxPages 页）
        const all: PointsRecord[] = []
        let offset = 0
        for (let page = 0; page < maxPages; page++) {
          const res = await apiPointsRecords({ limit, offset })
          const list = Array.isArray(res?.list) ? res.list : []
          list.forEach((item, i) => {
            all.push(backendRecordToPointsRecord(item as any, `${offset + i}-${item.transactionTime}`))
          })
          if (!res?.hasMore) break
          offset += limit
        }
        this.backendRecords = all
      } catch (e) {
        // 同步失败不阻断页面；仍可使用本地数据
        console.error('[points] refreshBackendPoints failed', e)
      } finally {
        this.backendLoading = false
      }
    },

    reset() {
      this.username = ''
      this.data = null
    },

    initForUser(username: string) {
      if (!username) return
      this.username = username

      const saved = loadUserPoints(username)
      this.data = saved || { version: 1, records: [], months: {} }

      const now = dayjs()
      ensureMonth(this.data!, monthKey(now), now)

      // 启动后端同步（不阻塞 UI）
      void this.refreshBackendPoints({ limit: 50, maxPages: 10 })

      this.persist()
    },

    persist() {
      if (!this.username || !this.data) return
      saveUserPoints(this.username, this.data)
    },

    logoutClearLocal() {
      if (this.username) clearUserPoints(this.username)
      this.reset()
    },

    getMonthState(key?: string) {
      if (!this.data) return null
      const d = key ? dayjs(key + '-01') : dayjs()
      const mk = key || monthKey(d)
      return ensureMonth(this.data, mk, d)
    },

    isSigned(dateStr: string) {
      if (!this.data) return false
      const d = dayjs(dateStr)
      const mk = monthKey(d)
      const m = ensureMonth(this.data, mk, d)
      return !!m.days[d.date()]?.signed
    },

    async checkinToday() {
      if (!this.data) return
      if (this.checkinLoading) return

      const d = dayjs()
      const mk = monthKey(d)
      const m = ensureMonth(this.data, mk, d)
      const day = d.date()

      // 先走后端：真正执行“签到 + 发积分”
      this.checkinLoading = true
      try {
        await apiCheckinDaily()
      } catch (e: any) {
        // 如果后端说“今日已签到”，把本地状态也同步成已签到（但不重复加分）
        const backendMsg =
          e?.response?.data?.message || e?.response?.data?.msg || e?.message || ''

        if (typeof backendMsg === 'string' && backendMsg.includes('已签到')) {
          if (!m.days[day]?.signed) {
            m.days[day] = { signed: true }
            this.persist()
          }
        }

        showFailToast(backendMsg || '签到失败')
        return
      } finally {
        this.checkinLoading = false
      }

      // 后端成功后，再更新本地展示（用于日历/明细/动画）
      if (m.days[day]?.signed) {
        // 兜底：本地已是已签到（比如换设备后刚同步），就不重复记账
        showSuccessToast('签到成功')
        return
      }

      m.days[day] = { signed: true }

      // 后端会负责记账与发积分，这里只更新本地日历状态，然后同步后端积分数据
      this.persist()
      await this.refreshBackendPoints({ limit: 50, maxPages: 10 })

      showSuccessToast('签到成功')
    },

    async makeup(dateStr: string) {
      if (!this.data) return
      const target = dayjs(dateStr)
      const today = dayjs().startOf('day')

      if (target.isAfter(today)) return showFailToast('不能补签未来日期')

      const mk = monthKey(target)
      const m = ensureMonth(this.data, mk, target)
      const day = target.date()

      if (m.days[day].signed) return showFailToast('该日已签到')
      if (mk !== this.currentMonthKey) return showFailToast('仅支持当月补签')

      // 后端会校验剩余补签次数/积分是否足够；前端这里只做基础拦截（避免明显无效请求）
      if (m.makeupUsed >= RULES.maxMakeupPerMonth) return showFailToast('本月补签次数已用完')
      if (this.totalPoints < RULES.makeupCost) return showFailToast('积分不足，无法补签')

      try {
        await apiCheckinRetroactive(target.format('YYYY-MM-DD'))
      } catch (e: any) {
        return showFailToast(e?.response?.data?.message || e?.message || '补签失败')
      }

      // 本地仅更新日历展示；积分明细/总分以服务端为准
      m.days[day] = { signed: true, makeup: true }
      m.makeupUsed += 1
      this.persist()
      await this.refreshBackendPoints({ limit: 50, maxPages: 10 })

      showSuccessToast('补签成功')
    },

    applyBonusesAfterChange(today: dayjs.Dayjs) {
      if (!this.data) return
      const mk = monthKey(today)
      const m = ensureMonth(this.data, mk, today)

      const streak = this.streakDays
      for (const ms of RULES.milestones) {
        const key = String(ms.days)
        if (streak >= ms.days && !m.bonusAwarded[key]) {
          m.bonusAwarded[key] = true
          this.data.records.unshift({
            id: uid(),
            date: today.format('YYYY-MM-DD'),
            title: `连签奖励（${ms.days} 天）`,
            type: 'bonus',
            points: ms.bonus
          })
        }
      }

      const daysInMonth = today.daysInMonth()
      let allSigned = true
      for (let i = 1; i <= daysInMonth; i++) {
        if (!m.days[i]?.signed) { allSigned = false; break }
      }
      if (allSigned && !m.bonusAwarded.full) {
        m.bonusAwarded.full = true
        this.data.records.unshift({
          id: uid(),
          date: today.format('YYYY-MM-DD'),
          title: '本月满签奖励',
          type: 'bonus',
          points: RULES.fullMonthBonus
        })
      }
    },

    getRecordsByMonth(key: string) {
      // 若已从后端拉取过记录，优先使用后端记录（保证一致）
      if (this.backendRecords.length) return this.backendRecords.filter(r => r.date.startsWith(key))
      if (!this.data) return []
      return this.data.records.filter(r => r.date.startsWith(key))
    }
  }
})
