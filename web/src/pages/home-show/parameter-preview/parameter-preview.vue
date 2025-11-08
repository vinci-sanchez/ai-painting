<template>
  <el-card style="width: 100%">
    <template #header>
      <div class="card-header">
        <span>分镜内容</span>
      </div>
    </template>
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
        <div>{{ now_time }}点</div>
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
    <template #footer
      ><el-button type="primary" style="margin-right: 30px"
        >123</el-button
      ></template
    >
  </el-card>
</template>
<script lang="ts" setup>
import { computed, ref, onMounted, onBeforeUnmount } from "vue";
import { ElMessage } from "element-plus";
import bus from "../eventBus.js";
import {
  Iphone,
  Location,
  OfficeBuilding,
  Tickets,
  User,
} from "@element-plus/icons-vue";
import { setComicText, comicText } from "../shared-text";

const novel = ref("");

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

const handleStoryboard = (payload: unknown) => {
  const data = (payload as StoryboardPayload) ?? {};
  scene.value = data.scene ?? "";
  prompt.value = data.prompt ?? "";
  character.value = data.character ?? "";
  console.log("收到分镜数据:", data);
};

const handleMeta = (payload: unknown) => {
  const data = (payload as StoryboardMetaPayload) ?? {};
  style.value = data.style ?? "";
  color.value = data.color ?? "";
  customTags.value = Array.isArray(data.hints) ? data.hints : [];
  console.log("收到分镜元数据:", data);
};

onMounted(() => {
  console.log("开始监听分镜生成事件");
  bus.on("storyboard-generated", handleStoryboard);
  bus.on("comic-generated", handleMeta);
});

onBeforeUnmount(() => {
  bus.off("storyboard-generated", handleStoryboard);
  bus.off("comic-generated", handleMeta);
});

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
</style>
