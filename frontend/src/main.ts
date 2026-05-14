// 文件说明：frontend/src/main.ts，对应当前模块的数据结构、状态逻辑或工具函数。
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './styles.css'

createApp(App).use(createPinia()).mount('#app')
