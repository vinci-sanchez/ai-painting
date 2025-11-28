<template>
  <div
    class="sidebar"
    :class="{ 'sidebar--collapsed': isCollapse }"
    @click="handleSidebarClick"
  >
    <el-menu
      :default-active="activeMenu"
      class="el-menu-vertical-demo sidebar-menu"
      :mode="menuMode"
      :collapse="isCollapse"
      @open="handleOpen"
      @close="handleClose"
      @select="handleSelect"
      style="height: 100%"
    >
      <el-menu-item index="home-welcome">
        <el-icon><Compass /></el-icon>
        <template #title>仪表盘</template>
      </el-menu-item>
      <el-menu-item index="home-crawlcopy">
        <el-icon><DocumentCopy /></el-icon>
        <template #title>爬取或上传</template>
      </el-menu-item>
      <el-menu-item index="home-segmented">
        <el-icon><icon-menu /></el-icon>
        <template #title>分段</template>
      </el-menu-item>
      <el-menu-item index="home-parameter-preview">
        <el-icon><document /></el-icon>
        <template #title>参数预览</template>
      </el-menu-item>
      <el-menu-item index="home-comic">
        <el-icon><Collection /></el-icon>
        <template #title>漫画</template>
      </el-menu-item>
    </el-menu>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import {
  Document,
  Menu as IconMenu,
  Location,
  Setting,
  HomeFilled,
} from "@element-plus/icons-vue";
import { useRoute, useRouter } from "vue-router";

const router = useRouter();
const route = useRoute();

const isCollapse = ref(true);
const menuMode = computed<"horizontal" | "vertical">(() => "vertical");
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

const activeMenu = computed(() => {
  const name = route.name;
  return typeof name === "string" ? name : "home-welcome";
});

const handleSelect = (index: string) => {
  if (index === activeMenu.value) {
    return;
  }
  router.push({ name: index });
};
</script>

<style>
.sidebar {
  height: 100%;
  width: auto; /*有点bug*/
  display: flex;
  background-color: var(--el-menu-bg-color, transparent);
  transition: width 0.2s ease;
  padding: 8px;
}

.sidebar-menu {
  flex: 0 0 auto;
  border-radius: 16px;
  box-shadow: 0 12px 30px rgba(20, 30, 68, 0.08);
  overflow: hidden;
}

.el-menu-vertical-demo:not(.el-menu--collapse) {
  width: 200px;
  min-height: 400px;
}
</style>
