import { createRouter, createWebHashHistory } from "vue-router";
import login from "./pages/login.vue";
import home from "./pages/home.vue";
import signup from "./pages/signup.vue";
import CrawlCopy from "./pages/home-show/crawl-or-copy/crawlcopy.vue";
import Segmented from "./pages/home-show/Segmented/Segmented.vue";
import ParameterPreview from "./pages/home-show/parameter-preview/parameter-preview.vue";
import Comic from "./pages/home-show/comic/comic.vue";
import Welcome from "./pages/home-show/home/welcome.vue";
import ReadBook from "./pages/home-show/11/ReadBook.vue";
import SampleGallery from "./pages/home-show/sample/sample-gallery.vue";
import AdminPanel from "./pages/home-show/admin/admin-panel.vue";

const routes = [
  {
    path: "/",
    component: home,
    redirect: { name: "home-welcome" },
    children: [
      { path: "welcome", name: "home-welcome", component: Welcome },
      { path: "crawlcopy", name: "home-crawlcopy", component: CrawlCopy },
      { path: "segmented", name: "home-segmented", component: Segmented },
      {
        path: "parameter-preview",
        name: "home-parameter-preview",
        component: ParameterPreview,
      },
      { path: "comic", name: "home-comic", component: Comic },
      { path: "sample", name: "home-sample-gallery", component: SampleGallery },
      { path: "11", name: "home-readbook", component: ReadBook },
      { path: "111", name: "home-admin", component: AdminPanel },
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
