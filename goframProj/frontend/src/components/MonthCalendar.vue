<template>
  <div class="rounded-3xl bg-slate-900 text-white shadow-sm overflow-hidden border border-white/10">
    <!-- Header -->
    <div class="px-4 pt-4 pb-3 flex items-center justify-between gap-3">
      <div class="min-w-0">
        <h2 class="mt-0.5 text-lg font-semibold tracking-tight truncate">{{ title }}</h2>
      </div>

      <div class="flex items-center gap-2">
        <button
          class="h-10 w-10 rounded-2xl bg-white/10 border border-white/10 flex items-center justify-center
                 active:scale-[0.98] transition-transform
                 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/40 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-900"
          aria-label="上一月"
          @click="emit('prev')"
        >
          <i class="fa-solid fa-chevron-left text-white/80 text-sm"></i>
        </button>

        <button
          class="h-10 w-10 rounded-2xl bg-white/10 border border-white/10 flex items-center justify-center
                 active:scale-[0.98] transition-transform
                 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/40 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-900"
          aria-label="下一月"
          @click="emit('next')"
        >
          <i class="fa-solid fa-chevron-right text-white/80 text-sm"></i>
        </button>
      </div>
    </div>

    <!-- Week header -->
    <div class="px-3 pb-2">
      <div class="grid grid-cols-7 text-[15px] font-semibold text-white/70">
        <div class="h-10 flex flex-col items-center justify-center"><div>Sun</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Mon</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Tue</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Wed</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Thu</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Fri</div></div>
        <div class="h-10 flex flex-col items-center justify-center"><div>Sat</div></div>
      </div>
    </div>

    <!-- Calendar grid -->
    <div class="px-3 pb-4">
      <div class="grid grid-cols-7 gap-y-3 place-items-center" aria-label="日历签到网格">
        <!-- leading blanks -->
        <div
          v-for="n in leadingBlanks"
          :key="'b' + n"
          class="h-10 flex aspect-square"
        ></div>

        <button
          v-for="day in daysInThisMonth"
          :key="day"
          type="button"
          class="h-10 aspect-square relative flex flex-col items-center justify-center
                 transition-transform duration-150 ease-out
                 active:scale-95
                 focus:outline-none focus-visible:ring-2 focus-visible:ring-white/40 focus-visible:ring-offset-2 focus-visible:ring-offset-slate-900"
          :class="cellClass(day)"
          :disabled="isDisabledDay(day)"
          @click="pick(day)"
        >
          <!-- selection -->
          <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div
              v-if="selectedDay === day"
              class="h-10 w-10 rounded-full bg-emerald-500
                     shadow-[0_10px_18px_-10px_rgba(16,185,129,0.75)]"
            ></div>
          </div>

          <!-- number -->
          <div
            class="relative text-[15px] font-semibold tabular-nums leading-none transition-colors duration-150"
            :class="numClass(day)"
          >
            {{ day }}
          </div>

          <!-- dot -->
          <div class="relative mt-1 h-1.5 flex items-center justify-center">
            <span
              v-if="dotColor(day)"
              class="h-1.5 w-1.5 rounded-full"
              :class="dotColor(day)"
            ></span>
          </div>
        </button>
      </div>

      <!-- Legend（保持你原样不动） -->
      <div class="mt-4 flex flex-wrap items-center gap-x-5 gap-y-2 text-xs text-white/65 px-1">
        <div class="inline-flex items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-full bg-emerald-500"></span>已签到
        </div>
        <div class="inline-flex items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-full bg-amber-400"></span>可补签
        </div>
        <div class="inline-flex items-center gap-2">
          <span class="h-2.5 w-2.5 rounded-full bg-rose-500"></span>漏签
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { usePointsStore } from '@/stores/points'

const props = defineProps<{
  year: number
  monthIndex: number // 0-11
}>()

const emit = defineEmits<{
  (e: 'select', dateStr: string): void
  (e: 'prev'): void
  (e: 'next'): void
}>()

const points = usePointsStore()

const monthStart = computed(() => dayjs(new Date(props.year, props.monthIndex, 1)))
const title = computed(() => `${props.year} 年 ${props.monthIndex + 1} 月`)
const daysInThisMonth = computed(() => monthStart.value.daysInMonth())

const selectedDay = ref<number | null>(dayjs().date())

const leadingBlanks = computed(() => monthStart.value.day())

function mk() {
  return monthStart.value.format('YYYY-MM')
}

function isCurrentMonth() {
  return mk() === dayjs().format('YYYY-MM')
}

/** ✅ 未来月份（下个月及之后） */
function isFutureMonth() {
  return monthStart.value.isAfter(dayjs(), 'month')
}

/** ✅ 是否禁用某一天：未来月份整月禁用；当前月禁用今天之后 */
function isDisabledDay(day: number) {
  if (isFutureMonth()) return true
  if (isFuture(day)) return true
  return false
}

watch(
  () => [props.year, props.monthIndex],
  () => {
    // ✅ 未来月份：不能选中任何日期（选中态清空）
    // 其他月份：默认选中“今天的日期号”（保持你原逻辑）
    selectedDay.value = isFutureMonth() ? null : dayjs().date()
  }
)

// ✅ 月份切换时，从后端同步该月日历（数据库真实数据）
watch(
  () => [props.year, props.monthIndex],
  () => {
    const ym = monthStart.value.format('YYYY-MM')
    void points.refreshCalendar(ym)
  },
  { immediate: true }
)

function dayInfo(day: number) {
  const m = points.getMonthState(mk())
  return m?.days?.[day] || { signed: false }
}

function isToday(day: number) {
  if (!isCurrentMonth()) return false
  return day === dayjs().date()
}

function isFuture(day: number) {
  if (!isCurrentMonth()) return false
  return day > dayjs().date()
}

function isPast(day: number) {
  if (!isCurrentMonth()) return false
  return day < dayjs().date()
}

function canMakeup(day: number) {
  if (!points.isCalendarReady(mk())) return false
  if (!isPast(day)) return false
  const info = dayInfo(day)
  if (info.signed) return false
  return points.makeupLeft > 0 && points.totalPoints >= 100
}

function isMissed(day: number) {
  if (!points.isCalendarReady(mk())) return false
  if (!isPast(day)) return false
  const info = dayInfo(day)
  return !info.signed
}

/**
 * ✅ 点的规则（按你最新要求）：
 * - 今天之后（含未来月份所有日期）：无点
 * - 历史/今天：没数据或没签到 -> 红点；已签到 -> 绿点
 */
function dotColor(day: number) {
  const date = monthStart.value.date(day)
  if (date.isAfter(dayjs(), 'day')) return ''

  if (!points.isCalendarReady(mk())) return 'bg-rose-400'
  const info = dayInfo(day)
  return info.signed ? 'bg-emerald-400' : 'bg-rose-400'
}

function cellClass(day: number) {
  const selected = selectedDay.value === day

  // ✅ 未来月份：整月禁用 + 变淡 + 无 hover
  if (isFutureMonth()) return 'rounded-2xl bg-transparent opacity-35 cursor-not-allowed'

  // 选中态（保持你原样）
  if (selected) return 'rounded-full bg-emerald-500'

  // 当前月未来日期禁用（保持你原样）
  if (isFuture(day)) return 'rounded-2xl bg-transparent opacity-35 cursor-not-allowed'

  return 'rounded-2xl bg-transparent hover:bg-white/5'
}

function numClass(day: number) {
  const selected = selectedDay.value === day

  // ✅ 未来月份：数字统一变淡（像你图里那样）
  if (isFutureMonth()) return 'text-white/35'

  if (selected) return 'text-white'
  if (isFuture(day)) return 'text-white/35'
  if (isToday(day)) return 'text-white'
  return 'text-white/80'
}

function pick(day: number) {
  // ✅ 未来月份/当前月未来日期：不能选中
  if (isDisabledDay(day)) return

  selectedDay.value = day
  emit('select', monthStart.value.date(day).format('YYYY-MM-DD'))
}
</script>
