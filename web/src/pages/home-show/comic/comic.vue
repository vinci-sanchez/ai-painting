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
            <div class="comic-menu-item">
              <span class="comic-menu-item__title">
                第{{ formatPageOrder(comicPage) }}页: {{ comicPage.title }}
              </span>
              <el-button
                v-if="canDeleteComic(comicPage)"
                class="comic-menu-item__delete"
                type="danger"
                link
                size="small"
                :loading="isDeletingComic(comicPage.id)"
                :title="`删除${comicPage.title}`"
                @click.stop="confirmDeleteComic(comicPage)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
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
        <div class="comic-card-header">
          <span class="page-title">{{ showPage.title }}</span>
          <div class="comic-right">
            <span>觉得不错？</span>
            <el-button type="primary" plain @click="handleShareClick">
              <el-icon><Share /></el-icon>
            </el-button>
          </div>
        </div>
        <!-- 分享面板 -->
        <el-dialog v-model="SharePanel" title=" 分享" :width="shareDialogWidth">
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
          v-loading="loading && !comicPages.length"
          element-loading-text="巴拉巴拉生成中..."
          :element-loading-spinner="svg"
          element-loading-svg-view-box="-10, -10, 50, 50"
          element-loading-background="rgba(122, 122, 122, 0.8)"
          style="width: 100%"
          class="page-preview"
        >
          <el-image
            style="
              width: 100%;
              min-height: 120px;
              height: auto;
              object-fit: contain;
            "
            :src="showPage.images || url"
            fit="cover"
          />
        </div>

        <div class="comic-actions">
          <div class="comic-actions__left">
            <span>不是喜欢的类型？-></span>
            <el-button type="primary" plain @click="ParameterModify">
              修改参数
            </el-button>
          </div>
          <div class="comic-right comic-nav">
            <el-button type="primary" plain @click="handlePrevPage">
              上一页
            </el-button>
            <el-button type="primary" plain @click="handleNextPage">
              下一页
            </el-button>
          </div>
        </div>
      </el-card>
    </el-main>
  </el-container>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, computed, watch } from "vue";
import QRCode from "qrcode";
import type { ImageProps } from "element-plus";
import { ElMessage, ElMessageBox } from "element-plus";
import bus from "../eventbus.js";
import { setParameterDraft } from "../shared-text";
import type { ParameterDraft } from "../shared-text";
import router from "../../../router.js";
import { currentUser } from "../auth-user";
import {
  refreshUserComics,
  userComics,
  deleteComicForCurrentUser,
} from "../user-comics";
import type { StoredComic } from "../user-comics";

type ComicPage = {
  id: number;
  title: string;
  images?: string;
  parameters?: ParameterDraft;
  pageNumber?: number;
  sourceUrl?: string;
  tempId?: number;
};
type ComicPayload = {
  url: string;
  page_id: number;
  title: string;
  parameters?: ParameterDraft;
  imageBase64?: string;
};
onMounted(() => {
  updateScreenMode();
  if (typeof window !== "undefined") {
    window.addEventListener("resize", updateScreenMode);
  }
  console.log("开始监听漫画生成事件");
  bus.on("make-imageover", Show_comic);
  bus.on("comic-persisted", handleComicPersisted);
});

onBeforeUnmount(() => {
  console.log("监听结束");
  bus.off("make-imageover", Show_comic);
  bus.off("comic-persisted", handleComicPersisted);
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", updateScreenMode);
  }
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
const isSmallScreen = ref(false);
const shareDialogWidth = computed(() =>
  isSmallScreen.value ? "92vw" : "500px"
);

const updateScreenMode = () => {
  if (typeof window === "undefined") {
    return;
  }
  isSmallScreen.value = window.innerWidth < 768;
};

const handleShareClick = async () => {
  SharePanel.value = true;
  Sharecomic.value = "";
  await generateShareComic();
};

const generateShareComic = async () => {
  if (SharecomicLoading.value) return;
  const targetUrl = "https://ai-painting.pages.dev/#/welcome";//currentComicImage.value;
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

const sessionPages = ref<ComicPage[]>([]);
const comicPages = ref<ComicPage[]>([]);
const activePageId = ref("");
const pageRefs = ref<Partial<Record<string, HTMLElement>>>({});
const deletingComicIds = ref<Record<string, boolean>>({});

const setPageRef = (el: HTMLElement | null, id: number) => {
  const key = id.toString();
  if (el) {
    pageRefs.value[key] = el;
  } else {
    delete pageRefs.value[key];
  }
};

const normalizeStoredComic = (record: StoredComic): ComicPage => {
  const metadata = record.metadata ?? null;
  const parameters = extractParameters(metadata);
  let sourceUrl: string | undefined;
  if (metadata && typeof metadata["image_url"] === "string") {
    sourceUrl = metadata["image_url"] as string;
  }
  return {
    id: record.id,
    title: record.title,
    images: record.imageBase64 || sourceUrl,
    pageNumber: record.pageNumber,
    sourceUrl,
    parameters,
  };
};

const extractParameters = (
  metadata?: Record<string, unknown> | null
): ParameterDraft | undefined => {
  if (!metadata) {
    return undefined;
  }

  const toRecord = (value: unknown): Record<string, unknown> | undefined => {
    if (!value || typeof value !== "object") {
      return undefined;
    }
    return value as Record<string, unknown>;
  };

  const buildDraftFromSources = (
    ...sources: Array<Record<string, unknown> | undefined>
  ): ParameterDraft | undefined => {
    const draft: ParameterDraft = {};
    let hasValue = false;
    const assignString = (
      key: Exclude<keyof ParameterDraft, "customTags">,
      value: unknown
    ) => {
      if (typeof value === "string") {
        draft[key] = value;
        hasValue = true;
      }
    };
    const assignTags = (value: unknown) => {
      if (!Array.isArray(value)) {
        return;
      }
      const tags = value.filter((item): item is string => typeof item === "string");
      draft.customTags = [...tags];
      hasValue = true;
    };

    sources.forEach((source) => {
      if (!source) {
        return;
      }
      assignString("title", source["title"]);
      assignString("scene", source["scene"]);
      assignString("prompt", source["prompt"]);
      assignString("dialogue", source["dialogue"]);
      assignString("character", source["character"]);
      assignString("style", source["style"]);
      assignString("color", source["color"]);
      assignTags(source["customTags"]);
      if (!source["customTags"]) {
        assignTags(source["hints"]);
      }
    });

    if (!hasValue) {
      return undefined;
    }
    if (Array.isArray(draft.customTags)) {
      draft.customTags = [...draft.customTags];
    }
    return draft;
  };

  const fromParameters = buildDraftFromSources(toRecord(metadata["parameters"]));
  if (fromParameters) {
    return fromParameters;
  }

  const storyboard = toRecord(metadata["storyboard"]);
  const legacyPayload = toRecord(metadata["StoryboardPayload"]);
  const legacyMeta = toRecord(metadata["StoryboardMetaPayload"]);
  return buildDraftFromSources(storyboard, legacyPayload, legacyMeta);
};

const syncComicPages = () => {
  const storedPages = userComics.value.map((item) =>
    normalizeStoredComic(item)
  );
  const combined = [...storedPages, ...sessionPages.value];
  comicPages.value = combined;
  if (!combined.length) {
    return;
  }
  const current = combined.find(
    (page) => page.id.toString() === activePageId.value
  );
  if (current) {
    showPage.value = current;
    return;
  }
  const first = combined[0];
  activePageId.value = first.id.toString();
  showPage.value = first;
};

watch(userComics, syncComicPages, { immediate: true, deep: true });
watch(
  sessionPages,
  () => {
    syncComicPages();
  },
  { deep: true }
);

watch(
  () => currentUser.value?.username,
  (username) => {
    if (username) {
      refreshUserComics(username);
    } else {
      sessionPages.value = [];
      comicPages.value = [];
      activePageId.value = "";
    }
  },
  { immediate: true }
);

//展示漫画
function Show_comic(payload: unknown) {
  const data = payload as ComicPayload;
  const tempId = typeof data.page_id === "number" ? data.page_id : Date.now();
  const newPage: ComicPage = {
    id: -Math.abs(tempId),
    tempId,
    title: data.title,
    images: data.imageBase64 || data.url,
    pageNumber: tempId,
    sourceUrl: data.url,
    parameters: data.parameters
      ? {
          ...data.parameters,
          customTags: data.parameters.customTags
            ? [...data.parameters.customTags]
            : [],
        }
      : undefined,
  };
  sessionPages.value = [newPage, ...sessionPages.value];
  activePageId.value = newPage.id.toString();
  showPage.value = newPage;
  loading.value = false;
}

function handleComicPersisted(payload: unknown) {
  const data = (payload ?? {}) as { temp_id?: number };
  const tempId = data.temp_id;
  if (typeof tempId !== "number") {
    return;
  }
  const index = sessionPages.value.findIndex((page) => page.tempId === tempId);
  if (index !== -1) {
    sessionPages.value.splice(index, 1);
  }
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

const formatPageOrder = (page: ComicPage) =>
  typeof page.pageNumber === "number" ? page.pageNumber : Math.abs(page.id);

const canDeleteComic = (page: ComicPage) => page.id > 0;
const isDeletingComic = (pageId: number) =>
  Boolean(deletingComicIds.value[pageId.toString()]);

const setDeletingState = (pageId: number, state: boolean) => {
  const key = pageId.toString();
  if (state) {
    deletingComicIds.value = { ...deletingComicIds.value, [key]: true };
  } else {
    const { [key]: _omit, ...rest } = deletingComicIds.value;
    deletingComicIds.value = rest;
  }
};

const confirmDeleteComic = async (page: ComicPage) => {
  if (!canDeleteComic(page) || isDeletingComic(page.id)) {
    return;
  }
  try {
    await ElMessageBox.confirm(
      `确认删除《${page.title}》吗？删除后无法恢复`,
      "删除漫画",
      {
        confirmButtonText: "删除",
        cancelButtonText: "取消",
        type: "warning",
      }
    );
  } catch {
    return;
  }
  await handleDeleteComic(page);
};

const handleDeleteComic = async (page: ComicPage) => {
  setDeletingState(page.id, true);
  try {
    await deleteComicForCurrentUser(page.id);
    ElMessage.success("漫画已删除");
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "删除漫画失败，请稍后重试";
    ElMessage.error(message);
  } finally {
    setDeletingState(page.id, false);
  }
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

.comic-menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
}

.comic-menu-item__title {
  flex: 1 1 auto;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.comic-menu-item__delete {
  flex-shrink: 0;
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
.page-preview :deep(.el-image) {
  width: min(100%, 480px);
  height: auto;
}

.comic-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.comic-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  flex-wrap: wrap;
}

.comic-actions__left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.comic-nav {
  gap: 8px;
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

@media (max-width: 960px) {
  .comic-layout {
    flex-direction: column;
  }

  .comic-sidebar {
    width: 100% !important;
    border-right: none;
    border-bottom: 1px solid var(--el-border-color);
    margin-bottom: 16px;
  }

  .comic-content {
    padding: 0;
  }

  .comic-menu {
    display: flex;
    flex-wrap: wrap;
  }

  .comic-menu .el-menu-item {
    flex: 1 1 50%;
  }
}

@media (max-width: 600px) {
  .comic-card-header,
  .comic-actions {
    flex-direction: column;
    align-items: flex-start;
  }

  .comic-right {
    justify-content: flex-start;
    width: 100%;
  }

  .comic-nav {
    width: 100%;
    justify-content: space-between;
  }

  .share-save {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
}
</style>
