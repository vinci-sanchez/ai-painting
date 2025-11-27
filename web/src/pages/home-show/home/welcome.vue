<template>
  <div class="welcome">
    <section class="welcome__hero card shadow-sm">
      <div class="welcome__hero-body">
        <p class="welcome__eyebrow">文生漫</p>
        <h1 class="welcome__title">欢迎光临！</h1>
        <p class="welcome__subtitle">
          在这里爬取/整理小说、配置参数并生成专属漫画。
        </p>
        <div class="welcome__actions">
          <el-button type="primary" @click="goTo('home-crawlcopy')">
            开始准备文稿
          </el-button>

        </div>
      </div>
      <!-- <img class="welcome__illustration" src="/img/Text-to-manga.svg" alt="text to manga" /> -->
    </section>

    <el-row :gutter="20" class="mt-4">
      <el-col :md="8" v-for="feature in features" :key="feature.title">
        <el-card shadow="hover" class="welcome__feature-card">
          <div class="welcome__feature-icon">
            <i :class="feature.icon"></i>
          </div>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.desc }}</p>
          <el-button type="text" @click="goTo(feature.target)">
            立即前往
          </el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const goTo = (name: string) => {
  router.push({ name });
};

const features = [
  {
    title: "爬取或输入小说",
    desc: "复制链接、粘贴文本或上传 TXT，快速完成内容准备。",
    icon: "fas fa-book",
    target: "home-crawlcopy",
  },
  {
    title: "智能分段",
    desc: "自动拆分段落、筛选关键剧情，保持文本结构清晰。",
    icon: "fas fa-cut",
    target: "home-segmented",
  },
  {
    title: "参数 & 漫画预览",
    desc: "调整模型参数并查看生成的漫画与历史章节。",
    icon: "fas fa-images",
    target: "home-parameter-preview",
  },
];

const steps = [
  {
    title: "1. 准备素材",
    desc: "通过“爬取/输入/上传”完成文本收集。",
  },
  {
    title: "2. 分析与分段",
    desc: "进入分段页优化段落，确认镜头脚本。",
  },
  {
    title: "3. 生成漫画",
    desc: "配置参数后生成漫画，随时在“漫画”页查看历史章节。",
  },
];
</script>

<style scoped>
.welcome {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.welcome__hero {
  display: flex;
  align-items: center;
  padding: 32px;
  border: none;
}
.welcome__hero-body {
  flex: 1;
}
.welcome__eyebrow {
  letter-spacing: 0.2em;
  text-transform: uppercase;
  margin-bottom: 8px;
  color: var(--el-color-primary);
  font-size: 12px;
}
.welcome__title {
  margin: 0 0 12px;
  font-size: 32px;
}
.welcome__subtitle {
  color: var(--el-text-color-secondary);
  margin-bottom: 24px;
  max-width: 480px;
}
.welcome__actions .el-button + .el-button {
  margin-left: 12px;
}
.welcome__illustration {
  width: 240px;
  max-width: 35%;
}
.welcome__feature-card {
  height: 100%;
}
.welcome__feature-icon {
  font-size: 32px;
  color: var(--el-color-primary);
  margin-bottom: 12px;
}
.welcome__section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.welcome__steps {
  padding-left: 18px;
  margin: 0;
  list-style: decimal;
}
.welcome__steps li {
  margin-bottom: 12px;
}
.welcome__steps h4 {
  margin: 0 0 4px;
}
</style>