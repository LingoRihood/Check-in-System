<template>
  <div class="rounded-3xl bg-slate-900 text-white shadow-sm overflow-hidden border border-white/10">
    <!-- Header -->
    <!-- Header：包括年份、月份以及“上一月”和“下一月”按钮。用户可以通过点击这些按钮来切换月份。 -->
    <!-- px-4：为左右添加了 4 单位的内边距，保证内容不会紧贴容器的两侧 -->
    <!-- gap-3：子元素之间的间距设置为 3 单位 -->
    <!-- items-center：将子元素沿纵轴（垂直方向）居中对齐 -->
    <div class="px-4 pt-4 pb-3 flex items-center justify-between gap-3">
      <div class="min-w-0">
        <!-- <p class="text-xs text-white/60">当前月签到详情</p> -->
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

    <!-- Week header (Sun..Sat) -->
    <!-- 显示日历的星期（Sun 到 Sat）。这部分是一个 grid 布局，每个星期显示一个英文缩写并居中对齐 -->
    <div class="px-3 pb-2">
      <!-- grid-cols-7：设置网格有 7 列（对应 7 天的星期），每列均等分配宽度 -->
      <!-- text-[13px]：设置文本的大小为 13 像素。
      font-semibold：设置字体为 半粗体，让文字更加突出。
      text-white/70：设置文本颜色为 白色，透明度为 70%，即文本不完全是纯白色，有一点透明感，增加设计感。 -->
      <div class="grid grid-cols-7 text-[15px] font-semibold text-white/70">
        <!-- h-10：设置每个单元格的高度为 10 单位，确保单元格高度一致。
        flex：启用 flexbox 布局，使得每个 div 内的内容（即星期名称）能够居中对齐。
        flex-col：使 flex 布局沿着纵轴排列内容，即竖直方向排列。
        items-center：使得每个 div 内的内容水平居中对齐。
        justify-center：使得每个 div 内的内容垂直居中对齐 -->
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
    <!-- 使用 grid 布局排列每个月的日期，leadingBlanks 用来计算前面未显示日期的空白格 -->
    <div class="px-3 pb-4">
      <!-- gap-1.5：每个格子的间距
      aria-label：无障碍属性，告诉读屏这是“日历签到网格” -->
      <!-- place-items-center: 相当于 justify-items: center;（每个格子的内容水平居中）align-items: center;（每个格子的内容垂直居中）这样每个格子里的 button 都会居中，列中心就对了 -->
      <!-- gap-y-3：表示每行之间的间距为 0.75rem（12px） -->
      <div class="grid grid-cols-7 gap-y-3 place-items-center" aria-label="日历签到网格">
        <!-- leading blanks -->
        <!-- 月初对齐用的“空格占位” -->
        <!-- 当这个月 1 号不是周日时，需要在前面补空格占位，保证 1 号出现在正确星期列。 -->
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
          :disabled="isFuture(day)"
          @click="pick(day)"
        >
          <!-- iOS-like selection / today ring (perfect circle, not affected by w-full) -->
          <div class="absolute inset-0 flex items-center justify-center pointer-events-none">
            <!-- Selected: filled circle around number -->
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

          <!-- dot placeholder (fixed height so alignment never drifts) -->
          <div class="relative mt-1 h-1.5 flex items-center justify-center">
            <span
              v-if="dotColor(day)"
              class="h-1.5 w-1.5 rounded-full"
              :class="dotColor(day)"
            ></span>
          </div>
        </button>
      </div>

      <!-- Legend -->
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

// Set the default selected day to today's date
const selectedDay = ref<number | null>(dayjs().date())

/**
 * 对齐关键：采用 Sun..Sat 视图
 * monthStart.day(): Sun=0..Sat=6，直接作为 leading blanks
 */
const leadingBlanks = computed(() => monthStart.value.day())

watch(
  () => [props.year, props.monthIndex],
  () => (selectedDay.value = dayjs().date()) // Ensure selectedDay is always today's date when month changes
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

function mk() {
  return monthStart.value.format('YYYY-MM')
}

function isCurrentMonth() {
  return mk() === dayjs().format('YYYY-MM')
}

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
  // ✅ 日历未从后端加载完成时，不做任何“可补签/漏签”推断（避免本地误标）
  if (!points.isCalendarReady(mk())) return false
  // 仅当月过去漏签才有意义
  if (!isPast(day)) return false
  const info = dayInfo(day)
  if (info.signed) return false
  return points.makeupLeft > 0 && points.totalPoints >= 100
}

function isMissed(day: number) {
  // ✅ 日历未从后端加载完成时，不做任何“漏签”推断（避免本地误标）
  if (!points.isCalendarReady(mk())) return false
  // 仅当月过去且未签
  if (!isPast(day)) return false
  const info = dayInfo(day)
  return !info.signed
}

/**
 * dot 颜色映射（参考你图的“每天一个点”）
 * - 已签到：绿点
 * - 可补签：黄点
 * - 漏签：红点
 * - 今天未签：灰点
 * - 未来：无点
 */
function dotColor(day: number) {
  const info = dayInfo(day)
  if (isFuture(day)) return ''
  if (info.signed) return 'bg-emerald-400'
  if (canMakeup(day)) return 'bg-amber-400'
  if (isMissed(day)) return 'bg-rose-400'
  if (isToday(day) && !info.signed) return 'bg-white/60'
  return ''
}

/**
 * cell 视觉：
 * - 选中：绿色实心圆（像你图的 27）
 * - 未来：禁用 + 变淡
 * - 其它：透明，hover 轻微高亮（桌面也好看）
 */
function cellClass(day: number) {
  const selected = selectedDay.value === day
  if (selected) return 'rounded-full bg-emerald-500'
  if (isFuture(day)) return 'rounded-2xl bg-transparent opacity-35 cursor-not-allowed'
  return 'rounded-2xl bg-transparent hover:bg-white/5'
}

function numClass(day: number) {
  const selected = selectedDay.value === day
  if (selected) return 'text-white'
  if (isFuture(day)) return 'text-white/35'
  if (isToday(day)) return 'text-white' // 今天稍亮
  return 'text-white/80'
}

function pick(day: number) {
  if (isFuture(day)) return
  selectedDay.value = day
  emit('select', monthStart.value.date(day).format('YYYY-MM-DD'))
}
</script>
