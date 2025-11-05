<template>
  <div class="card-container">
    <el-card class="card card--wide">
      <template #header>
        <div class="card-header">
          <span>分段</span>
        </div>
      </template>
      <p class="text item" style="height: 250px">{{ ` ${textParagraph}` }}</p>
      <template #footer
        ><button type="button" class="btn btn-outline-secondary">
          <i class="fas fa-cut"></i> 下一段
        </button>
        <button type="button" class="btn btn-outline-secondary">
          <i class="fas fa-cut"></i> 上一段
        </button>
      </template>
    </el-card>

    <el-card class="card card--narrow">
      <template #header>
        <div class="card-header">
          <span>自定义</span>
        </div>
      </template>

      <p>风格</p>
      <p>
        <el-select v-model="style" placeholder="风格" style="width: 240px">
          <el-option
            v-for="item in style_options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </p>
      <p>色彩</p>
      <p>
        <el-select v-model="color" placeholder="色彩" style="width: 240px">
          <el-option
            v-for="item in color_options"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </p>
      <p>其他</p>
      <p>
        <el-input-tag
          v-model="other_hint"
          style="width: 240px"
          placeholder="其他提示"
          aria-label="Please click the Enter key after input"
        />
      </p>
      <template #footer
        ><button type="button" class="btn btn-outline-info">
          <i class="fas fa-film"></i> 提取分镜和内容
        </button></template
      >
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount } from "vue";
import { ElMessage } from "element-plus";
import bus from "../eventBus.js";
// 定义当 A 组件任务完成时要执行的函数
function handleTask(info) {
  console.log("🚀 B组件收到信息：", info);
  startWork(info);
}

function startWork(info) {
  console.log(
    `B组件开始执行任务！任务状态：${info.status}，时间：${info.time}${info.messsage}`
  );
  segmentText(info.message);
}

onMounted(() => {
  bus.on("task-finished", handleTask);
  console.log("B组件已开始监听 A 的任务事件");
});

onBeforeUnmount(() => {
  bus.off("task-finished", handleTask);
  console.log("B组件已停止监听");
});

const props = defineProps<{ text?: string }>();

const textParagraph = ref(
  "这是一个示例段落，用于展示自动分段功能。系统会根据标点符号和段落长度将文本分割成多个适合阅读的小段落，从而提升用户的阅读体验。"
);

const other_hint = ref<string[]>();

const color = ref("");
const style = ref("");
const style_options = [
  {
    value: "写实",
    label: "写实",
  },
  {
    value: "卡通",
    label: "卡通",
  },
  {
    value: "赛博朋克",
    label: "赛博朋克",
  },
  {
    value: "蒸汽波",
    label: "蒸汽波",
  },
];
const color_options = [
  {
    value: "黑白",
    label: "黑白",
  },
  {
    value: "彩色",
    label: "彩色",
  },
  {
    value: "手绘",
    label: "手绘",
  },
  {
    value: "水彩",
    label: "水彩",
  },
];

function segmentText(inputText: string) {
  if (!inputText) {
    ElMessage("请输入小说文本");
    return;
  }

  try {
    let cleanedText = inputText
      .replace(/\n+/g, " ")
      .replace(/\s+/g, " ")
      .trim();

    const rawSegments = cleanedText
      .split(/(?<=[.!?。！？])[\s]+|(?<=[\'"])[\s]+/)
      .map((s) => s.trim())
      .filter(Boolean);

    const segments: string[] = [];
    let currentSegment = "";
    const minSegmentLength = 200;
    const maxSegmentLength = 400;

    for (const part of rawSegments) {
      const combined = currentSegment + part;

      if (combined.length >= maxSegmentLength) {
        segments.push(combined.trim());
        currentSegment = "";
        continue;
      }

      if (combined.length >= minSegmentLength) {
        segments.push(combined.trim());
        currentSegment = "";
      } else {
        currentSegment = `${combined} `;
      }
    }

    if (currentSegment.trim().length > 0) {
      segments.push(currentSegment.trim());
    }

    textParagraph.value = segments[0] || cleanedText;
    console.log("分段结果:", segments);
  } catch (error) {
    console.error("分段失败:", error);
    ElMessage.error(`分段失败: ${(error as Error).message}`);
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
