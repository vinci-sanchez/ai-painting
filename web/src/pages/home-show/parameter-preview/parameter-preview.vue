<template>
  <el-card class="parameter-preview">
    <el-descriptions
      class="margin-top"
      title="项目内容"
      :column="descriptionColumns"
      :size="size"
      border
    >
      <!-- <template #extra>
        <el-button type="primary">Operation</el-button>
      </template> -->
      <el-descriptions-item>
        <template #label>
          <div class="cell-item">
            <el-icon :style="iconStyle">
              <user />
            </el-icon>
            项目名称
          </div>
        </template>
        <el-input v-model="novel" placeholder="可填可不填" />
      </el-descriptions-item>
      <el-descriptions-item>
        <template #label>
          <div class="cell-item">
            <el-icon :style="iconStyle">
              <Clock />
            </el-icon>
            时间
          </div>
        </template>
        <div>{{ now_time }}时</div>
      </el-descriptions-item>

      <el-descriptions-item>
        <template #label>
          <div class="cell-item">
            <el-icon :style="iconStyle">
              <location />
            </el-icon>
            位置
          </div>
        </template>
        <div>{{ Locate }}（大概率不准）</div>
      </el-descriptions-item>

      <el-descriptions-item>
        <template #label>
          <div class="cell-item">
            <el-icon :style="iconStyle">
              <CollectionTag />
            </el-icon>
            自定义提示词
          </div>
        </template>
        <el-descriptions-item label="提示词">
          <el-space wrap>
            <el-tag
              v-for="seg in promptSegments"
              :key="seg"
              size="small"
              type="info"
            >
              {{ seg }}
            </el-tag>
          </el-space>
        </el-descriptions-item>
      </el-descriptions-item>
    </el-descriptions>

    <!-- 其他配置 -->
    <div class="parameter-actions">
      <el-button
        type="primary"
        @click="ConfigurationClick"
        style="
          width: 100px;
          display: block;
          height: 35px;
          width: fit-content;
        "
        >其他配置</el-button
      >
      <el-button
        type="primary"
        @click="UploadImg"
        style="
          width: 100px;
          display: block;
          height: 35px;
          width: fit-content;
        "
        >上传图像</el-button
      >
    </div>
    <el-dialog
      v-model="Configuration"
      title=" 配置(不懂默认就行)"
      :width="dialogWidth"
    >
      <el-form label-position="top" label-width="100px">
        <el-form-item label="BASE_URL">
          <el-input v-model="UPDATE_BASE_URL" />
        </el-form-item>
        <el-form-item label="接入点模型ID">
          <el-input v-model="UPDATE_IMAGE_MODEL" />
        </el-form-item>
        <el-form-item label="API_KEY">
          <el-input
            v-model="UPDATE_API_KEY"
            placeholder="如果更改,此项为必填项"
          />
        </el-form-item>
      </el-form>
    </el-dialog>
      <el-dialog
        v-model="uploadDialogVisible"
        title="上传图像"
        :width="dialogWidth"
      >
        <div class="upload-dialog-content">
   
          <div class="upload-section">
        
            <el-upload
              class="upload-block"
              drag
              :auto-upload="false"
              action=""
              accept="image/*"
              :file-list="uploadFiles"
              :on-change="handleUploadChange"
              :on-remove="handleUploadRemove"
            >
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="el-upload__text">
                将文件拖到此处，或 <em>点击上传</em>
              </div>
              <template #tip>
                <div class="el-upload__tip">仅用于参考展示，当前不会上传到服务器</div>
              </template>
            </el-upload>
            <div v-if="uploadPreview" class="upload-preview">
              <img :src="uploadPreview" alt="预览图" />
            </div>
          </div>
          <div class="upload-section">
            <h6>图像介绍</h6>
            <el-input
              v-model="uploadDescription"
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 6 }"
              placeholder="请输入对参考图像的简要说明，图为人物长相或其他信息"
            />
          </div>
        </div>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="uploadDialogVisible = false">取 消</el-button>
            <el-button type="primary" @click="handleUploadConfirm">完 成</el-button>
          </div>
        </template>
      </el-dialog>
    <div
      v-loading="loading"
      element-loading-text="巴拉巴拉生成中..."
      :element-loading-spinner="svg"
      element-loading-svg-view-box="-10, -10, 50, 50"
      element-loading-background="rgba(122, 122, 122, 0.8)"
      style="width: 100%"
      class="page-preview"
    >
      <el-descriptions
        class="margin-top"
        title="详细配置"
        :column="descriptionColumns"
        :size="size"
        :style="blockMargin"
      >
        <template #extra> </template>

        <el-descriptions-item label="项目名称" :span="3">
          {{ novel }}
        </el-descriptions-item>
        <el-descriptions-item label="创造时间">
          {{ now_time }}
        </el-descriptions-item>
        <el-descriptions-item label="风格">{{ color }}</el-descriptions-item>
        <el-descriptions-item label="色调">{{ style }}</el-descriptions-item>

        <el-descriptions-item label="提示词">
          <el-tag v-for="value in customTags" size="small">{{ value }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <p>分镜内容</p>
 
       
      <h6>场景</h6>
      <el-input
        v-model="scene"
        :autosize="{ minRows: 10, maxRows: 20 }"
      ></el-input>
      <h6>提示词</h6>
      <el-input
        v-model="prompt"
        :autosize="{ minRows: 10, maxRows: 20 }"
      ></el-input>
      <h6>对白</h6>
      <el-input
        v-model="dialogue"
        :autosize="{ minRows: 10, maxRows: 20 }"
      ></el-input>
      <h6>人物</h6>
      <el-input
        v-model="character"
        :autosize="{ minRows: 10, maxRows: 20 }"
      ></el-input>
    </div>
    <template #footer
      ><div class="card-footer-actions">
        <el-button type="primary" @click="goGenerateComic">生成漫画</el-button>
      </div>
    </template>
  </el-card>
</template>
<script lang="ts" setup>
defineOptions({ name: "parameter-preview" });
import { computed, ref, onMounted, onBeforeUnmount, watch } from "vue";
import { ElMessage } from "element-plus";
import type { UploadProps, UploadUserFile } from "element-plus";
import bus from "../eventbus.js";
import { currentUser } from "../auth-user";
import {
  Iphone,
  Location,
  OfficeBuilding,
  Tickets,
  User,
  UploadFilled,
} from "@element-plus/icons-vue";
import {
  setComicText,
  comicText,
  setComicImage,
  comicimage,
  parameterDraft,
  setParameterDraft,
} from "../shared-text";
import type { ParameterDraft as ComicParameterDraft } from "../shared-text";
import router from "../../../router.js";
import config from "../../config.json";
import { saveComicForCurrentUser } from "../user-comics";
const BACK_URL = config.BACK_URL;
const API_BASE_URL = config.BASE_URL;
const novel = ref("");
const loading = ref(false);
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

/////////////////获取时间
import type { ComponentSize } from "element-plus";
function gettime_str() {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  const hour = String(now.getHours()).padStart(2, "0");
  const formattedDate = `${year}  ${month}.${day} ${hour}`;
  console.log(formattedDate); // 示例输出：2025  10.04 14:30 (取决于当前时间)
  return formattedDate;
}
const now_time = ref(gettime_str());
novel.value = gettime_str();

////////////////////ip定位
const Locate = ref("定位中...(不会进行记录,单纯玩)");
async function getlocate() {
  try {
    const res = await fetch("https://vincisanchez.dpdns.org/get-city");
    const data = await res.json();
    Locate.value = data.city || "未知";
  } catch (e) {
    console.error("定位失败", e);
    Locate.value = "定位失败";
    ElMessage.error("定位失败,请关闭浏览器跟踪防护");
  }
}
onMounted(getlocate);

///////////////自定义提示词
const promptSegments = computed(() =>
  (novel.value || "")
    .split(/[，,；;。\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
);
/////自定义配置
const UPDATE_IMAGE_MODEL = ref(
  config.IMAGE_TO_IMAGE_MODEL || config.TEXT_TO_IMAGE_MODEL
);
const UPDATE_BASE_URL = ref(API_BASE_URL);
const UPDATE_API_KEY = ref("");
/////分镜接受
const scene = ref("");
const prompt = ref("");
const dialogue = ref("");
const character = ref("");
const style = ref("");
const color = ref("");
const customTags = ref<string[]>([]);

const applyParameterDraft = (draft: ComicParameterDraft) => {
  if (draft.title !== undefined) {
    novel.value = draft.title ?? "";
  }
  if (draft.scene !== undefined) {
    scene.value = draft.scene ?? "";
  }
  if (draft.prompt !== undefined) {
    prompt.value = draft.prompt ?? "";
  }
  if (draft.dialogue !== undefined) {
    dialogue.value = draft.dialogue ?? "";
  }
  if (draft.character !== undefined) {
    character.value = draft.character ?? "";
  }
  if (draft.style !== undefined) {
    style.value = draft.style ?? "";
  }
  if (draft.color !== undefined) {
    color.value = draft.color ?? "";
  }
  if (draft.customTags !== undefined) {
    customTags.value = [...draft.customTags];
  }
  loading.value = false;
};

watch(
  () => parameterDraft.value,
  (value) => {
    if (value) {
      applyParameterDraft(value);
      setParameterDraft(null);
    }
  },
  { immediate: true }
);
type StoryboardPayload = {
  scene?: string;
  prompt?: string;
  dialogue?: string;
  character?: string;
};

type StoryboardMetaPayload = {
  style?: string;
  color?: string;
  hints?: string[];
};
type ComicData = {
  StoryboardPayload: StoryboardPayload;
  StoryboardMetaPayload: StoryboardMetaPayload;
};
var page_id = 1;

const buildParametersForPage = (): ComicParameterDraft => ({
  title: novel.value,
  scene: scene.value,
  prompt: prompt.value,
  dialogue: dialogue.value,
  character: character.value,
  style: style.value,
  color: color.value,
  customTags: [...customTags.value],
});

const buildMetadataForComic = (
  parameters: ComicParameterDraft,
  imageUrl: string
) => {
  const storyboard = {
    title: parameters.title ?? "",
    scene: parameters.scene ?? "",
    prompt: parameters.prompt ?? "",
    dialogue: parameters.dialogue ?? "",
    character: parameters.character ?? "",
    style: parameters.style ?? "",
    color: parameters.color ?? "",
    customTags: [...(parameters.customTags ?? [])],
  };
  return {
    parameters,
    image_url: imageUrl,
    storyboard,
  };
};
//////监听
onMounted(() => {
  updateResponsiveState();
  if (typeof window !== "undefined") {
    window.addEventListener("resize", updateResponsiveState);
  }
  console.log("开始监听分镜生成事件");
  loading.value = true;
  bus.on("storyboard-generated", handleStoryboard);
  bus.on("comic-generated", handleMeta);
});

onBeforeUnmount(() => {
  bus.off("storyboard-generated", handleStoryboard);
  bus.off("comic-generated", handleMeta);
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", updateResponsiveState);
  }
});

const handleStoryboard = (payload: unknown) => {
  const data = (payload as StoryboardPayload) ?? {};
  scene.value = data.scene ?? "";
  prompt.value = data.prompt ?? "";
  dialogue.value = data.dialogue ?? "";
  character.value = data.character ?? "";
  console.log("收到分镜数据:", data);
  loading.value = false;
  return data;
};

const handleMeta = (payload: unknown) => {
  const data = (payload as StoryboardMetaPayload) ?? {};
  style.value = data.style ?? "";
  color.value = data.color ?? "";
  customTags.value = Array.isArray(data.hints) ? data.hints : [];
  console.log("收到分镜元数据:", data);
  return data;
};
//////////////////////配置弹窗
const Configuration = ref(false);
const descriptionColumns = ref(3);
const isCompactScreen = ref(false);
const dialogWidth = computed(() => (isCompactScreen.value ? "96vw" : "500px"));
const uploadDialogVisible = ref(false);
const uploadFiles = ref<UploadUserFile[]>([]);
const uploadPreview = ref("");
const uploadDescription = ref("");
const uploadImageBase64 = ref("");
const uploadImageType = ref("");
const ConfigurationClick = async () => {
  Configuration.value = true;
};

const updateResponsiveState = () => {
  if (typeof window === "undefined") {
    descriptionColumns.value = 3;
    isCompactScreen.value = false;
    return;
  }
  const width = window.innerWidth;
  isCompactScreen.value = width < 768;
  if (width < 640) {
    descriptionColumns.value = 1;
  } else if (width < 1024) {
    descriptionColumns.value = 2;
  } else {
    descriptionColumns.value = 3;
  }
};

/////////////////////上传图片
function UploadImg() {
  uploadDialogVisible.value = true;
}

const handleUploadChange: UploadProps["onChange"] = (_, files) => {
  uploadFiles.value = [...files];
  const latest = files[files.length - 1];
  if (latest?.raw) {
    const reader = new FileReader();
    reader.onload = () => {
      const result = typeof reader.result === "string" ? reader.result : "";
      uploadPreview.value = result;
      if (result) {
        const match = result.match(/^data:(.+);base64,(.+)$/);
        if (match) {
          const mime = match[1];
          const typePart =
            mime.split("/")[1]?.split(";")[0]?.toLowerCase() || "png";
          uploadImageType.value = typePart;
          uploadImageBase64.value = match[2];
        } else {
          uploadImageType.value = "";
          uploadImageBase64.value = "";
        }
      }
    };
    reader.readAsDataURL(latest.raw);
  }
};

const handleUploadRemove: UploadProps["onRemove"] = (_, files) => {
  uploadFiles.value = [...files];
  if (!files.length) {
    uploadPreview.value = "";
    uploadImageBase64.value = "";
    uploadImageType.value = "";
  }
};

const handleUploadConfirm = () => {
  if (uploadImageBase64.value && !uploadDescription.value.trim()) {
    ElMessage.warning("请填写图像介绍");
    return;
  }
  uploadDescription.value = uploadDescription.value.trim();
  uploadDialogVisible.value = false;
};

//////////////////////////////////////url转base64
const last_image = "";
const last_image_type = "";

function urlToBase64(url: string): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = "Anonymous"; // 解决跨域问题
    img.onload = () => {
      const canvas = document.createElement("canvas");
      canvas.width = img.width;
      canvas.height = img.height;
      const ctx = canvas.getContext("2d");
      if (ctx) {
        ctx.drawImage(img, 0, 0);
        const dataURL = canvas.toDataURL("image/png");
        resolve(dataURL);
      } else {
        reject(new Error("无法获取Canvas上下文"));
      }
    };
    img.onerror = (err) => {
      reject(err);
    };
    img.src = url;
  });
}

async function imageUrlToBase64(url: string): Promise<string> {
  try {
    return await urlToBase64(url);
  } catch (error) {
    console.warn("Fallback to fetch for base64 conversion", error);
  }
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`无法下载图片: ${response.status}`);
  }
  const contentType = response.headers.get("content-type") ?? "image/png";
  const buffer = await response.arrayBuffer();
  const encoded = bufferToBase64(buffer);
  return `data:${contentType};base64,${encoded}`;
}

function bufferToBase64(buffer: ArrayBuffer): string {
  let binary = "";
  const bytes = new Uint8Array(buffer);
  for (let i = 0; i < bytes.length; i += 0x8000) {
    const chunk = bytes.subarray(i, i + 0x8000);
    binary += String.fromCharCode(...chunk);
  }
  return btoa(binary);
}

//转到生成漫画
async function goGenerateComic() {
  const parameters = buildParametersForPage();
  const comicData: ComicData = {
    StoryboardPayload: {
      scene: scene.value,
      prompt: prompt.value,
      dialogue: dialogue.value,
      character: character.value,
    },
    StoryboardMetaPayload: {
      style: style.value,
      color: color.value,
      hints: customTags.value,
    },
  };
  console.log("comdata数据:", comicData);
  try {
    loading.value = true;
    const result = await generateImage(comicData);
    const payload = {
      page_id: page_id++,
      title: novel.value || "\u672a\u547d\u540d\u7ae0\u8282",
      url: result.image_url,
      parameters,
      imageBase64: "",
    };
    console.log("Image generation result:", result);
    setComicImage([novel.value, result.image_url]);

    if (currentUser.value?.username) {
      try {
        const base64 = await imageUrlToBase64(result.image_url);
        payload.imageBase64 = base64;
        const metadata = buildMetadataForComic(parameters, result.image_url);
        const saved = await saveComicForCurrentUser({
          title: payload.title,
          pageNumber: payload.page_id,
          imageBase64: base64,
          imageUrl: result.image_url,
          metadata,
        });
        if (saved) {
          bus.emit("comic-persisted", {
            temp_id: payload.page_id,
            storage_id: saved.id,
          });
        }
      } catch (error) {
        console.error("Failed to persist comic", error);
        const metadata = buildMetadataForComic(parameters, result.image_url);
        const saved = await saveComicForCurrentUser({
          title: payload.title,
          pageNumber: payload.page_id,
          imageUrl: result.image_url,
          metadata,
        }).catch((err) => {
          console.error("Server side base64 conversion failed", err);
          return null;
        });
        if (saved) {
          bus.emit("comic-persisted", {
            temp_id: payload.page_id,
            storage_id: saved.id,
          });
        }
      }
    }

    router.push({ name: "home-comic" });
    loading.value = false;
    setTimeout(() => {
      bus.emit("make-imageover", payload);
    }, 1000);
  } catch (error) {
    const err = error as Error;
    console.error("图像生成失败:", err);
    ElMessage.error(`图像生成失败: ${err.message}`);
  }
}

// API 调用：生成图像
async function generateImage(ComicData: ComicData) {
  console.log("发送 /api/image 请求");
  console.log("图像生成输入数据:", ComicData);
  const targetBaseUrl = (UPDATE_BASE_URL.value || "").trim();
  const requestBody: Record<string, unknown> = {
    apiKey: (UPDATE_API_KEY.value || "").trim(),
    baseUrl: targetBaseUrl,
    model: UPDATE_IMAGE_MODEL.value,
    prompt: `以${style.value}风格，
    ${color.value}色调，
    ${customTags.value.join(",")}，
    创作一个包含场景:${scene.value}，
    剧情为${prompt.value}，人物:${character.value},人物对白:${
      dialogue.value
    }的一个极具创意的漫画页面排版，自由构图，非网格化布局，人物和背景元素跨越多个分镜格，复杂的页面设计；若有颜色形容词则必须按照${
      style.value
    }风格，
    ${color.value}色调来绘制，(比如若色调为黑白则必须为黑白，不可有多余的色彩)`,
    role: "user",
  };
  if (uploadImageBase64.value && uploadImageType.value) {
    requestBody.image = `data:image/${uploadImageType.value};base64,${uploadImageBase64.value}`;
    const desc = uploadDescription.value.trim();
    if (desc) {
      requestBody.image_description = desc;
    }
  } else if (last_image && last_image_type) {
    requestBody.image = `data:image/${last_image_type};base64,${last_image}`;
  }

  console.log("请求体内容:", requestBody);
  const response: Response = await fetch(`${BACK_URL}/api/image`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(requestBody),
  });
  if (!response.ok) {
    const errorData = await response.json();
    console.error("错误响应 /api/image:", errorData);
    throw new Error(`图像生成失败: ${errorData.error || response.statusText}`);
  }
  const dataResponse = await response.json();
  console.log("收到 /api/image 响应:", dataResponse);

  if (
    !dataResponse.data ||
    !dataResponse.data[0] ||
    !dataResponse.data[0].url
  ) {
    throw new Error("图像生成失败：响应中缺少有效的图像 URL");
  }
  var last_image = dataResponse.data[0].url;
  console.log("生成的图像 URL:", last_image);
  return { image_url: dataResponse.data[0].url };
}

////不知道原理
const size = ref<"default" | "small" | "large">("default");

const iconStyle = computed(() => {
  const marginMap = {
    large: "8px",
    default: "6px",
    small: "4px",
  };
  return {
    marginRight: marginMap[size.value] || marginMap.default,
  };
});
const blockMargin = computed(() => {
  const marginMap = {
    large: "32px",
    default: "28px",
    small: "24px",
  };
  return {
    marginTop: marginMap[size.value] || marginMap.default,
  };
});
defineProps<{ text?: string }>();
</script>
<style scoped>
.parameter-preview {
  width: 100%;
}

.parameter-preview :deep(.el-card__body) {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.el-descriptions {
  margin-top: 20px;
}
.cell-item {
  display: flex;
  align-items: center;
}
.margin-top {
  margin-top: 20px;
}
.page-preview {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.parameter-preview :deep(.el-descriptions__cell) {
  word-break: break-word;
}

.parameter-actions {
  width: 100%;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-bottom: 12px;
}

.card-footer-actions {
  display: flex;
  justify-content: flex-end;
  padding-right: 30px;
}

@media (max-width: 960px) {
  .parameter-preview :deep(.el-card__body) {
    padding: 8px;
  }

  .card-footer-actions {
    justify-content: center;
    padding-right: 0;
  }
}

@media (max-width: 640px) {
  .parameter-preview :deep(.el-descriptions__label) {
    align-items: flex-start;
  }

.parameter-preview :deep(.el-input__inner),
.parameter-preview :deep(.el-textarea__inner) {
  font-size: 14px;
}
}

.upload-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.upload-dialog-tip {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.upload-section h6 {
  margin: 0 0 8px;
}

.upload-block {
  width: 100%;
}

.upload-preview {
  margin-top: 12px;
  border: 1px dashed var(--el-border-color);
  padding: 8px;
  border-radius: 8px;
  text-align: center;
}

.upload-preview img {
  max-width: 100%;
  height: auto;
  display: inline-block;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
