import { createRouter, createWebHistory } from "vue-router";
import login from "./pages/login.vue";
import home from "./pages/home.vue";
import signup from "./pages/signup.vue";
import { createWebHashHistory } from "vue-router";
//import Doing from "./pages/doing.vue";
// import * as VueRouter from 'vue-router';
// console.log(VueRouter); // 查看导出的内容
const routes = [
  { path: "/", component: home },
  { path: "/login", component: login },
  { path: "/signup", component: signup },
];

const router = createRouter({
  history: createWebHashHistory(),
  // history: createWebHistory(),

  routes,
});

export default router;
