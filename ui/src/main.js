import {createApp} from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import VueAxios from 'vue-axios'
import App from './App.vue'
import { createPinia } from 'pinia'

const app = createApp(App)
app.use(ElementPlus).use(VueAxios).use(createPinia())
app.mount('#app')
