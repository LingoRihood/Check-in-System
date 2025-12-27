<template>
  <div class="px-4 pt-4 pb-6">
    <header class="flex items-center justify-between">
      <div class="flex items-center gap-2">
        <div class="h-10 w-10 rounded-2xl bg-slate-900 text-white flex items-center justify-center shadow-sm">
          <i class="fa-solid fa-calendar-check"></i>
        </div>
        <div>
          <h1 class="text-lg font-semibold leading-tight">每日签到</h1>
          <p class="text-xs text-slate-500 leading-tight">登录后开始领取积分</p>
        </div>
      </div>
    </header>

    <div class="mt-4 overflow-hidden rounded-3xl bg-slate-900 shadow-sm">
      <div class="relative">
        <img
          alt="banner"
          class="h-36 w-full object-cover opacity-80"
          src="https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=1200&q=70" />
        <div class="absolute inset-0 bg-gradient-to-t from-slate-900/80 via-slate-900/20 to-transparent"></div>
        <div class="absolute bottom-3 left-3 right-3">
          <p class="text-white font-semibold">每天签到 +1</p>
          <p class="text-white/70 text-xs mt-0.5">连签还有额外奖励，补签可恢复连续</p>
        </div>
      </div>
    </div>

    <div class="mt-4 rounded-3xl bg-white border border-slate-200 shadow-sm overflow-hidden">
      <van-tabs v-model:active="tab" animated>
        <van-tab title="登录">
          <div class="p-4 space-y-3">
            <van-field v-model="loginForm.username" label="用户名" placeholder="请输入用户名" clearable />
            <van-field v-model="loginForm.password" label="密码" placeholder="请输入密码" type="password" clearable />
            <van-button
              type="primary"
              block
              class="!rounded-2xl !h-11 !bg-slate-900 !border-slate-900"
              @click="onLogin"
              :loading="loading"
            >
              登录
            </van-button>
            <p class="text-xs text-slate-500">没有账号？切换到「注册」创建账号。</p>
          </div>
        </van-tab>

        <van-tab title="注册">
          <div class="p-4 space-y-3">
            <van-field v-model="regForm.username" label="用户名" placeholder="3-20 位" clearable />
            <van-field v-model="regForm.email" label="邮箱" placeholder="请输入邮箱" clearable />
            <van-field v-model="regForm.password" label="密码" placeholder="6-20 位" type="password" clearable />
            <van-field v-model="regForm.confirmPassword" label="确认密码" placeholder="再次输入密码" type="password" clearable />
            <van-button
              type="primary"
              block
              class="!rounded-2xl !h-11 !bg-slate-900 !border-slate-900"
              @click="onRegister"
              :loading="loading"
            >
              注册
            </van-button>
            <p class="text-xs text-slate-500">注册成功后请返回「登录」。</p>
          </div>
        </van-tab>
      </van-tabs>
    </div>

    <div class="mt-4 rounded-3xl bg-white border border-slate-200 shadow-sm p-4">
      <p class="font-semibold">规则概览</p>
      <ul class="mt-2 space-y-1.5 text-sm text-slate-600">
        <li class="flex gap-2"><i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>每日签到：+1 积分</span></li>
        <li class="flex gap-2"><i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>连签奖励：3 天 +5、7 天 +10、15 天 +20</span></li>
        <li class="flex gap-2"><i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>满签奖励：当月满签 +100（演示：签满即发）</span></li>
        <li class="flex gap-2"><i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>补签：每次消耗 100 分，每月最多 3 次</span></li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const tab = ref(0)
const loading = ref(false)

const loginForm = ref({ username: '', password: '' })
const regForm = ref({ username: '', email: '', password: '', confirmPassword: '' })

async function onLogin() {
  loading.value = true
  try {
    await auth.login(loginForm.value.username.trim(), loginForm.value.password)
    const redirect = (route.query.redirect as string) || '/'
    router.replace(redirect)
  } finally {
    loading.value = false
  }
}

async function onRegister() {
  loading.value = true
  try {
    await auth.register(
      regForm.value.username.trim(),
      regForm.value.email.trim(),
      regForm.value.password,
      regForm.value.confirmPassword
    )
    tab.value = 0
  } finally {
    loading.value = false
  }
}
</script>
