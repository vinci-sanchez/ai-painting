<template> <h2>漫画列表</h2>
<el-card >
<div v-for="ComicPage in comicPages" :key="ComicPage.id">
    <el-card>第{{ ComicPage.id }}页:{{ ComicPage.title }}</el-card>
</div>
</el-card>
  <el-card>
    <div
      v-loading="loading"
      element-loading-text="巴拉巴拉生成中..."
      :element-loading-spinner="svg"
      element-loading-svg-view-box="-10, -10, 50, 50"
      element-loading-background="rgba(122, 122, 122, 0.8)"
      style="width: 100%"
      class="demo-image"
    >
      <div v-for="fit in fits" :key="fit" class="block">
        <div class="demonstration">{{ fit }}</div>
        <el-image style="width: 100px; height: 100px" :src="url" :fit="fit" />
      </div>
    </div>
  </el-card>
</template>
<script lang="ts" setup>
import { ref } from "vue";

const loading = ref(true);
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
const tableData = [
  {
    date: "2016-05-02",
    name: "John Smith",
    address: "No.1518,  Jinshajiang Road, Putuo District",
  },
  {
    date: "2016-05-04",
    name: "John Smith",
    address: "No.1518,  Jinshajiang Road, Putuo District",
  },
  {
    date: "2016-05-01",
    name: "John Smith",
    address: "No.1518,  Jinshajiang Road, Putuo District",
  },
];
import type { ImageProps } from "element-plus";

const fits = [
  "fill",
  "contain",
  "cover",
  "none",
  "scale-down",
] as ImageProps["fit"][];
const comicPages = []as Array<{id:number,title:string}>;
const url =
  "https://fuss10.elemecdn.com/e/5d/4a731a90594a4af544c0c25941171jpeg.jpeg";
</script>
<style scoped>
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
