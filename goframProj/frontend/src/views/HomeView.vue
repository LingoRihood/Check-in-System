<template>
  <div class="min-h-screen bg-slate-50">
    <div class="mx-auto max-w-md px-4 pt-4 pb-28">
      <!-- Header -->
      <header class="flex items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <div class="h-10 w-10 rounded-2xl bg-slate-900 text-white flex items-center justify-center shadow-sm">
            <i class="fa-solid fa-calendar-check"></i>
          </div>
          <div>
            <h1 class="text-lg font-semibold leading-tight">每日签到</h1>
            <p class="text-xs text-slate-500 leading-tight">坚持打卡，积分每天涨</p>
          </div>
        </div>

        <!-- 用户头像（点击可退出登录） -->
        <button
          class="h-10 w-10 rounded-full bg-white border border-slate-200 shadow-sm flex items-center justify-center overflow-hidden active:scale-[0.98]"
          aria-label="用户头像，点击退出登录"
          @click="onLogout"
        >
          <!-- 确保头像容器与图片大小一致，没有空隙 -->
          <van-image :src="avatarUrl" width="48" height="48" round fit="cover" alt="avatar" />
        </button>
      </header>

      <!-- Banner -->
      <div class="mt-4 overflow-hidden rounded-3xl bg-slate-900 shadow-sm">
        <div class="relative">
          <img
            alt="签到积分主题图"
            class="h-44 w-full object-cover opacity-85"
            src="https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=1200&q=70"
          />
          <div class="absolute inset-0 bg-gradient-to-t from-slate-900/85 via-slate-900/25 to-transparent"></div>

          <!-- ✅ 改造：label 单独一行，不参与左右对齐 -->
          <div class="absolute bottom-3 left-3 right-3">
            <p class="text-white/80 text-xs">当前账号</p>

            <!-- ✅ 这一行只负责左右对齐：左玻璃条 + 右徽章 -->
            <div class="mt-2 flex items-stretch justify-between gap-3">
              <!-- 左侧玻璃条（固定高度，确保和右侧齐） -->
              <div
                class="flex-1 min-w-0 h-16 flex items-center gap-3 rounded-2xl bg-slate-900/35 border border-white/10 px-3 backdrop-blur"
              >
                <!-- ✅ 头像容器：强制居中 -->
                <div
                  class="h-12 w-12 rounded-full overflow-hidden shrink-0 ring-1 ring-white/15 flex items-center justify-center"
                >
                  <van-image :src="avatarUrl" width="56" height="56" round fit="cover" class="block" />
                </div>

                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2 min-w-0">
                    <p class="text-white font-semibold leading-tight truncate">
                      {{ auth.user?.username || '-' }}
                    </p>

                    <span
                      class="shrink-0 inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
                      :class="isTodaySigned ? 'bg-emerald-500/15 text-emerald-100' : 'bg-white/10 text-white/80'"
                    >
                      <i :class="isTodaySigned ? 'fa-regular fa-circle-check' : 'fa-regular fa-clock'"></i>
                      {{ isTodaySigned ? '已签到' : '未签到' }}
                    </span>
                  </div>

                  <p class="mt-1 text-white/70 text-xs truncate">
                    {{ isTodaySigned ? '保持连签可触发奖励' : '今天签到可得 +1 积分' }}
                  </p>
                </div>
              </div>

              <!-- 右侧徽章（固定高度 + 固定宽度 + 垂直居中） -->
              <div
                class="h-16 w-[84px] shrink-0 rounded-2xl bg-white/10 border border-white/15 px-3 backdrop-blur flex flex-col justify-center"
              >
                <p class="text-[10px] text-white/70 leading-none text-center">本月累计</p>
                <p class="mt-1 text-white text-sm font-semibold tabular-nums leading-none text-center">
                  {{ points.monthPoints }} 分
                </p>
              </div>
            </div>
          </div>
          <!-- ✅ Banner 结束 -->
        </div>
      </div>

      <!-- Summary cards -->
      <div class="mt-4 grid grid-cols-12 gap-3">
        <!-- Total points -->
        <div class="col-span-7 rounded-3xl bg-white border border-slate-200 shadow-sm p-4">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-xs text-slate-500">总积分</p>
              <p class="text-2xl font-semibold mt-1 tabular-nums">{{ points.totalPoints }}</p>

              <button
                class="mt-2 inline-flex items-center gap-2 text-sm text-slate-900 font-medium active:scale-[0.99]"
                @click="router.push('/points')"
              >
                查看明细
                <i class="fa-solid fa-chevron-right text-xs text-slate-500"></i>
              </button>
            </div>

            <div class="h-10 w-10 rounded-2xl bg-slate-100 flex items-center justify-center">
              <i class="fa-solid fa-coins text-amber-500"></i>
            </div>
          </div>
        </div>

        <!-- Month stats -->
        <div class="col-span-5 rounded-3xl bg-slate-900 text-white shadow-sm p-4">
          <p class="text-xs text-white/70">本月积分</p>
          <p class="text-xl font-semibold mt-1 tabular-nums">{{ points.monthPoints }}</p>

          <div class="mt-2 flex items-center gap-2 text-xs text-white/80">
            <i class="fa-solid fa-fire"></i>
            <span>连签 <span class="font-semibold tabular-nums">{{ points.streakDays }}</span> 天</span>
          </div>

          <div class="mt-1 flex items-center gap-2 text-xs text-white/80">
            <i class="fa-solid fa-wand-magic-sparkles"></i>
            <span>补签 <span class="font-semibold tabular-nums">{{ points.makeupLeft }}</span> 次</span>
          </div>
        </div>
      </div>

      <!-- Calendar -->
      <div class="mt-4">
        <MonthCalendar
          :year="cursor.year"
          :monthIndex="cursor.monthIndex"
          @select="openDaySheet"
          @prev="prevMonth"
          @next="nextMonth"
        />
      </div>

      <!-- Rules -->
      <div class="mt-4 rounded-3xl bg-white border border-slate-200 shadow-sm p-4">
        <div class="flex items-start gap-3">
          <div class="h-10 w-10 rounded-2xl bg-slate-100 flex items-center justify-center">
            <i class="fa-solid fa-lightbulb text-slate-700"></i>
          </div>
          <div class="min-w-0">
            <p class="font-semibold">规则与激励</p>
            <ul class="mt-2 space-y-1.5 text-sm text-slate-600">
              <li class="flex gap-2">
                <i class="fa-solid fa-check text-emerald-600 mt-1"></i>
                <span>每日签到固定奖励：<span class="font-medium text-slate-900">+1</span> 分</span>
              </li>
              <li class="flex gap-2">
                <i class="fa-solid fa-check text-emerald-600 mt-1"></i>
                <span>连签加成：第 3/7/15 天额外 <span class="font-medium text-slate-900">+5/+10/+20</span> 分</span>
              </li>
              <li class="flex gap-2">
                <i class="fa-solid fa-check text-emerald-600 mt-1"></i>
                <span>补签：每次消耗 <span class="font-medium text-slate-900">100</span> 分，每月最多 <span class="font-medium text-slate-900">3</span> 次</span>
              </li>
              <li class="flex gap-2">
                <i class="fa-solid fa-check text-emerald-600 mt-1"></i>
                <span>本月满签：额外 <span class="font-medium text-slate-900">+100</span> 分（演示：签满即发）</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <DaySheet :show="sheet.show" :dateStr="sheet.dateStr" @close="sheet.show = false" />
    </div>

    <!-- Bottom CTA -->
    <div class="fixed inset-x-0 bottom-0 pb-[calc(env(safe-area-inset-bottom)+12px)] pt-3 bg-gradient-to-t from-slate-50 via-slate-50 to-slate-50/0">
      <div class="mx-auto max-w-md px-4">
        <div class="rounded-3xl bg-white border border-slate-200 shadow-sm p-3 flex items-center gap-3">
          <div class="h-11 w-11 rounded-2xl bg-slate-900 text-white flex items-center justify-center shrink-0">
            <i class="fa-solid fa-gift"></i>
          </div>

          <div class="min-w-0 flex-1">
            <p class="text-sm font-semibold leading-tight">{{ isTodaySigned ? '今天已签到' : '今天还没签到' }}</p>
            <p class="text-xs text-slate-500 leading-tight">
              {{ isTodaySigned ? '保持连签可触发奖励' : '签到可得 +1 积分，连签还有额外奖励' }}
            </p>
          </div>

          <button
            class="h-11 px-4 rounded-2xl font-semibold active:scale-[0.99] disabled:bg-slate-200 disabled:text-slate-500"
            :class="isTodaySigned ? 'bg-slate-200 text-slate-500' : 'bg-slate-900 text-white'"
            :disabled="isTodaySigned"
            @click="points.checkinToday()"
          >
            {{ isTodaySigned ? '已签到' : '立即签到' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, computed, onMounted } from 'vue'
import dayjs from 'dayjs'
import { useRouter } from 'vue-router'
import { showConfirmDialog } from 'vant'

import { useAuthStore } from '@/stores/auth'
import { usePointsStore } from '@/stores/points'
import MonthCalendar from '@/components/MonthCalendar.vue'
import DaySheet from '@/components/DaySheet.vue'

const router = useRouter()
const auth = useAuthStore()
const points = usePointsStore()

const cursor = reactive({
  year: dayjs().year(),
  monthIndex: dayjs().month()
})

const sheet = reactive({
  show: false,
  dateStr: dayjs().format('YYYY-MM-DD')
})

const isTodaySigned = computed(() => points.isSigned(dayjs().format('YYYY-MM-DD')))

const avatarUrl = computed(() => {
  return (
    auth.user?.avatar ||
    'https://images.unsplash.com/photo-1520975693411-7a2b0d5441f4?auto=format&fit=crop&w=200&q=70'
  )
})

function monthKey(y: number, monthIndex: number) {
  return `${y}-${String(monthIndex + 1).padStart(2, '0')}`
}

function prevMonth() {
  cursor.monthIndex -= 1
  if (cursor.monthIndex < 0) {
    cursor.monthIndex = 11
    cursor.year -= 1
  }
  points.getMonthState(monthKey(cursor.year, cursor.monthIndex))
}

function nextMonth() {
  cursor.monthIndex += 1
  if (cursor.monthIndex > 11) {
    cursor.monthIndex = 0
    cursor.year += 1
  }
  points.getMonthState(monthKey(cursor.year, cursor.monthIndex))
}

function openDaySheet(dateStr: string) {
  sheet.dateStr = dateStr
  sheet.show = true
}

async function onLogout() {
  try {
    await showConfirmDialog({ title: '退出登录', message: '确定要退出当前账号吗？' })
    auth.logout()
    router.replace('/login')
  } catch {
    // cancel
  }
}

onMounted(() => {
  points.getMonthState(monthKey(cursor.year, cursor.monthIndex))
})
</script>
