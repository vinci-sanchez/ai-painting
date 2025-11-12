<template>
  <h2>漫画列表</h2>
  <el-container class="comic-layout">
    <el-aside width="220px" class="comic-sidebar">
      <el-card shadow="never">
        <template #header>
          <span>章节导航</span>
        </template>
        <el-menu
          class="comic-menu"
          :default-active="activePageId"
          @select="handlePageSelect"
        >
          <el-menu-item
            v-for="comicPage in comicPages"
            :key="comicPage.id"
            :index="comicPage.id.toString()"
          >
            第{{ comicPage.id }}页: {{ comicPage.title }}
          </el-menu-item>
        </el-menu>
        <el-empty
          v-if="!comicPages.length"
          description="暂无漫画页"
          class="sidebar-empty"
        />
      </el-card>
    </el-aside>

    <el-main class="comic-content">
      <el-card>
        <p class="page-title">{{ showPage.title }}</p>
        <div
          v-loading="loading"
          element-loading-text="巴拉巴拉生成中..."
          :element-loading-spinner="svg"
          element-loading-svg-view-box="-10, -10, 50, 50"
          element-loading-background="rgba(122, 122, 122, 0.8)"
          style="width: 100%"
          class="page-preview"
        >
          <el-image
            style="width: 100%; min-height: 120px; height: auto"
            :src="showPage.images || url"
            fit="cover"
          />
        </div>
        <el-button type="primary" plain @click="handlePrevPage">上一章</el-button>
        <el-button type="primary" plain @click="handleNextPage">下一章</el-button>
      </el-card>
    </el-main>
  </el-container>
</template>

<script lang="ts" setup>
import { ref ,onMounted,onBeforeUnmount} from "vue";
import type { ImageProps } from "element-plus";
import bus from "../eventBus.js";
import { comicimage,setComicImage } from "../shared-text";

type ComicPage = {
  id: number;
  title: string;
  images?: string;
};
onMounted(() => {
  console.log("开始监听漫画生成事件");
  bus.on("make-imageover", Show_comic);
});

onBeforeUnmount(() => {
  console.log("监听结束");
  bus.off("make-imageover", Show_comic);
});

const loading = ref(true);
const showPage = ref<ComicPage>({
  id: 1,
  title: "示例漫画页",
  images: "<img src='https://example.com/comic1.jpg' />",
});
const svg = `
        <path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>
      `;

const fits = [
  "fill",
  "contain",
  "cover",
  "none",
  "scale-down",
] as ImageProps["fit"][];

const url =
  "https://fuss10.elemecdn.com/e/5d/4a731a90594a4af544c0c25941171jpeg.jpeg";

const comicPages = ref<ComicPage[]>([]);
const activePageId = ref("");
const pageRefs = ref<Partial<Record<string, HTMLElement>>>({});

const setPageRef = (el: HTMLElement | null, id: number) => {
  const key = id.toString();
  if (el) {
    pageRefs.value[key] = el;
  } else {
    delete pageRefs.value[key];
  }
};

const handlePageSelect = (index: string) => {
  activePageId.value = index;
  const target = pageRefs.value[index];
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
};
//展示漫画
function Show_comic(payload: unknown) {
  const data = payload as { url: string; page_id: number; title: string };
  console.log("收到漫画数据:", data);
  const newPage: ComicPage = {
    id: data.page_id,
    title: data.title,
    images: data.url,
  };
  comicPages.value.push(newPage);
  if (comicPages.value.length === 1) {
    activePageId.value = newPage.id.toString();
    showPage.value = newPage;
  }
  loading.value = false;
  return data;
}




///////上一张
function handlePrevPage() {
  const currentIndex = comicPages.value.findIndex(
    (page) => page.id.toString() === activePageId.value
  );
  if (currentIndex > 0) {
    const prevPage = comicPages.value[currentIndex - 1];
    activePageId.value = prevPage.id.toString();
    showPage.value = prevPage;
  }
}
/////////下一张
function handleNextPage() {
  const currentIndex = comicPages.value.findIndex(
    (page) => page.id.toString() === activePageId.value
  );
  if (currentIndex < comicPages.value.length - 1) {
    const nextPage = comicPages.value[currentIndex + 1];
    activePageId.value = nextPage.id.toString();
    showPage.value = nextPage;
  }
}
</script>

<style scoped>
.comic-layout {
  min-height: 400px;
}

.comic-sidebar {
  border-right: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
}

.comic-menu {
  border-right: none;
  border-left: none;
}

.sidebar-empty {
  padding: 20px 0;
}

.comic-content {
  padding: 0 0 0 24px;
}

.comic-page-section {
  margin-bottom: 16px;
}

.page-title {
  margin: 0 0 12px;
  font-weight: 600;
}

.page-preview {
  display: flex;
  justify-content: center;
}

.demo-card {
  margin-top: 24px;
}

.demo-image .block {
  padding: 30px 0;
  text-align: center;
  border-right: solid 1px var(--el-border-color);
  display: inline-block;
  width: 20%;
  min-width: 100px;
  box-sizing: border-box;
  vertical-align: top;
}

.demo-image .block:last-child {
  border-right: none;
}

.demo-image .demonstration {
  display: block;
  color: var(--el-text-color-secondary);
  font-size: 14px;
  margin-bottom: 20px;
}
</style>
