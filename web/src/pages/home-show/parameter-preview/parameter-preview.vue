<template>
  <el-card style="width: 100%">
    <el-descriptions
      class="margin-top"
      title="项目内容"
      :column="3"
      :size="size"
      border
    >
      <template #extra>
        <el-button type="primary">Operation</el-button>
      </template>
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
              <iphone />
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
              <tickets />
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
        :column="3"
        :size="size"
        :style="blockMargin"
      >
        <template #extra>
          <el-button type="primary">Operation</el-button>
        </template>
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
import { computed, ref, onMounted, onBeforeUnmount, toRaw } from "vue";
import { ElMessage } from "element-plus";
import bus from "../eventbus.js";
import {
  Iphone,
  Location,
  OfficeBuilding,
  Tickets,
  User,
} from "@element-plus/icons-vue";
import { setComicText, comicText } from "../shared-text";
import { setComicImage, comicimage } from "../shared-text";
import router from "../../../router.js";
import config from "../../config.json";
// const TEXT_MODEL = config.TEXT_MODEL;
// const BACK_URL = config.BACK_URL;

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

/////分镜接受
const scene = ref("");
const prompt = ref("");
const character = ref("");
const style = ref("");
const color = ref("");
const customTags = ref<string[]>([]);
type StoryboardPayload = {
  scene?: string;
  prompt?: string;
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
//////监听
onMounted(() => {
  console.log("开始监听分镜生成事件");
  loading.value = true;
  bus.on("storyboard-generated", handleStoryboard);
  bus.on("comic-generated", handleMeta);
});

onBeforeUnmount(() => {
  loading.value = false;
  bus.off("storyboard-generated", handleStoryboard);
  bus.off("comic-generated", handleMeta);
});

const handleStoryboard = (payload: unknown) => {
  const data = (payload as StoryboardPayload) ?? {};
  scene.value = data.scene ?? "";
  prompt.value = data.prompt ?? "";
  character.value = data.character ?? "";
  console.log("收到分镜数据:", data);
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

//转到生成漫画
async function goGenerateComic() {
  const comicData: ComicData = {
    StoryboardPayload: {
      scene: scene.value,
      prompt: prompt.value,
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
    router.push({ name: "home-comic" });
    const payload = {
      page_id: page_id++,
      title: novel.value || "未命名章节",
      url: result.image_url,
    };
    console.log("图像生成结果:", result);
    setComicImage([novel.value, result.image_url]);
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
  const requestBody = {
    model: "ep-20251021153509-xh86n",
    prompt: `以${style.value}风格，
    ${color.value}色调，
    ${customTags.value.join(",")}，
    创作一个包含场景:${scene.value}，
    剧情为${prompt.value}，人物:${character.value}的四格漫画`,
    image: `data:image/${last_image_type};base64,${last_image}`,
    role: "user",
  };
  console.log("请求体内容:", requestBody);
  const response: Response = await fetch(`${config.BACK_URL}/api/image`, {
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
.card-footer-actions {
  display: flex;
  justify-content: flex-end;
  padding-right: 30px;
}
</style>
