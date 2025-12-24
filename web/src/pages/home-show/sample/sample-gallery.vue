<template>
  <div class="sample-gallery">
    <section class="sample-gallery__hero card shadow-sm">
      <div>
        <p class="sample-gallery__eyebrow">分享墙</p>
        <h2>示例漫画</h2>
        <p>
          实时展示点赞和留言热度最高的分享漫画，了解灵感趋势，也能直接与作者互动。
        </p>
      </div>
      <el-button type="primary" @click="goToComic">前往漫画页</el-button>
    </section>

    <el-skeleton
      v-if="!initialLoaded && loading"
      :rows="4"
      animated
      class="sample-comics__skeleton"
    />
    <el-empty
      v-else-if="initialLoaded && !comics.length"
      description="暂时没有分享的漫画"
      class="sample-gallery__empty"
    />
    <el-row v-else :gutter="16">
      <el-col
        v-for="comic in comics"
        :key="comic.id"
        :xs="24"
        :sm="12"
        :md="8"
      >
        <el-card shadow="hover" class="sample-card">
          <div class="sample-card__cover">
            <el-image
              :src="resolveCover(comic)"
              fit="cover"
              lazy
              style="width: 100%; height: 180px; cursor: zoom-in"
              @click="openPreview(comic)"
            />
            <el-tag
              class="sample-card__badge"
              :type="comic.isShared ? 'success' : 'info'"
              effect="dark"
            >
              {{ comic.isShared ? "已分享" : "未分享" }}
            </el-tag>
          </div>
          <div class="sample-card__body">
            <h3>{{ comic.title }}</h3>
            <p class="sample-card__id">ID: {{ comic.id }}</p>
            <div class="sample-card__meta">
              <el-button
                type="danger"
                link
                :loading="likeLoadingId === comic.id"
                @click="handleLikeComic(comic)"
              >
                <i class="fas fa-heart"></i> {{ comic.likesCount }} 喜欢
              </el-button>
              <span><i class="fas fa-comment-dots"></i> {{ comic.commentsCount }} 留言</span>
            </div>
            <div class="sample-card__actions">
              <el-button text size="small" @click="openCommentDialog(comic)">
                留言互动
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div
      v-if="initialLoaded && comics.length"
      ref="loadMoreRef"
      class="load-more-sentinel"
    >
      <span v-if="loading && hasMore">加载中...</span>
      <span v-else-if="!hasMore">已经到底啦</span>
    </div>

    <el-divider />
    <section class="my-comics">
      <div class="sample-comics__header">
        <div>
          <h3>我的分享</h3>
          <p class="sample-comics__sub">自动展示我已保存并分享的作品</p>
        </div>
      </div>
      <el-empty
        v-if="!mySharedComics.length"
        description="暂无个人分享，去生成一张吧！"
      />
      <el-row v-else :gutter="16">
        <el-col
          v-for="comic in mySharedComics"
          :key="comic.id"
          :xs="24"
          :sm="12"
          :md="8"
        >
          <el-card shadow="hover" class="sample-card">
            <div class="sample-card__cover">
              <el-image
                :src="resolveCover(comic)"
                fit="cover"
                lazy
                style="width: 100%; height: 180px; cursor: zoom-in"
                @click="openPreview(comic)"
              />
              <el-tag class="sample-card__badge" type="primary" effect="dark">
                我的作品
              </el-tag>
            </div>
            <div class="sample-card__body">
              <h3>{{ comic.title }}</h3>
              <p class="sample-card__id">ID: {{ comic.id }}</p>
              <div class="sample-card__meta">
                <el-button
                  type="danger"
                  link
                  :loading="likeLoadingId === comic.id"
                  @click="handleLikeComic(comic)"
                >
                  <i class="fas fa-heart"></i> {{ comic.likesCount }} 喜欢
                </el-button>
                <span><i class="fas fa-comment-dots"></i> {{ comic.commentsCount }} 留言</span>
              </div>
              <div class="sample-card__actions">
                <el-button text size="small" @click="openCommentDialog(comic)">
                  查看留言
                </el-button>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </section>

    <el-dialog
      v-model="imagePreviewVisible"
      class="preview-dialog"
      width="80vw"
      align-center
    >
      <img :src="imagePreviewSrc" alt="preview" class="preview-image" />
    </el-dialog>

    <el-dialog
      v-model="commentDialogVisible"
      :title="commentTarget?.title || '留言互动'"
      width="520px"
    >
      <div v-if="commentTarget" class="comment-dialog">
        <p class="comment-dialog__hint">
          给《{{ commentTarget.title }}》留句话吧
        </p>
        <el-input
          v-model="commentForm.author"
          placeholder="昵称（可选）"
          class="comment-dialog__field"
          :disabled="commentSubmitting"
        />
        <el-input
          v-model="commentForm.content"
          type="textarea"
          :rows="3"
          maxlength="120"
          show-word-limit
          placeholder="写下留言，120 字以内"
          class="comment-dialog__field"
          :disabled="commentSubmitting"
        />
        <div class="comment-dialog__actions">
          <el-button
            type="primary"
            :loading="commentSubmitting"
            @click="submitComment"
          >
            提交留言
          </el-button>
          <el-button text @click="goToComic">前往漫画页</el-button>
        </div>
        <el-divider />
        <el-skeleton v-if="commentsLoading" :rows="3" animated />
        <el-empty v-else-if="!commentList.length" description="还没有留言" />
        <ul v-else class="comment-list">
          <li v-for="item in commentList" :key="item.id" class="comment-item">
            <div class="comment-item__header">
              <strong>{{ item.author }}</strong>
              <span>{{ formatCommentTime(item.createdAt) }}</span>
            </div>
            <p class="comment-item__content">{{ item.content }}</p>
          </li>
        </ul>
      </div>
    </el-dialog>
  </div>
</template>

<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { useRouter } from "vue-router";
import {
  fetchFeaturedComics,
  fetchComicComments,
  addComicComment,
  likeComic,
  userComics,
  refreshUserComics,
  type StoredComic,
  type ComicComment,
} from "../user-comics";
import { currentUser } from "../auth-user";

const comics = ref<StoredComic[]>([]);
const loading = ref(false);
const initialLoaded = ref(false);
const hasMore = ref(true);
const pageSize = 12;
const router = useRouter();
const loadMoreRef = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const commentDialogVisible = ref(false);
const commentTarget = ref<StoredComic | null>(null);
const commentList = ref<ComicComment[]>([]);
const commentsLoading = ref(false);
const commentSubmitting = ref(false);
const commentForm = ref({
  author: "",
  content: "",
});
const likeLoadingId = ref<number | null>(null);
const imagePreviewVisible = ref(false);
const imagePreviewSrc = ref("");

const mySharedComics = computed(() =>
  userComics.value.filter((item) => item.isShared)
);

const loadMoreComics = async () => {
  if (loading.value || !hasMore.value) {
    return;
  }
  loading.value = true;
  try {
    const next = await fetchFeaturedComics(pageSize, comics.value.length);
    if (next.length < pageSize) {
      hasMore.value = false;
    }
    comics.value = [...comics.value, ...next];
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "无法加载分享漫画";
    ElMessage.error(message);
  } finally {
    loading.value = false;
    initialLoaded.value = true;
  }
};

const refreshComics = async () => {
  comics.value = [];
  hasMore.value = true;
  initialLoaded.value = false;
  await loadMoreComics();
};

const goToComic = () => {
  router.push({ name: "home-comic" });
};

const resolveCover = (comic: StoredComic) => {
  if (comic.imageBase64) {
    return comic.imageBase64;
  }
  const meta = comic.metadata as Record<string, unknown> | null;
  if (meta && typeof meta["image_url"] === "string") {
    return meta["image_url"] as string;
  }
  return "https://picsum.photos/seed/sample/600/360";
};

const formatCommentTime = (value: string) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};

const openCommentDialog = async (comic: StoredComic) => {
  commentTarget.value = comic;
  if (!commentForm.value.author && currentUser.value?.username) {
    commentForm.value.author = currentUser.value.username;
  }
  commentForm.value.content = "";
  commentDialogVisible.value = true;
  await loadComments(comic.id);
};

const loadComments = async (comicId: number) => {
  commentsLoading.value = true;
  try {
    commentList.value = await fetchComicComments(comicId);
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "无法获取留言";
    ElMessage.error(message);
    commentList.value = [];
  } finally {
    commentsLoading.value = false;
  }
};

const submitComment = async () => {
  const target = commentTarget.value;
  if (!target) {
    return;
  }
  const content = commentForm.value.content.trim();
  if (!content) {
    ElMessage.warning("请先填写留言内容");
    return;
  }
  try {
    commentSubmitting.value = true;
    const author =
      commentForm.value.author.trim() ||
      currentUser.value?.username ||
      "游客";
    const created = await addComicComment(target.id, {
      author,
      content,
    });
    commentList.value = [created, ...commentList.value];
    commentForm.value.content = "";
    ElMessage.success("留言已提交");
    await refreshComics();
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "留言失败，请稍后再试";
    ElMessage.error(message);
  } finally {
    commentSubmitting.value = false;
  }
};

const openPreview = (comic: StoredComic) => {
  imagePreviewSrc.value = resolveCover(comic);
  imagePreviewVisible.value = true;
};

const handleLikeComic = async (comic: StoredComic) => {
  if (!comic?.id || likeLoadingId.value === comic.id) {
    return;
  }
  likeLoadingId.value = comic.id;
  try {
    const likes = await likeComic(comic.id);
    comics.value = comics.value.map((item) =>
      item.id === comic.id ? { ...item, likesCount: likes } : item
    );
    if (currentUser.value?.username) {
      refreshUserComics(currentUser.value.username);
    }
    ElMessage.success("已点赞");
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "点赞失败，请稍后再试";
    ElMessage.error(message);
  } finally {
    likeLoadingId.value = null;
  }
};

watch(
  () => currentUser.value?.username,
  (name) => {
    if (name) {
      refreshUserComics(name);
      if (!commentForm.value.author) {
        commentForm.value.author = name;
      }
    }
  },
  { immediate: true }
);

watch(
  loadMoreRef,
  (el, prev) => {
    if (prev && observer) {
      observer.unobserve(prev);
    }
    if (el && observer) {
      observer.observe(el);
    }
  },
  { flush: "post" }
);

onMounted(() => {
  loadMoreComics();
  observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        loadMoreComics();
      }
    });
  });
  if (loadMoreRef.value) {
    observer.observe(loadMoreRef.value);
  }
});

onBeforeUnmount(() => {
  if (observer && loadMoreRef.value) {
    observer.unobserve(loadMoreRef.value);
  }
  observer = null;
});
</script>

<style scoped>
.sample-gallery {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sample-gallery__hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
  border: none;
}

.sample-gallery__eyebrow {
  letter-spacing: 0.2em;
  text-transform: uppercase;
  font-size: 12px;
  color: var(--el-color-primary);
  margin-bottom: 8px;
}

.sample-gallery__empty {
  margin-top: 24px;
}

.sample-comics__skeleton {
  margin-top: 24px;
}

.sample-card {
  border-radius: 12px;
  overflow: hidden;
  margin-bottom: 12px;
}

.sample-card__cover {
  position: relative;
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

.sample-card__id {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.sample-card__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  gap: 8px;
}

.sample-card__meta .el-button {
  padding: 0;
}

.sample-card__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.comment-dialog__field {
  margin-bottom: 8px;
}

.comment-dialog__actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.comment-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.comment-item {
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
}

.comment-item__header {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.comment-item__content {
  margin: 4px 0 0;
  white-space: pre-wrap;
  line-height: 1.4;
}

.load-more-sentinel {
  text-align: center;
  padding: 12px 0;
  color: var(--el-text-color-secondary);
}

.preview-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.preview-image {
  width: 100%;
  display: block;
  object-fit: contain;
}

@media (max-width: 768px) {
  .sample-gallery__hero {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>
