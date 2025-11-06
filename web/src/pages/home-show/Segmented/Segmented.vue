<template>
  <div class="card-container">
    <el-card class="card card--wide">
      <template #header>
        <div class="card-header">
          <span>手动</span>
        </div>
      </template>
      <!-- <p class="text item" style="height: 250px">{{ textParagraph }}</p> -->
      <el-input
        v-model="textParagraph"
        :autosize="{ minRows: 10, maxRows: 20 }"
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
          <span>自动</span>
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
          placeholder="请输入提示词"
          aria-label="Please click the Enter key after input"
        />
      </p>
      <template #footer>
        <button type="button" class="btn btn-outline-info">
          <i class="fas fa-film"></i> 提取分镜
        </button>
      </template>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import bus from "../eventBus.js";
import { sharedText } from "../shared-text";

type TaskInfo = {
  status?: string;
  time?: string;
  message?: string;
};

const props = defineProps<{ text?: string }>();

const defaultParagraph =
  "这是一个示例段落，用于展示自动分段功能。系统会在保留标点和语义的前提下，将文本按照合适的长度拆分，帮助你快速检视内容。";
const textParagraph = ref(defaultParagraph);
const segments = ref<string[]>([]);
const currentIndex = ref(0);

const other_hint = ref<string[]>();
const color = ref("");
const style = ref("");
const style_options = [
  { value: "写实", label: "写实" },
  { value: "普通", label: "普通" },
  { value: "赛博朋克", label: "赛博朋克" },
  { value: "幻想", label: "幻想" },
];
const color_options = [
  { value: "黑暗", label: "黑暗" },
  { value: "暖色", label: "暖色" },
  { value: "冷色", label: "冷色" },
  { value: "水彩", label: "水彩" },
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

import config from"../../config.json"
const TEXT_MODEL=config.TEXT_MODEL;


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
    model: TEXT_MODEL,
    messages: [
      {
        role: "system",
        content:
          "你是一个漫画创作者。将输入文本转换为1个图片的漫画，要求含有场景，提示词和人物(若有的话依照人物名称生成符合该人物的服饰，长相，没有就返回null)，输出格式为纯文本，字段按以下格式：scene:...;prompt:...;character:...;字段间用分号分隔，不包含换行符，用中文回答我。",
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
    const response = await fetch("http://localhost:3000/api/text", {
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
      character: "",
    };

    const sceneMatch = textContent.match(/scene:([\s\S]*?)(?=;prompt:|$)/);
    const promptMatch = textContent.match(/prompt:([\s\S]*?)(?=;character:|$)/);
    const characterMatch = textContent.match(/character:([\s\S]*?)(?=;|$)/);

    console.log("解析结果:", {
      scene: sceneMatch ? sceneMatch[1] : null,
      prompt: promptMatch ? promptMatch[1] : null,
      character: characterMatch ? characterMatch[1] : null,
    });
    if (sceneMatch && sceneMatch[1]) {
      parsedData.scene = sceneMatch[1].trim();
    }
    if (promptMatch && promptMatch[1]) {
      parsedData.prompt =
        promptMatch[1].trim() +
        "Output in a style similar to that of Japanese anime";
    }
    if (characterMatch && characterMatch[1] && characterMatch[1] !== "null") {
      parsedData.character = characterMatch[1].trim();
    }
    //上一张图片（为啥会在这里？）
    // if (last_image !== null) {
    //   parsedData.init_image = last_image;
    // }
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
