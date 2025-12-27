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
import type { PointsRecord } from '@/utils/pointsStorage'

const props = defineProps<{ record: PointsRecord }>()

const meta = computed(() => {
  switch (props.record.type) {
    case 'checkin':
      return { icon: 'fa-solid fa-calendar-check', bg: 'bg-emerald-500/10', color: 'text-emerald-700' }
    case 'bonus':
      return { icon: 'fa-solid fa-fire', bg: 'bg-amber-500/10', color: 'text-amber-700' }
    case 'makeup_cost':
      return { icon: 'fa-solid fa-receipt', bg: 'bg-rose-500/10', color: 'text-rose-700' }
    default:
      return { icon: 'fa-solid fa-plus', bg: 'bg-slate-500/10', color: 'text-slate-700' }
  }
})
</script>
