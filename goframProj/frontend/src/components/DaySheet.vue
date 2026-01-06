<template>
  <van-popup
    v-model:show="innerShow"
    position="bottom"
    round
    :style="{ paddingBottom: 'env(safe-area-inset-bottom)' }"
    @close="emit('close')"
  >
    <div class="p-4">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="text-base font-semibold">{{ title }}</p>
          <p class="text-xs text-slate-500 mt-0.5">{{ sub }}</p>
        </div>
        <button
          class="h-10 w-10 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-center active:scale-[0.98]"
          aria-label="关闭"
          @click="innerShow = false"
        >
          <i class="fa-solid fa-xmark text-slate-700"></i>
        </button>
      </div>

      <div class="mt-3 rounded-2xl bg-slate-50 border border-slate-200 p-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="h-9 w-9 rounded-xl border flex items-center justify-center" :class="iconWrapClass">
              <i :class="iconClass"></i>
            </div>
            <div>
              <p class="text-sm font-semibold">{{ statusText }}</p>
              <p class="text-xs text-slate-500">{{ hintText }}</p>
            </div>
          </div>

          <p class="text-sm font-semibold tabular-nums">{{ pointsText }}</p>
        </div>
      </div>

      <div class="mt-4 grid grid-cols-2 gap-3">
        <button
          class="h-11 rounded-2xl font-semibold active:scale-[0.99] disabled:bg-slate-200 disabled:text-slate-500"
          :class="primaryBtnClass"
          :disabled="!action.canAct"
          @click="onAction"
        >
          {{ action.label }}
        </button>
        <button
          class="h-11 rounded-2xl bg-white border border-slate-200 font-semibold text-slate-900 active:scale-[0.99]"
          @click="innerShow = false"
        >
          关闭
        </button>
      </div>

      <p class="mt-3 text-xs text-slate-500">
        * 补签仅支持当月，且需要消耗 100 积分（每月最多 3 次）。
      </p>
    </div>
  </van-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { usePointsStore } from '@/stores/points'

const props = defineProps<{ show: boolean; dateStr: string }>()

const emit = defineEmits<{ (e: 'close'): void }>()

const points = usePointsStore()
const innerShow = ref(props.show)

watch(() => props.show, (v) => (innerShow.value = v))
watch(innerShow, (v) => { if (!v) emit('close') })

// ✅ 确保弹窗打开的月份日历来自后端（数据库真实数据）
watch(
  () => props.dateStr,
  (v) => {
    const ym = dayjs(v).format('YYYY-MM')
    void points.refreshCalendar(ym)
  },
  { immediate: true }
)

const d = computed(() => dayjs(props.dateStr))
const title = computed(() => `${d.value.format('YYYY 年 M 月 D 日')}`)
const sub = computed(() => (props.dateStr === dayjs().format('YYYY-MM-DD') ? '今天' : '日期详情'))

const calendarYM = computed(() => d.value.format('YYYY-MM'))
const calendarReady = computed(() => points.isCalendarReady(calendarYM.value) && !points.isCalendarLoading(calendarYM.value))

const signed = computed(() => points.isSigned(props.dateStr))
const isFuture = computed(() => d.value.isAfter(dayjs().startOf('day')))
const isCurrentMonth = computed(() => d.value.format('YYYY-MM') === dayjs().format('YYYY-MM'))
const isToday = computed(() => props.dateStr === dayjs().format('YYYY-MM-DD'))

const action = computed(() => {
  // ✅ 日历未加载完成前，不给出“补签/漏签”判断，避免误导
  if (!calendarReady.value) return { canAct: false, label: '加载中', kind: 'none' as const }
  if (signed.value) return { canAct: false, label: '已完成', kind: 'none' as const }
  if (isFuture.value) return { canAct: false, label: '未开始', kind: 'none' as const }
  if (isToday.value) return { canAct: true, label: '签到', kind: 'checkin' as const }

  if (!isCurrentMonth.value) return { canAct: false, label: '不可补签', kind: 'none' as const }
  const canMakeup = points.makeupLeft > 0 && points.totalPoints >= 100
  return canMakeup ? { canAct: true, label: '补签', kind: 'makeup' as const } : { canAct: false, label: '不可补签', kind: 'none' as const }
})

const statusText = computed(() => {
  if (signed.value) return '已签到'
  if (isFuture.value) return '未开始'
  if (isToday.value) return '可签到'
  return action.value.kind === 'makeup' ? '可补签' : '漏签'
})

const hintText = computed(() => {
  if (signed.value) return '已获得积分奖励'
  if (isFuture.value) return '日期尚未到来'
  if (isToday.value) return '签到可得 +1 积分'
  if (action.value.kind === 'makeup') return `补签将消耗 100 积分（剩余 ${points.makeupLeft} 次）`
  return points.makeupLeft === 0 ? '本月补签次数已用完' : '积分不足，无法补签'
})

const pointsText = computed(() => {
  if (signed.value) return '+1 分'
  if (isFuture.value) return '—'
  if (isToday.value) return '+1 分'
  if (action.value.kind === 'makeup') return '+1 分'
  return '0 分'
})

const iconWrapClass = computed(() => {
  if (signed.value) return 'bg-emerald-500/10 border-emerald-200'
  if (isFuture.value) return 'bg-slate-500/10 border-slate-200'
  if (isToday.value) return 'bg-slate-900 border-slate-900'
  if (action.value.kind === 'makeup') return 'bg-amber-500/10 border-amber-200'
  return 'bg-rose-500/10 border-rose-200'
})

const iconClass = computed(() => {
  if (signed.value) return 'fa-solid fa-check text-emerald-700'
  if (isFuture.value) return 'fa-regular fa-clock text-slate-600'
  if (isToday.value) return 'fa-solid fa-hand-sparkles text-white'
  if (action.value.kind === 'makeup') return 'fa-solid fa-wand-magic-sparkles text-amber-700'
  return 'fa-solid fa-xmark text-rose-700'
})

const primaryBtnClass = computed(() => (action.value.canAct ? 'bg-slate-900 text-white' : 'bg-slate-200 text-slate-500'))

function onAction() {
  if (!action.value.canAct) return
  if (action.value.kind === 'checkin') points.checkinToday()
  if (action.value.kind === 'makeup') points.makeup(props.dateStr)
  innerShow.value = false
}
</script>
