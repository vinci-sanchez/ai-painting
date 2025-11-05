import { createRouter, createWebHashHistory } from "vue-router";
import login from "./pages/login.vue";
import home from "./pages/home.vue";
import signup from "./pages/signup.vue";
import CrawlCopy from "./pages/home-show/crawl-or-copy/crawlcopy.vue";
import Segmented from "./pages/home-show/Segmented/Segmented.vue";
import ParameterPreview from "./pages/home-show/parameter-preview/parameter-preview.vue";
import Comic from "./pages/home-show/comic/comic.vue";

const routes = [
  {
    path: "/",
    component: home,
    redirect: { name: "home-crawlcopy" },
    children: [
      { path: "crawlcopy", name: "home-crawlcopy", component: CrawlCopy },
      { path: "segmented", name: "home-segmented", component: Segmented },
      {
        path: "parameter-preview",
        name: "home-parameter-preview",
        component: ParameterPreview,
      },
      { path: "comic", name: "home-comic", component: Comic },
    ],
  },
  { path: "/login", component: login },
  { path: "/signup", component: signup },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
