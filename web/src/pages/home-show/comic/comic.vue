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
        <span class="page-title">{{ showPage.title }}</span>
        <div class="comic-right">
          <span>觉得不错？</span>
          <el-button type="primary" plain @click="handleShareClick">
            <el-icon><Share /></el-icon>
          </el-button>
        </div>
        <!-- 分享面板 -->
        <el-dialog v-model="SharePanel" title=" 分享" width="500">
          <!-- 二维码分享 -->
          <div class="share-qr">
            <span class="share-section-title">生成二维码</span>
            <div v-if="Sharecomic" class="share-qr-preview">
              <img
                :src="Sharecomic"
                alt="二维码..."
                style="width: 200px; min-height: 120px"
              />
              <el-button type="primary" link @click="saveShareQr">
                保存二维码
              </el-button>
            </div>
            <span v-else-if="SharecomicLoading">Generating QR code...</span>
            <el-button v-else type="primary" link @click="generateShareComic">Retry</el-button>
          </div>
          <!-- 保存至本地 -->
          <div class="share-save">
            <div>
              <div class="share-section-title">保存至本地</div>
              <div class="share-section-desc">保存当前显示的漫画图片</div>
            </div>
            <el-button
              type="primary"
              :loading="saveComicLoading"
              :disabled="!currentComicImage"
              @click="saveCurrentComic"
            >
              保存图片
            </el-button>
          </div>
          <template #footer>
            <div class="dialog-footer">
              <!-- <el-button @click="SharePanel = false">关闭</el-button>
              <el-button type="primary" @click="SharePanel = false">
                Confirm
              </el-button> -->
            </div>
          </template>
        </el-dialog>
        <!-- 漫画面板 -->
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
            style="
              width: 400px;
              min-height: 120px;
              height: auto;
              object-fit: contain;
            "
            :src="showPage.images || url"
            fit="cover"
          />
        </div>

        <span>不是喜欢的类型？-></span>
        <el-button type="primary" plain @click="ParameterModify"
          >修改参数</el-button
        ><span class="comic-right">
          <el-button type="primary" plain @click="handlePrevPage"
            >上一章</el-button
          >
          <el-button type="primary" plain @click="handleNextPage"
            >下一章</el-button
          ></span
        >
      </el-card>
    </el-main>
  </el-container>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, computed } from "vue";
import QRCode from "qrcode";
import type { ImageProps } from "element-plus";
import { ElMessage } from "element-plus";
import bus from "../eventbus.js";
import { comicimage, setComicImage, setParameterDraft } from "../shared-text";
import type { ParameterDraft } from "../shared-text";
import router from "../../../router.js";

type ComicPage = {
  id: number;
  title: string;
  images?: string;
  parameters?: ParameterDraft;
};
type ComicPayload = {
  url: string;
  page_id: number;
  title: string;
  parameters?: ParameterDraft;
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
const SharePanel = ref(false);
const Sharecomic = ref("");
const SharecomicLoading = ref(false);
const saveComicLoading = ref(false);
const currentComicImage = computed(() => showPage.value.images || "");

const handleShareClick = async () => {
  SharePanel.value = true;
  Sharecomic.value = "";
  await generateShareComic();
};

const generateShareComic = async () => {
  if (SharecomicLoading.value) return;
  const targetUrl = currentComicImage.value;
  if (!targetUrl) {
    Sharecomic.value = "";
    return;
  }
  try {
    SharecomicLoading.value = true;
    Sharecomic.value = await makeQr(targetUrl);
  } catch (error) {
    console.error("二维码生成失败", error);
    Sharecomic.value = "";
  } finally {
    SharecomicLoading.value = false;
  }
};
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

//展示漫画
function Show_comic(payload: unknown) {
  const data = payload as ComicPayload;
  const newPage: ComicPage = {
    id: data.page_id,
    title: data.title,
    images: data.url,
    parameters: data.parameters
      ? {
          ...data.parameters,
          customTags: data.parameters.customTags
            ? [...data.parameters.customTags]
            : [],
        }
      : undefined,
  };
  comicPages.value.push(newPage);
  activePageId.value = newPage.id.toString();
  showPage.value = newPage;
  loading.value = false;
}

const handlePageSelect = (index: string) => {
  activePageId.value = index;
  const selected = comicPages.value.find(
    (page) => page.id.toString() === index
  );
  if (selected) {
    showPage.value = selected;
  }
};
///////修改参数
function ParameterModify() {
  if (!comicPages.value.length) {
    router.push({ name: "home-parameter-preview" });
    return;
  }
  const currentPage = showPage.value;
  if (!currentPage) {
    router.push({ name: "home-parameter-preview" });
    return;
  }
  const parameters = currentPage.parameters;
  if (parameters) {
    setParameterDraft({
      ...parameters,
      title: currentPage.title,
      customTags: parameters.customTags ? [...parameters.customTags] : [],
    });
  } else {
    setParameterDraft({
      title: currentPage.title,
    });
  }
  router.push({ name: "home-parameter-preview" });
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
///////二维码生成
async function makeQr(url: string) {
  const dataUrl = await QRCode.toDataURL(url, {
    errorCorrectionLevel: "M",
    margin: 2,
    width: 320,
  });
  return dataUrl; // base64 png，可直接 <img src=...>
}

const saveCurrentComic = () => {
  const imageSrc = currentComicImage.value;
  if (!imageSrc) {
    ElMessage.warning("当前没有可保存的漫画图片");
    return;
  }
  try {
    saveComicLoading.value = true;
    const title = (showPage.value.title || "comic").replace(/\s+/g, "_");
    const link = document.createElement("a");
    link.href = imageSrc;
    link.download = `${title}_${showPage.value.id || "page"}.png`;
    link.target = "_blank";
    link.rel = "noopener";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  } catch (error) {
    console.error("保存漫画失败", error);
    ElMessage.error("保存失败，请稍后重试");
  } finally {
    saveComicLoading.value = false;
  }
};

const saveShareQr = () => {
  if (!Sharecomic.value) {
    ElMessage.warning("当前没有可保存的二维码");
    return;
  }
  const link = document.createElement("a");
  const title = (showPage.value.title || "comic").replace(/\s+/g, "_");
  link.href = Sharecomic.value;
  link.download = `${title}_qr.png`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
};
</script>

<style scoped>
.comic-layout {
  min-height: 400px;
}

.comic-right {
  display: flex;
  align-items: center;
  justify-content: flex-end; /* 改成 flex-start 就是靠左 */
  gap: 8px;
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

.share-qr {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.share-qr-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.share-save {
  margin-top: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.share-section-title {
  font-weight: 600;
  margin-bottom: 4px;
  display: inline-block;
}

.share-section-desc {
  font-size: 12px;
  color: #909399;
}
</style>
