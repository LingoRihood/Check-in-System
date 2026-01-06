<template>
  <div class="rounded-3xl bg-white border border-slate-200 shadow-sm p-4 flex items-start justify-between gap-3">
    <div class="flex items-start gap-3 min-w-0">
      <div class="h-11 w-11 rounded-2xl flex items-center justify-center shrink-0" :class="meta.bg">
        <i :class="meta.icon + ' ' + meta.color"></i>
      </div>
      <div class="min-w-0">
        <p class="font-semibold leading-tight truncate">{{ record.title }}</p>
        <p class="text-xs text-slate-500 mt-1">{{ record.date }}</p>
      </div>
    </div>

    <div class="text-right shrink-0">
      <p class="text-base font-semibold tabular-nums">
        {{ record.points > 0 ? `+${record.points}` : record.points }} 分
      </p>
      <p class="text-xs text-slate-500 mt-1">{{ record.points > 0 ? '获取' : '消耗' }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { PointsRecord } from '@/stores/points' // ✅ 改成从 store 导入（你现在真实数据的类型）

const props = defineProps<{ record: PointsRecord }>()

const meta = computed(() => {
  // ✅ 优先用 store 里新增的 bizType（最稳定）
  const bt = (props.record as any)?.bizType as string | undefined
  if (bt === 'checkin') {
    return { icon: 'fa-solid fa-calendar-check', bg: 'bg-emerald-500/10', color: 'text-emerald-700' }
  }
  if (bt === 'bonus') {
    return { icon: 'fa-solid fa-fire', bg: 'bg-amber-500/10', color: 'text-amber-700' }
  }
  if (bt === 'makeup_cost') {
    return { icon: 'fa-solid fa-receipt', bg: 'bg-rose-500/10', color: 'text-rose-700' }
  }

  // ✅ 兜底 1：如果没有 bizType，但 store 里保留了 transactionType（后端字段）
  const txType = Number((props.record as any)?.transactionType ?? 0)
  if (txType === 1) {
    return { icon: 'fa-solid fa-calendar-check', bg: 'bg-emerald-500/10', color: 'text-emerald-700' }
  }
  if (txType === 2) {
    return { icon: 'fa-solid fa-fire', bg: 'bg-amber-500/10', color: 'text-amber-700' }
  }
  if (txType === 3) {
    return { icon: 'fa-solid fa-receipt', bg: 'bg-rose-500/10', color: 'text-rose-700' }
  }

  // ✅ 兜底 2：再退一步，按 title 关键字（不依赖后端字段也能恢复大部分）
  const title = String(props.record.title || '')
  if (title.includes('每日签到')) {
    return { icon: 'fa-solid fa-calendar-check', bg: 'bg-emerald-500/10', color: 'text-emerald-700' }
  }
  if (title.includes('连续签到')) {
    return { icon: 'fa-solid fa-fire', bg: 'bg-amber-500/10', color: 'text-amber-700' }
  }
  if (title.includes('补签') && title.includes('消耗')) {
    return { icon: 'fa-solid fa-receipt', bg: 'bg-rose-500/10', color: 'text-rose-700' }
  }

  // ✅ 最后兜底：保持原来 default 的 +
  return { icon: 'fa-solid fa-plus', bg: 'bg-slate-500/10', color: 'text-slate-700' }
})
</script>
