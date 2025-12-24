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
      <el-col
        v-for="feature in features"
        :key="feature.title"
        :xs="24"
        :sm="12"
        :md="8"
      >
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

    <section class="sample-comics card shadow-sm">
      <div class="sample-comics__header">
        <div>
          <h3>示例漫画</h3>
          <p class="sample-comics__sub">
            来自其他人分享的漫画
          </p>
        </div>
        <el-button type="text" @click="goTo('home-sample-gallery')">
          查看全部
        </el-button>
      </div>
      <el-skeleton
        v-if="sampleLoading"
        :rows="4"
        animated
        class="sample-comics__skeleton"
      />
      <el-empty
        v-else-if="!sampleComics.length"
        description="还没有分享的漫画"
      />
      <el-row v-else :gutter="16">
        <el-col
          v-for="sample in sampleComics"
          :key="sample.id"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <el-card shadow="hover" class="sample-card">
            <div class="sample-card__cover">
              <img :src="resolveCover(sample)" :alt="sample.title" />
              <el-tag
                :type="sample.isShared ? 'success' : 'info'"
                effect="dark"
                class="sample-card__badge"
              >
                {{ sample.isShared ? "已分享" : "未分享" }}
              </el-tag>
            </div>
            <div class="sample-card__body">
              <h4>{{ sample.title }}</h4>
              <p class="sample-card__highlight">
                {{ sample.shareMessage || "创作者暂未填写副标题" }}
              </p>
              <div class="sample-card__meta">
                <el-tag size="small" effect="plain" type="success">
                  {{ sample.likesCount }} 喜欢
                </el-tag>
                <el-tag size="small" effect="plain" type="info">
                  {{ sample.commentsCount }} 留言
                </el-tag>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { fetchFeaturedComics, type StoredComic } from "../user-comics";

const router = useRouter();

const goTo = (name: string) => {
  router.push({ name });
};

const sampleComics = ref<StoredComic[]>([]);
const sampleLoading = ref(false);

const loadSampleComics = async () => {
  sampleLoading.value = true;
  try {
    sampleComics.value = await fetchFeaturedComics(5);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "无法获取分享漫画";
    ElMessage.error(message);
    sampleComics.value = [];
  } finally {
    sampleLoading.value = false;
  }
};

const resolveCover = (comic: StoredComic) => {
  if (comic.imageBase64) {
    return comic.imageBase64;
  }
  const meta = comic.metadata as Record<string, unknown> | null;
  if (meta && typeof meta["image_url"] === "string") {
    return meta["image_url"] as string;
  }
  return "https://picsum.photos/seed/welcome/600/360";
};

onMounted(loadSampleComics);

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

.sample-comics {
  padding: 24px;
  border-radius: 16px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sample-comics__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.sample-comics__sub {
  margin: 4px 0 0;
  color: var(--el-text-color-secondary);
}

.sample-card {
  border-radius: 12px;
  overflow: hidden;
}

.sample-card__cover {
  position: relative;
}

.sample-card__cover img {
  width: 100%;
  height: 180px;
  object-fit: cover;
  display: block;
}

.sample-card__badge {
  position: absolute;
  top: 8px;
  left: 8px;
}

.sample-card__body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 12px;
}

.sample-card__highlight {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin: 0;
}

.sample-card__message {
  margin: 0;
  font-size: 13px;
}

.sample-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-top: 8px;
}

.sample-card__stat {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

@media (max-width: 960px) {
  .welcome__hero {
    flex-direction: column;
    padding: 24px;
    gap: 16px;
  }

  .welcome__subtitle {
    max-width: 100%;
  }
}

@media (max-width: 640px) {
  .welcome__actions {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .welcome__actions .el-button {
    width: 100%;
  }
}
</style>
