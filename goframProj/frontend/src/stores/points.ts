import { defineStore } from 'pinia'
import dayjs from 'dayjs'
import { showFailToast, showSuccessToast } from 'vant'
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

export const usePointsStore = defineStore('points', {
  state: () => ({
    username: '' as string,
    data: null as UserPointsState | null
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
      if (!this.data) return 0
      return sumPoints(this.data.records)
    },
    monthPoints(): number {
      if (!this.data) return 0
      const key = this.currentMonthKey
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

    checkinToday() {
      if (!this.data) return
      const d = dayjs()
      const mk = monthKey(d)
      const m = ensureMonth(this.data, mk, d)
      const day = d.date()

      if (m.days[day].signed) {
        showFailToast('今天已签到')
        return
      }

      m.days[day] = { signed: true }

      this.data.records.unshift({
        id: uid(),
        date: d.format('YYYY-MM-DD'),
        title: '每日签到',
        type: 'checkin',
        points: RULES.daily
      })

      this.applyBonusesAfterChange(d)
      this.persist()

      showSuccessToast('签到成功 +1')
    },

    makeup(dateStr: string) {
      if (!this.data) return
      const target = dayjs(dateStr)
      const today = dayjs().startOf('day')

      if (target.isAfter(today)) return showFailToast('不能补签未来日期')

      const mk = monthKey(target)
      const m = ensureMonth(this.data, mk, target)
      const day = target.date()

      if (m.days[day].signed) return showFailToast('该日已签到')
      if (mk !== this.currentMonthKey) return showFailToast('仅支持当月补签')
      if (m.makeupUsed >= RULES.maxMakeupPerMonth) return showFailToast('本月补签次数已用完')
      if (this.totalPoints < RULES.makeupCost) return showFailToast('积分不足，无法补签')

      this.data.records.unshift({
        id: uid(),
        date: target.format('YYYY-MM-DD'),
        title: '补签消耗',
        type: 'makeup_cost',
        points: -RULES.makeupCost
      })
      this.data.records.unshift({
        id: uid(),
        date: target.format('YYYY-MM-DD'),
        title: '补签签到',
        type: 'checkin',
        points: RULES.daily
      })

      m.days[day] = { signed: true, makeup: true }
      m.makeupUsed += 1

      this.applyBonusesAfterChange(dayjs())
      this.persist()

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
      if (!this.data) return []
      return this.data.records.filter(r => r.date.startsWith(key))
    }
  }
})
