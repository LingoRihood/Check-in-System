import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'

import './styles/tailwind.css'
import 'vant/lib/index.css'
import '@fortawesome/fontawesome-free/css/all.min.css'

import { Button, Field, Tabs, Tab, Popup, List, PullRefresh, Divider, Image as VanImage } from 'vant'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.use(Button)
app.use(Field)
app.use(Tabs)
app.use(Tab)
app.use(Popup)
app.use(List)
app.use(PullRefresh)
app.use(Divider)
app.use(VanImage)

app.mount('#app')
