<template>
  <div>
    <div>
      <el-tabs v-model="activeName">
        <!-- 爬取小说 -->
        <el-tab-pane label="爬取小说" name="first">
          <section id="crawl-section" class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">小说爬取</h2>
              <p class="text-muted text-center">
                支持从比如 https://www.blqukk.cc/ 爬取小说章节(因为服务器在国外，网站可能会出现404错误限制访问的情况，请酌情使用)
              </p>
              <div class="input-group mb-3">
                <span class="input-group-text">
                  <i class="fas fa-link"></i>
                </span>
                <input
                  type="text"
                  id="novelUrl"
                  class="form-control"
                  placeholder="请输入小说目录 URL"
                  aria-label="小说目录 URL"
                />
                <el-input-number
                  v-model="num"
                  :min="1"
                  type="number"
                  id="index"
                  style="width: 200px"
                  placeholder="请输入小说章节数量"
                >
                  <template #suffix> <span>章</span></template>
                </el-input-number>
                <button v-on:click="crawl" class="btn btn-primary" id="crawl">
                  <i class="fas fa-spider"></i> 开始爬取
                </button>
              </div>
              <div
                v-show="crawl_result"
                id="result"
                class="mt-4 p-3 bg-light rounded"
                style="max-height: 300px; overflow-y: auto"
              ></div>
            </div>
          </section>
        </el-tab-pane>
        <!-- 输入小说 -->
        <el-tab-pane label="输入小说" name="second">
          <section id="input-section" class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">小说文本输入</h2>
              <p class="text-muted text-center">将小说文本转换为段落内容</p>
              <div class="mb-3">
                <label for="textInput" class="form-label">小说文本</label>
                <textarea
                  class="form-control"
                  id="textInput"
                  v-model="textContent"
                  rows="8"
                  placeholder="请输入小说文本，例如：&#10;第一章 森林冒险&#10;小猫踏入森林深处的道路..."
                  aria-label="小说文本"
                >
                </textarea>
              </div>
            </div>
          </section>
        </el-tab-pane>
        <el-tab-pane label="上传文件" name="third">
          <section id="input-section" class="card mb-5 shadow-lg">
            <div class="card-body">暂无内容</div>
          </section>
        </el-tab-pane>
        <button
          type="button"
          class="btn btn-outline-secondary"
          :disabled="!canNavigate"
          @click="goToSegmented"
        >
          <i class="fas fa-cut"></i> 去分段
        </button>
      </el-tabs>
    </div>
  </div>
</template>

<script lang="ts" setup>
defineOptions({ name: "crawlcopy" });
import { computed, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { useRouter } from "vue-router";

import config from "../../config.json";
import bus from "../eventbus.js";
import { setSharedText, sharedText } from "../shared-text";

type Chapter = { title: string; content: string };
type CrawlResponse = {
  chapters: Chapter[];
  next_index: number;
  total_chapters: number;
  error?: string;
};

const activeName = ref("first");
const crawl_result = ref(true);
const num = ref(1);

const isDone = ref(false);
const router = useRouter();
const textContent = ref(sharedText.value);

const canNavigate = computed(() => textContent.value.trim().length > 0);

const goToSegmented = () => {
  if (!canNavigate.value) {
    ElMessage.warning("请先输入内容或完成爬取");
    return;
  }
  setSharedText(textContent.value);
  router.push({ name: "home-segmented" });
};

function finishTask(input?: string) {
  isDone.value = true;
  const info = {
    status: "done",
    time: new Date().toLocaleTimeString(),
    message: input ?? "",
  };

  console.log("A 触发任务完成事件: ", info);

  // 全局广播信息
  bus.emit("task-finished", info);
  textContent.value = input ?? "";
  setSharedText(textContent.value);
}

//爬取小说

//let nextIndex = 0;
async function crawl() {
  const urlInput = document.getElementById("novelUrl") as HTMLInputElement;
  const resultDiv = document.getElementById("result");
  const errorMessage = document.getElementById("error-message");

  if (!urlInput || !urlInput.value) {
    ElMessage("请输入有效的 URL");
    return;
  }
  try {
    const url = urlInput.value;
    let index = num.value - 1;
    const response = await fetch(`${config.BACK_URL}/api/crawl`, {
      //http://localhost:3000/api/crawl
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        novel_url: url,
        start_index: index,
        limit: 1,
      }),
    });

    if (!response.ok) {
      if (resultDiv) {
        resultDiv.innerHTML += `<h3>HTTP错误</h3><p>状态码 ${response.status}，抓取章节失败: 未找到对应的章节链接，可能是站点结构变化或 URL 无效</p><hr>`;
      }
      throw new Error(`HTTP错误状态码 ${response.status}`);
    }

    const data = (await response.json()) as CrawlResponse;
    console.log("API响应: ", data);

    if (data.error) {
      if (errorMessage) {
        errorMessage.textContent = `错误: ${data.error}`;
        errorMessage.classList.remove("d-none");
      }
      return;
    }

    if (!data.chapters || !Array.isArray(data.chapters)) {
      if (errorMessage) {
        errorMessage.textContent = "错误: 未接收到有效章节数据";
        errorMessage.classList.remove("d-none");
      }
      return;
    }

    // 更新 num
    num.value = data.next_index;
    console.log("num", num.value);
    console.log("data.next_index", data.next_index);

    // 更新 resultDiv 显示当前爬取的章节
    if (resultDiv) {
      resultDiv.innerHTML = "";
      data.chapters.forEach((ch) => {
        resultDiv.innerHTML += `<h3>${ch.title}</h3><p>${ch.content}</p><hr>`;
      });
    }

    // 如果还有后续章节则追加按钮
    if (resultDiv) {
      if (num.value < data.total_chapters) {
        resultDiv.innerHTML += `<button id="continue-crawl" class="btn btn-primary">继续爬取</button>`;
        const continueBtn = document.getElementById("continue-crawl");
        if (continueBtn) {
          continueBtn.addEventListener("click", crawl);
        }
      } else {
        resultDiv.innerHTML += "<p>已爬取全部章节</p>";
      }
    }

    // 爬取成功后自动填充输入框
    textContent.value = data.chapters
      .map((ch) => `${ch.title}\n${ch.content}`)
      .join("\n");
    finishTask(textContent.value);
  } catch (error) {
    const err = error as Error;
    console.error("爬取出错", err);
    if (errorMessage) {
      errorMessage.textContent = `爬取失败: ${err.message}`;
      errorMessage.classList.remove("d-none");
    }
  }
}

watch(
  () => textContent.value,
  (value) => {
    setSharedText(value);
  }
);
</script>
