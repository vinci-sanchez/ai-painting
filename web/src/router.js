import { createRouter, createWebHistory } from "vue-router";
import Home from "./pages/Home.vue";
import about from "./pages/about.vue";
//import use from "./pages/use.vue";
//import Doing from "./pages/doing.vue";
import { createWebHashHistory } from "vue-router";
// import * as VueRouter from 'vue-router';
// console.log(VueRouter); // 查看导出的内容
const routes = [
  { path: "/", component: Home },
  { path: "/doing", component: Doing },
  { path: "/use", component: use },
];

const router = createRouter({
  history: createWebHashHistory(),
  // history: createWebHistory(),

  routes,
});

export default router;
