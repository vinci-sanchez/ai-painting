import { createApp } from 'vue';
//import Vue from 'vue';
// import ElementUI from 'element-ui';
// import 'element-ui/lib/theme-chalk/index.css';
import ElementPlus from 'element-plus';   // 2. 导入 Element Plus
import 'element-plus/dist/index.css';
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue';
import router from './router';


const app = createApp(App);
app.use(router);
app.use(ElementPlus); // 注册 Element Plus
app.mount('#app');
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}