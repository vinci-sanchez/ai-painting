import { defineConfig } from "vite";
//import vue from "@vitejs/plugin-vue";

export default defineConfig({
  //plugins: [vue()],
  base: "./",
  build: {
    outDir: "dist", // 确保构建到 dist
    assetsDir: "", // 可选，简化路径
  },
  server: {
    port: 5173, // 确保与 Electron 中的 URL 匹配
  },
});
