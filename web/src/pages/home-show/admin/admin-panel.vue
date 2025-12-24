<template>
  <div class="admin-panel">
    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h3>上传分享漫画</h3>
          <span>管理员可代用户上传，至少填写用户名、标题与图片</span>
        </div>
      </template>
      <el-form label-width="120px" label-position="top">
        <el-form-item label="目标用户名">
          <el-input v-model="uploadForm.username" placeholder="输入要代传的用户名" />
        </el-form-item>
        <el-form-item label="标题（分享时展示）">
          <el-input v-model="uploadForm.title" placeholder="示例：城市侦探：黎明篇" />
        </el-form-item>
        <el-form-item label="页码（可选）">
          <el-input-number v-model="uploadForm.pageNumber" :min="1" />
        </el-form-item>
        <el-form-item label="图片 URL（或上传本地文件）">
          <el-input
            v-model="uploadForm.imageUrl"
            placeholder="https://example.com/demo.png"
          />
          <input
            class="file-input"
            type="file"
            accept="image/*"
            @change="handleFileChange"
          />
        </el-form-item>
        <el-form-item label="分享副标题 / 留言">
          <el-input
            v-model="uploadForm.shareMessage"
            type="textarea"
            :rows="2"
            placeholder="仅填写标题也会自动分享，这里可选"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="uploadLoading"
            @click="submitUpload"
          >
            上传漫画
          </el-button>
          <el-button @click="resetUploadForm">清空</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h3>喜欢数管理</h3>
          <span>输入漫画 ID，可立即为该作品 +1 喜欢</span>
        </div>
      </template>
      <el-form label-position="top">
        <el-form-item label="漫画 ID">
          <el-input v-model="likeComicId" placeholder="整数 ID" />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="likeLoading"
            @click="handleAdminLike"
          >
            +1 喜欢
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h3>删除指定漫画</h3>
          <span>输入用户名 + 漫画 ID 即可删除</span>
        </div>
      </template>
      <el-form label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="deleteForm.username" placeholder="username" />
        </el-form-item>
        <el-form-item label="漫画 ID">
          <el-input
            v-model="deleteForm.comicId"
            placeholder="输入整数 ID"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="danger"
            :loading="deleteLoading"
            @click="handleDeleteComic"
          >
            删除漫画
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="admin-card">
      <template #header>
        <div class="card-header">
          <h3>留言管理</h3>
          <span>查询指定漫画的留言，并可单独删除</span>
        </div>
      </template>
      <div class="comment-tools">
        <el-input
          v-model="commentComicId"
          placeholder="输入漫画 ID"
          style="max-width: 240px"
        />
        <el-button type="primary" :loading="commentsLoading" @click="loadComments">
          读取留言
        </el-button>
        <el-button text @click="clearComments">清空</el-button>
      </div>
      <el-empty
        v-if="!commentsLoading && !commentList.length"
        description="暂无留言数据"
      />
      <el-skeleton v-else-if="commentsLoading" :rows="3" animated />
      <el-timeline v-else>
        <el-timeline-item
          v-for="item in commentList"
          :key="item.id"
          :timestamp="formatCommentTime(item.createdAt)"
        >
          <div class="comment-item">
            <div class="comment-item__header">
              <strong>{{ item.author }}</strong>
              <el-button
                type="danger"
                link
                size="small"
                :loading="commentDeleting === item.id"
                @click="deleteComment(item.id)"
              >
                删除
              </el-button>
            </div>
            <p class="comment-item__content">{{ item.content }}</p>
          </div>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import config from "../../config.json";

const BACK_URL =
  (config as Record<string, string | undefined>).BACK_URL ??
  "http://localhost:3000";

const uploadForm = reactive({
  username: "",
  title: "",
  pageNumber: 1,
  imageUrl: "",
  imageBase64: "",
  shareMessage: "",
});
const uploadLoading = ref(false);

const deleteForm = reactive({
  username: "",
  comicId: "",
});
const deleteLoading = ref(false);
const likeComicId = ref("");
const likeLoading = ref(false);

const commentComicId = ref("");
const commentList = ref<
  Array<{ id: number; author: string; content: string; createdAt: string }>
>([]);
const commentsLoading = ref(false);
const commentDeleting = ref<number | null>(null);

const fileToBase64 = (file: File): Promise<string> =>
  new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = typeof reader.result === "string" ? reader.result : "";
      resolve(result);
    };
    reader.onerror = (error) => reject(error);
    reader.readAsDataURL(file);
  });

const handleFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) {
    return;
  }
  try {
    uploadForm.imageBase64 = await fileToBase64(file);
    ElMessage.success("已读取本地图片");
  } catch (error) {
    console.error(error);
    ElMessage.error("读取图片失败");
  } finally {
    input.value = "";
  }
};

const resetUploadForm = () => {
  uploadForm.username = "";
  uploadForm.title = "";
  uploadForm.pageNumber = 1;
  uploadForm.imageUrl = "";
  uploadForm.imageBase64 = "";
  uploadForm.shareMessage = "";
};

const submitUpload = async () => {
  const payloadUsername = uploadForm.username.trim();
  if (!payloadUsername) {
    ElMessage.warning("请输入目标用户名");
    return;
  }
  if (!uploadForm.title.trim()) {
    ElMessage.warning("标题不能为空");
    return;
  }
  if (!uploadForm.imageBase64 && !uploadForm.imageUrl.trim()) {
    ElMessage.warning("请提供图片 URL 或上传本地图片");
    return;
  }
  const body = {
    title: uploadForm.title.trim(),
    page_number: uploadForm.pageNumber || 1,
    image_base64: uploadForm.imageBase64 || "",
    image_url: uploadForm.imageBase64 ? "" : uploadForm.imageUrl.trim(),
    metadata: {},
    is_shared: true,
    share_message: uploadForm.shareMessage.trim(),
  };
  uploadLoading.value = true;
  try {
    const response = await fetch(
      `${BACK_URL}/api/users/${encodeURIComponent(payloadUsername)}/comics`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      }
    );
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || "上传失败");
    }
    ElMessage.success("漫画已上传并自动分享");
    resetUploadForm();
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "上传失败，请稍后重试";
    ElMessage.error(message);
  } finally {
    uploadLoading.value = false;
  }
};

const handleDeleteComic = async () => {
  const username = deleteForm.username.trim();
  const comicId = Number(deleteForm.comicId);
  if (!username || !comicId) {
    ElMessage.warning("请输入用户名与漫画ID");
    return;
  }
  try {
    await ElMessageBox.confirm(
      `确定删除 ${username} 的漫画 #${comicId} 吗？`,
      "删除确认",
      { type: "warning" }
    );
  } catch {
    return;
  }

  deleteLoading.value = true;
  try {
    const response = await fetch(
      `${BACK_URL}/api/users/${encodeURIComponent(username)}/comics/${comicId}`,
      { method: "DELETE" }
    );
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || "删除失败");
    }
    ElMessage.success("漫画已删除");
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "删除失败，请稍后再试";
    ElMessage.error(message);
  } finally {
    deleteLoading.value = false;
  }
};

const loadComments = async () => {
  const comicId = Number(commentComicId.value);
  if (!comicId) {
    ElMessage.warning("请输入有效的漫画ID");
    return;
  }
  commentsLoading.value = true;
  try {
    const response = await fetch(`${BACK_URL}/api/comics/${comicId}/comments`);
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || "获取留言失败");
    }
    commentList.value = Array.isArray(data?.comments)
      ? data.comments.map((item: Record<string, unknown>) => ({
          id: Number(item.id) || 0,
          author: typeof item.author === "string" ? item.author : "游客",
          content: typeof item.content === "string" ? item.content : "",
          createdAt:
            typeof item.created_at === "string" ? item.created_at : "",
        }))
      : [];
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "获取留言失败，请稍后再试";
    ElMessage.error(message);
    commentList.value = [];
  } finally {
    commentsLoading.value = false;
  }
};

const clearComments = () => {
  commentList.value = [];
  commentComicId.value = "";
};

const deleteComment = async (commentId: number) => {
  try {
    await ElMessageBox.confirm("确定删除该留言吗？", "删除确认", {
      type: "warning",
    });
  } catch {
    return;
  }
  commentDeleting.value = commentId;
  try {
    const response = await fetch(
      `${BACK_URL}/api/comics/comments/${commentId}`,
      { method: "DELETE" }
    );
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || "删除失败");
    }
    commentList.value = commentList.value.filter(
      (item) => item.id !== commentId
    );
    ElMessage.success("留言已删除");
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "删除失败，请稍后重试";
    ElMessage.error(message);
  } finally {
    commentDeleting.value = null;
  }
};

const handleAdminLike = async () => {
  const comicId = Number(likeComicId.value);
  if (!comicId) {
    ElMessage.warning("请输入有效的漫画ID");
    return;
  }
  likeLoading.value = true;
  try {
    const response = await fetch(
      `${BACK_URL}/api/comics/${comicId}/like`,
      { method: "POST" }
    );
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data?.error || "点赞失败");
    }
    ElMessage.success(
      `点赞成功，当前喜欢数：${typeof data.likes === "number" ? data.likes : "未知"}`
    );
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "点赞失败，请稍后再试";
    ElMessage.error(message);
  } finally {
    likeLoading.value = false;
  }
};

const formatCommentTime = (value: string) => {
  if (!value) {
    return "";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};
</script>

<style scoped>
.admin-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.admin-card {
  border: none;
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.file-input {
  margin-top: 8px;
}

.comment-tools {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.comment-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.comment-item__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.comment-item__content {
  margin: 0;
  white-space: pre-wrap;
}
</style>
