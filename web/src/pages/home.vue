<template>
  <div class="home-layout">
    <aside class="home-layout__nav">
      <LeftNavigation />
    </aside>
    <main class="home-layout__content">
      <CrawlCopy />
      <Segmented />
    </main>
  </div>
</template>

<script setup>
import CrawlCopy from "./home-show/crawl-or-copy/crawlcopy.vue";
import LeftNavigation from "../components/LeftNavigation.vue";
import Segmented from "./home-show/Segmented/Segmented.vue";

</script>

<style scoped>
.home-layout {
  display: flex; /* 左右排列 */
  height: 100%; /* 视需求设置高度 */
  
}

.home-layout__nav {
  width: auto; /* 左侧固定宽度，可按需要调整 */
  flex-shrink: 0;
}

.home-layout__content {
  flex: 1; /* 右侧撑满剩余空间 */
  overflow: auto; /* 内容溢出时滚动 */
  padding: 16px; /* 可选的内边距 */
}
</style>
