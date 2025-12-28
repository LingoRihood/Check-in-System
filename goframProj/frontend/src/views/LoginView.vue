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
          src="https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=1200&q=70"
        />
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
            <!-- ✅ 登录用户名：右侧 × 真清除（不再用 clearable） -->
            <van-field
              ref="loginUserField"
              v-model="loginForm.username"
              class="input-wide"
              label="用户名"
              placeholder="请输入用户名"
              autocomplete="username"
            >
              <template #right-icon>
                <van-icon
                  v-if="loginForm.username"
                  name="clear"
                  size="18"
                  class="text-slate-400"
                  @mousedown.prevent
                  @click="clearField('loginUsername')"
                />
              </template>
            </van-field>

            <!-- ✅ 登录密码：不切 type（永远 text），用 CSS mask，光标不跳 -->
            <van-field
              v-model="loginForm.password"
              label="密码"
              placeholder="请输入密码"
              type="text"
              autocomplete="current-password"
              :class="['input-wide', 'pwd-field', { 'pwd-masked': !showLoginPwd }]"
            >
              <template #right-icon>
                <van-icon
                  :name="showLoginPwd ? 'eye-o' : 'closed-eye'"
                  size="18"
                  class=" text-slate-500"
                  @mousedown.prevent
                  @click="showLoginPwd = !showLoginPwd"
                />
              </template>
            </van-field>

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
            <!-- ✅ 注册用户名：右侧 × 真清除 -->
            <van-field
              ref="regUserField"
              v-model="regForm.username"
              class="input-wide"
              label="用户名"
              placeholder="3-20 位"
              autocomplete="username"
            >
              <template #right-icon>
                <van-icon
                  v-if="regForm.username"
                  name="clear"
                  size="18"
                  class=" text-slate-400"
                  @mousedown.prevent
                  @click="clearField('regUsername')"
                />
              </template>
            </van-field>

            <!-- ✅ 邮箱：右侧 × 真清除 -->
            <van-field
              ref="regEmailField"
              v-model="regForm.email"
              class="input-wide"
              label="邮箱"
              placeholder="请输入邮箱"
              autocomplete="email"
            >
              <template #right-icon>
                <van-icon
                  v-if="regForm.email"
                  name="clear"
                  size="18"
                  class=" text-slate-400"
                  @mousedown.prevent
                  @click="clearField('regEmail')"
                />
              </template>
            </van-field>

            <!-- ✅ 注册密码：CSS mask -->
            <van-field
              v-model="regForm.password"
              label="密码"
              placeholder="6-20 位"
              type="text"
              autocomplete="new-password"
              :class="['input-wide', 'pwd-field', { 'pwd-masked': !showRegPwd }]"
            >
              <template #right-icon>
                <van-icon
                  :name="showRegPwd ? 'eye-o' : 'closed-eye'"
                  size="18"
                  class=" text-slate-500"
                  @mousedown.prevent
                  @click="showRegPwd = !showRegPwd"
                />
              </template>
            </van-field>

            <!-- ✅ 确认密码：CSS mask -->
            <van-field
              v-model="regForm.confirmPassword"
              label="确认密码"
              placeholder="再次输入密码"
              type="text"
              autocomplete="new-password"
              :class="['input-wide', 'pwd-field', { 'pwd-masked': !showRegConfirmPwd }]"
            >
              <template #right-icon>
                <van-icon
                  :name="showRegConfirmPwd ? 'eye-o' : 'closed-eye'"
                  size="18"
                  class=" text-slate-500"
                  @mousedown.prevent
                  @click="showRegConfirmPwd = !showRegConfirmPwd"
                />
              </template>
            </van-field>

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
        <li class="flex gap-2">
          <i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>连签奖励：3 天 +5、7 天 +10、15 天 +20</span>
        </li>
        <li class="flex gap-2">
          <i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>满签奖励：当月满签 +100（演示：签满即发）</span>
        </li>
        <li class="flex gap-2">
          <i class="fa-solid fa-check text-emerald-600 mt-1"></i><span>补签：每次消耗 100 分，每月最多 3 次</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const tab = ref(0)
const loading = ref(false)

const loginForm = ref({ username: '', password: '' })
const regForm = ref({ username: '', email: '', password: '', confirmPassword: '' })

// ✅ 控制“是否显示明文”
const showLoginPwd = ref(false)
const showRegPwd = ref(false)
const showRegConfirmPwd = ref(false)

// ✅ 为了清除后仍保持焦点（可选，但体验更好）
const loginUserField = ref<any>(null)
const regUserField = ref<any>(null)
const regEmailField = ref<any>(null)

function focusField(fieldRef: any) {
  const root = fieldRef?.value?.$el as HTMLElement | undefined
  const input = root?.querySelector('input') as HTMLInputElement | null
  input?.focus()
}

async function clearField(which: 'loginUsername' | 'regUsername' | 'regEmail') {
  if (which === 'loginUsername') loginForm.value.username = ''
  if (which === 'regUsername') regForm.value.username = ''
  if (which === 'regEmail') regForm.value.email = ''

  await nextTick()
  if (which === 'loginUsername') focusField(loginUserField)
  if (which === 'regUsername') focusField(regUserField)
  if (which === 'regEmail') focusField(regEmailField)
}

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
    const ok = await auth.register(
      regForm.value.username.trim(),
      regForm.value.email.trim(),
      regForm.value.password,
      regForm.value.confirmPassword
    )
    if (ok) tab.value = 0
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* ✅ 在 Chrome/Edge/Safari 上把 text 输入“伪装成密码圆点”，不改变 DOM，不会跳光标 */
.pwd-masked :deep(.van-field__control) {
  -webkit-text-security: disc;
}

/* ✅ 所有输入框：加大字符间隔 */
.input-wide :deep(.van-field__control) {
  letter-spacing: 0.08em;
}

/* ✅ 强制修正：输入区用文本光标，其他区域用默认光标（避免手型/grab） */
:deep(.van-field) {
  cursor: default !important;
}

:deep(.van-field__control),
:deep(.van-field__control input),
:deep(.van-field__control textarea),
:deep(input),
:deep(textarea) {
  cursor: text !important;
}

/* 右侧图标区域如果你也不想手型 */
:deep(.van-field__right-icon),
:deep(.van-field__button),
:deep(.van-icon) {
  cursor: default !important;
}
</style>
