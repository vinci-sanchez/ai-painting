<template>
  <div class="home-layout">
    <aside class="home-layout__nav">
      <LeftNavigation />
    </aside>
    <main class="home-layout__content">
      <RouterView v-slot="{ Component }">
        <KeepAlive>
          <component :is="Component" />
        </KeepAlive>
      </RouterView>
    </main>
  </div>
</template>

<script setup>
import LeftNavigation from "../components/LeftNavigation.vue";
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
