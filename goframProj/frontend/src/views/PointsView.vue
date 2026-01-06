<template>
  <div class="px-4 pt-4 pb-6">
    <header class="flex items-center justify-between gap-3">
      <button
        class="h-10 w-10 rounded-xl bg-white border border-slate-200 shadow-sm flex items-center justify-center active:scale-[0.98]"
        aria-label="返回"
        @click="router.back()"
      >
        <i class="fa-solid fa-chevron-left text-slate-700"></i>
      </button>

      <div class="text-center flex-1">
        <h1 class="text-lg font-semibold leading-tight">积分明细</h1>
        <p class="text-xs text-slate-500 leading-tight">查看每笔积分变动</p>
      </div>

      <!-- ✅ 头像：变大 + 无空隙贴边框 -->
      <div class="h-12 w-12 rounded-xl bg-white border border-slate-200 shadow-sm overflow-hidden">
        <van-image
          :src="avatarUrl"
          width="100%"
          height="100%"
          fit="cover"
          class="w-full h-full"
          alt="avatar"
        />
      </div>
    </header>

    <div class="mt-4 rounded-3xl bg-slate-900 text-white shadow-sm p-4">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="text-xs text-white/70">总积分</p>
          <p class="text-2xl font-semibold mt-1 tabular-nums">{{ points.totalPoints }}</p>
          <p class="text-xs text-white/70 mt-1">当前选择：{{ cursorTitle }}</p>
        </div>

        <div class="flex items-center gap-2">
          <button
            class="h-9 w-9 rounded-xl bg-white/10 border border-white/20 flex items-center justify-center active:scale-[0.98]"
            aria-label="上一月"
            @click="prevMonth"
          >
            <i class="fa-solid fa-chevron-left text-white/90 text-sm"></i>
          </button>
          <button
            class="h-9 w-9 rounded-xl bg-white/10 border border-white/20 flex items-center justify-center active:scale-[0.98]"
            aria-label="下一月"
            @click="nextMonth"
          >
            <i class="fa-solid fa-chevron-right text-white/90 text-sm"></i>
          </button>
        </div>
      </div>

      <div class="mt-3 grid grid-cols-3 gap-2">
        <div class="rounded-2xl bg-white/10 border border-white/15 p-3">
          <p class="text-[10px] text-white/70">本月积分</p>
          <p class="text-sm font-semibold tabular-nums">{{ monthPoints }}</p>
        </div>
        <div class="rounded-2xl bg-white/10 border border-white/15 p-3">
          <p class="text-[10px] text-white/70">连签天数</p>
          <p class="text-sm font-semibold tabular-nums">{{ points.streakDays }}</p>
        </div>
        <div class="rounded-2xl bg-white/10 border border-white/15 p-3">
          <p class="text-[10px] text-white/70">剩余补签</p>
          <p class="text-sm font-semibold tabular-nums">{{ points.makeupLeft }}</p>
        </div>
      </div>
    </div>

    <div class="mt-4 flex items-center gap-2">
      <button
        class="px-3 h-9 rounded-2xl border text-sm font-semibold active:scale-[0.99]"
        :class="filter==='all' ? 'bg-slate-900 text-white border-slate-900' : 'bg-white text-slate-900 border-slate-200'"
        @click="filter='all'"
      >
        全部
      </button>
      <button
        class="px-3 h-9 rounded-2xl border text-sm font-semibold active:scale-[0.99]"
        :class="filter==='earn' ? 'bg-slate-900 text-white border-slate-900' : 'bg-white text-slate-900 border-slate-200'"
        @click="filter='earn'"
      >
        获取
      </button>
      <button
        class="px-3 h-9 rounded-2xl border text-sm font-semibold active:scale-[0.99]"
        :class="filter==='spend' ? 'bg-slate-900 text-white border-slate-900' : 'bg-white text-slate-900 border-slate-200'"
        @click="filter='spend'"
      >
        消耗
      </button>
    </div>

    <div class="mt-4 space-y-3">
      <PointsRecordItem v-for="r in filtered" :key="r.id" :record="r" />
      <div v-if="filtered.length === 0" class="rounded-3xl bg-white border border-slate-200 shadow-sm p-6 text-center">
        <div class="mx-auto h-12 w-12 rounded-2xl bg-slate-100 flex items-center justify-center">
          <i class="fa-regular fa-folder-open text-slate-600"></i>
        </div>
        <p class="mt-3 font-semibold">暂无记录</p>
        <p class="text-sm text-slate-500 mt-1">本月还没有积分变动</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import { usePointsStore } from '@/stores/points'
import { useAuthStore } from '@/stores/auth'
import PointsRecordItem from '@/components/PointsRecordItem.vue'

const router = useRouter()
const points = usePointsStore()
const auth = useAuthStore()

const avatarUrl = computed(() => {
  return auth.user?.avatar || `https://images.unsplash.com/photo-1520975693411-7a2b0d5441f4?auto=format&fit=crop&w=200&q=70`
})

const cursor = ref(dayjs().format('YYYY-MM')) // 当前选中的月份
const filter = ref<'all' | 'earn' | 'spend'>('all')

const cursorTitle = computed(() => `${cursor.value.replace('-', ' 年 ')} 月`)

const records = computed(() => points.getRecordsByMonth(cursor.value))

const filtered = computed(() => {
  if (filter.value === 'all') return records.value
  if (filter.value === 'earn') return records.value.filter(r => r.points > 0)
  return records.value.filter(r => r.points < 0)
})

const monthPoints = computed(() => {
  return records.value.reduce((s, r) => s + r.points, 0)
})

function prevMonth() {
  const d = dayjs(cursor.value + '-01').subtract(1, 'month')
  cursor.value = d.format('YYYY-MM')
  points.getMonthState(cursor.value)
}
function nextMonth() {
  const d = dayjs(cursor.value + '-01').add(1, 'month')
  cursor.value = d.format('YYYY-MM')
  points.getMonthState(cursor.value)
}
</script>
