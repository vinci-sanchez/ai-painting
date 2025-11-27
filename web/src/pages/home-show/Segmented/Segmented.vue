<template>
  <div class="card-container">
    <el-card class="card card--wide">
      <template #header>
        <div class="card-header">
          <span>调整分段</span>
        </div>
      </template>
      <!-- <p class="text item" style="height: 250px">{{ textParagraph }}</p> -->
      <el-input
        v-model="textParagraph"
        :autosize="{ minRows: 15, maxRows: 15 }"
        type="textarea"
        placeholder="Please input"
      />
      <template #footer>
        <button
          type="button"
          class="btn btn-outline-secondary"
          @click="showPreviousSegment"
        >
          <i class="fas fa-cut"></i> 上一段
        </button>
        <button
          type="button"
          class="btn btn-outline-secondary"
          @click="showNextSegment"
        >
          <i class="fas fa-cut"></i> 下一段
        </button>
      </template>
    </el-card>

    <el-card class="card card--narrow">
      <template #header>
        <div class="card-header">
          <span>参数</span>
        </div>
      </template>

      <p>风格</p>
      <p>
        <el-select v-model="style" placeholder="选择风格" style="width: 240px">
          <el-option
            v-for="item in style_options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </p>
      <p>色调</p>
      <p>
        <el-select v-model="color" placeholder="选择色调" style="width: 240px">
          <el-option
            v-for="item in color_options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </p>
      <p>其他提示词</p>
      <p>
        <el-input-tag
          v-model="other_hint"
          style="width: 240px"
          placeholder="请输入提示词(可填可不填)"
          aria-label="Please click the Enter key after input"
        />
      </p>
      <template #footer>
        <button
          type="button"
          class="btn btn-outline-info"
          @click="gotoparameter_preview"
        >
          <i class="fas fa-film"></i> 提取分镜
        </button>
      </template>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
defineOptions({ name: "Segmented" });
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import bus from "../eventbus.js";
import {
  sharedText,
  setSharedText,
  setStoryboard,
  Storyboard,
} from "../shared-text";
import { comicText, setComicText } from "../shared-text";
import config from "../../config.json";
import { preview } from "vite";
import { te } from "element-plus/es/locale/index.mjs";
import router from "../../../router.js";
const TEXT_MODEL = config.TEXT_MODEL;
const BACK_URL = config.BACK_URL;
const API_KEY = config.API_KEY;
type TaskInfo = {
  status?: string;
  time?: string;
  message?: string;
};
type ComicInfo = {
  style?: string;
  color?: string;
  hints?: string[];
};

const props = defineProps<{ text?: string }>();

const defaultParagraph =
  "这是一个示例段落，用于展示自动分段功能。系统会在保留标点和语义的前提下，将文本按照合适的长度拆分，帮助你快速检视内容。";
const textParagraph = ref(defaultParagraph);
const segments = ref<string[]>([]);
const currentIndex = ref(0);

const other_hint = ref<string[]>();
const color = ref("彩绘");
const style = ref("日式漫画");

const style_options = [
  { value: "日式漫画", label: "日式漫画" },
  { value: "美式漫画", label: "美式漫画" },
  { value: "中式漫画", label: "中式漫画" },
  { value: "赛博朋克", label: "赛博朋克" },
  { value: "线描", label: "线描" },
];
const color_options = [
  { value: "彩绘", label: "彩绘" },
  { value: "暖色", label: "暖色" },
  { value: "冷色", label: "冷色" },
  { value: "黑白", label: "黑白" },
];

function showSegment(index: number) {
  if (segments.value.length === 0) {
    textParagraph.value = defaultParagraph;
    return;
  }
  const safeIndex = Math.min(Math.max(index, 0), segments.value.length - 1);
  currentIndex.value = safeIndex;
  textParagraph.value = segments.value[safeIndex];
}

function showPreviousSegment() {
  if (currentIndex.value > 0) {
    showSegment(currentIndex.value - 1);
  }
}

function showNextSegment() {
  if (currentIndex.value < segments.value.length - 1) {
    showSegment(currentIndex.value + 1);
  }
}

function handleTask(payload: unknown) {
  const info = (payload ?? {}) as TaskInfo;
  console.log("B 收到任务完成信息:", info);
  startWork(info);
}

function startWork(info: TaskInfo) {
  console.log(
    `B 开始处理任务，状态：${info.status}，时间：${info.time}，内容长度：${
      info.message?.length ?? 0
    }`
  );
  segmentText(info.message ?? "");
}

function segmentText(inputText: string) {
  const trimmed = inputText.trim();
  if (!trimmed) {
    ElMessage("请输入小说文本");
    segments.value = [];
    showSegment(0);
    return;
  }

  try {
    const cleanedText = trimmed
      .replace(/\n+/g, " ")
      .replace(/\s+/g, " ")
      .trim();

    const rawSegments = cleanedText
      .split(/(?<=[.!?。！？])\s+|(?<=[\'"])\s+/)
      .map((s) => s.trim())
      .filter(Boolean);

    const nextSegments: string[] = [];
    let currentSegment = "";
    const minSegmentLength = 200;
    const maxSegmentLength = 400;

    for (const part of rawSegments) {
      const combined = `${currentSegment}${part}`;

      if (combined.length >= maxSegmentLength) {
        nextSegments.push(combined.trim());
        currentSegment = "";
        continue;
      }

      if (combined.length >= minSegmentLength) {
        nextSegments.push(combined.trim());
        currentSegment = "";
      } else {
        currentSegment = `${combined} `;
      }
    }

    if (currentSegment.trim().length > 0) {
      nextSegments.push(currentSegment.trim());
    }

    segments.value = nextSegments.length > 0 ? nextSegments : [cleanedText];
    showSegment(0);
    console.log("分段结果:", segments.value);
  } catch (error) {
    console.error("分段失败:", error);
    ElMessage.error(`分段失败: ${(error as Error).message}`);
  }
}

onMounted(() => {
  bus.on("task-finished", handleTask);
  console.log("B 开始监听任务完成事件");
  if (sharedText.value.trim()) {
    segmentText(sharedText.value);
  }
});

onBeforeUnmount(() => {
  bus.off("task-finished", handleTask);
  console.log("B 停止监听任务完成事件");
});

watch(
  () => sharedText.value,
  (value) => {
    if (value.trim()) {
      segmentText(value);
    }
  }
);

const emitStoryboard = (data: {
  scene?: string;
  prompt?: string;
  character?: string;
}) => {
  console.log("分镜数据:", data);

  setStoryboard(Storyboard.value);
  bus.emit("storyboard-generated", data);
};

const emitComic = (data: {
  style?: string;
  color?: string;
  hints?: string[];
}) => {
  console.log("漫画类型数据:", data);
  setComicText(comicText.value);
  bus.emit("comic-generated", data);
};

/////跳转
async function gotoparameter_preview() {
  if (!textParagraph.value || textParagraph.value.trim().length === 0) {
    ElMessage.warning("请确认文本后再生成分镜");
    return;
  }
  router.push({ name: "home-parameter-preview" });
  await generateStoryboard(textParagraph.value);
  console.log("自动分镜生成完成");
}
/////分镜
async function generateStoryboard(text?: string) {
  const promptTags = document.getElementById("promptTags");
  const prompts = promptTags
    ? Array.from(promptTags.querySelectorAll(".badge")).map((tag) =>
        tag.textContent.trim().replace(/×$/, "").trim()
      )
    : [];
  const extraPrompts =
    prompts.length > 0 ? `, 附加提示词: ${prompts.join(", ")}` : "";
  const requestBody = {
    apiKey: API_KEY,
    model: TEXT_MODEL,
    messages: [
      {
        role: "system",
        content: `你是一个漫画创作者。将输入文本转换为1个图片的四格漫画,要求含有场景,提示词,人物对话(格式为谁在什么时候说了什么)和人物长相(若有的话依照人物名称生成符合该人物的服饰,长相,没有就返回null);若有颜色则必须符合风格：${style.value},色调：${color.value};输出格式为纯文本,字段按以下格式:scene:...;prompt:...;dialogue:....;character:...;字段间用分号分隔，不包含换行符，用中文回答我。`,
      },
      { role: "user", content: text + extraPrompts },
    ],
    max_tokens: 300,
    temperature: 0.7,
  };
  console.log("发送 /api/text 请求体:", JSON.stringify(requestBody, null, 2));

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 30000); // 30秒超时
    const response = await fetch(`${BACK_URL}/api/text`, {
      //http://localhost:3000
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(requestBody),
      signal: controller.signal,
    });
    clearTimeout(timeoutId);
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      console.error("错误响应 /api/text:", errorData);
      throw new Error(
        `分镜生成失败: ${errorData.error || response.statusText}`
      );
    }
    const data = await response.json();
    console.log("收到 /api/text 响应:", JSON.stringify(data, null, 2));

    if (!data.data || !data.data.message) {
      throw new Error("后端响应格式错误，缺少 message 字段");
    }

    // 解析纯文本
    const textContent = data.data.message;
    const parsedData = {
      scene: "默认场景",
      prompt: "默认提示词",
      dialogue: "默认对话",
      character: "",
    };
    const comicData = {
      style: style.value,
      color: color.value,
      hints: other_hint.value || [],
    };
    const sceneMatch = textContent.match(/scene:([\s\S]*?)(?=;prompt:|$)/);
    const promptMatch = textContent.match(/prompt:([\s\S]*?)(?=;dialogue:|$)/);
    const dialogueMatch = textContent.match(
      /dialogue:([\s\S]*?)(?=;character:|$)/
    );
    const characterMatch = textContent.match(/character:([\s\S]*?)(?=;|$)/);

    console.log("解析结果:", {
      scene: sceneMatch ? sceneMatch[1] : null,
      prompt: promptMatch ? promptMatch[1] : null,
      character: characterMatch ? characterMatch[1] : null,
      dialogue: dialogueMatch ? dialogueMatch[1] : null,
    });
    if (sceneMatch && sceneMatch[1]) {
      parsedData.scene = sceneMatch[1].trim();
    }
    if (promptMatch && promptMatch[1]) {
      parsedData.prompt = promptMatch[1].trim();
    }
    if (dialogueMatch && dialogueMatch[1]) {
      parsedData.dialogue = dialogueMatch[1].trim();
    }
    if (characterMatch && characterMatch[1] && characterMatch[1] !== "null") {
      parsedData.character = characterMatch[1].trim();
    }
    //上一张图片（为啥会在这里？）
    // if (last_image !== null) {
    //   parsedData.init_image = last_image;
    // }
    ////////////////////////////发送总线信号
    emitStoryboard(parsedData);
    emitComic(comicData);

    return { data: parsedData };
  } catch (error) {
    console.error("generateStoryboard 错误:", error);
    throw error;
  }
}
</script>

<style>
.card-container {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}
.card--wide {
  width: 55vw;
}

.card--narrow {
  width: 30vw;
}
</style>
