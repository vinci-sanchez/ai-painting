<template>
  <div
    class="sidebar"
    :class="{ 'sidebar--collapsed': isCollapse }"
    @click="handleSidebarClick"
  >
    <el-menu
      default-active="2"
      class="el-menu-vertical-demo sidebar-menu"
      :collapse="isCollapse"
      @open="handleOpen"
      @close="handleClose"
      style="height: 100vh"
    >
      <el-menu-item index="1">
        <el-icon><location /></el-icon>
        <template #title>crawl-or-copy</template>
      </el-menu-item>
      <el-menu-item index="2">
        <el-icon><icon-menu /></el-icon>
        <template #title>Segmented</template>
      </el-menu-item>
      <el-menu-item index="3">
        <el-icon><document /></el-icon>
        <template #title>parameter-preview</template>
      </el-menu-item>
      <el-menu-item index="4">
        <el-icon><setting /></el-icon>
        <template #title>comic</template>
      </el-menu-item>
    </el-menu>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import {
  Document,
  Menu as IconMenu,
  Location,
  Setting,
} from "@element-plus/icons-vue";

const isCollapse = ref(true);
const handleOpen = (key: string, keyPath: string[]) => {
  console.log(key, keyPath);
};
const handleClose = (key: string, keyPath: string[]) => {
  console.log(key, keyPath);
};
const handleSidebarClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  if (target.closest(".el-menu-item, .el-sub-menu__title")) {
    return;
  }
  isCollapse.value = !isCollapse.value;
};
</script>

<style>
.sidebar {
  height: 100vh;
  width: auto; /*有点bug*/
  display: flex;
  background-color: var(--el-menu-bg-color, transparent);
  transition: width 0.2s ease;
}

/* .sidebar--collapsed {
  width: 64px;
} */

.sidebar-menu {
  flex: 0 0 auto;
}

.el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 200px;
  min-height: 400px;
}
</style>
