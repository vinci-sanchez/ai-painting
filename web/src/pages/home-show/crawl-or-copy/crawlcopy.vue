<template>
  <div class="crawl-page">
    <el-tabs v-model="activeName">
        <!-- 爬取小说 -->
        <el-tab-pane label="爬取小说" name="first">
          <section id="crawl-section" class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">小说爬取</h2>
              <p class="text-muted text-center">
                支持从比如 <a href="https://www.blqukk.cc/" target="_blank" rel="noopener noreferrer">https://www.blqukk.cc/</a> 爬取小说章节(因为服务器在国外，网站可能会出现404错误限制访问的情况，请酌情使用)
              </p>
              <div class="input-group mb-3 input-group--responsive">
                <span class="input-group-text" style="overflow: hidden;">
                  <i class="fas fa-link"></i>
               
                <input
                  type="text"
                  id="novelUrl"
                  class="form-control"
                  placeholder="请输入小说目录 URL"
                  aria-label="小说目录 URL"
                /> </span>
                <el-input-number
                  v-model="num"
                  :min="1"
                  type="number"
                  id="index"

                  placeholder="请输入小说章节数"
                >
                  <template #suffix> <span>章</span></template>
                </el-input-number>
          
                <!-- <button v-on:click="crawl" class="btn btn-primary" style="max-width: 150px;" id="crawl">
                  <i class="fas fa-spider"></i> 开始爬取
                </button>  -->
          
              </div>
              <el-button type="primary" v-on:click="crawl"  id="crawl" style=" border-radius: 5px;  display: block;height:35px; width: fit-content; margin-left: auto;"><i class="fas fa-spider"></i> 开爬！</el-button>
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
          <section class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">TXT 文件上传</h2>
              <p class="text-muted text-center mb-4">
                拖动或选择 .txt 小说文件，系统会自动填充内容并同步到分段页面
              </p>
              <div
                class="upload-dropzone"
                :class="{ 'upload-dropzone--active': isDragging }"
                @dragover.prevent="handleDragOver"
                @dragleave.prevent="handleDragLeave"
                @drop.prevent="handleFileDrop"
              >
                <i class="fas fa-file-alt upload-dropzone__icon"></i>
                <p class="mb-1">将 TXT 文件拖拽到这里</p>
                <p class="text-muted mb-3">或</p>
                <button
                  type="button"
                  class="btn btn-outline-primary"
                  @click="triggerFileDialog"
                >
                  选择 TXT 文件
                </button>
                <p v-if="selectedFileName" class="text-success mt-3">
                  已选择：{{ selectedFileName }}
                </p>
                <p v-if="uploadError" class="text-danger mt-2">
                  {{ uploadError }}
                </p>
              </div>
              <input
                ref="fileInputRef"
                type="file"
                accept=".txt,text/plain"
                class="upload-file-input"
                @change="handleFileChange"
              />
            </div>
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
const isDragging = ref(false);
const uploadError = ref("");
const selectedFileName = ref("");
const fileInputRef = ref<HTMLInputElement | null>(null);

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
    const index = num.value - 1;
    const response = await fetch(`${config.Crawle_url}`,
     // `${config.BACK_URL}/api/crawl`
       {
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
        resultDiv.innerHTML += `<h3>HTTP Error</h3><p>Received status ${response.status} from the server. Please check the URL and try again.</p><hr>`;
      }
      throw new Error(`HTTP ${response.status}`);
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
    // if (resultDiv) {
    //   if (num.value < data.total_chapters) {
    //     resultDiv.innerHTML += `<button id="continue-crawl" class="btn btn-primary">继续爬取</button>`;
    //     const continueBtn = document.getElementById("continue-crawl");
    //     if (continueBtn) {
    //       continueBtn.addEventListener("click", crawl);
    //     }
    //   } else {
    //     resultDiv.innerHTML += "<p>已爬取全部章节</p>";
    //   }
    // }

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

const isTxtFile = (file: File) =>
  file.type === "text/plain" || file.name.toLowerCase().endsWith(".txt");

const resetFileInput = () => {
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
};

const processTxtFile = async (file: File) => {
  uploadError.value = "";
  if (!isTxtFile(file)) {
    uploadError.value = "仅支持上传扩展名为 .txt 的纯文本文件";
    ElMessage.error(uploadError.value);
    resetFileInput();
    return;
  }

  try {
    const content = await file.text();
    if (!content.trim()) {
      uploadError.value = "文件内容为空，请确认文件是否正确";
      ElMessage.warning(uploadError.value);
      return;
    }
    selectedFileName.value = file.name;
    textContent.value = content;
    finishTask(content);
    ElMessage.success(`已读取 ${file.name}`);
  } catch (error) {
    uploadError.value = `文件读取失败: ${(error as Error).message}`;
    ElMessage.error(uploadError.value);
  } finally {
    resetFileInput();
  }
};

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  const [file] = target.files ?? [];
  if (file) {
    processTxtFile(file);
  }
};

const handleDragOver = (event: DragEvent) => {
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "copy";
  }
  isDragging.value = true;
};

const handleDragLeave = () => {
  isDragging.value = false;
};

const handleFileDrop = (event: DragEvent) => {
  event.preventDefault();
  const [file] = event.dataTransfer?.files ?? [];
  if (file) {
    processTxtFile(file);
  }
  isDragging.value = false;
};

const triggerFileDialog = () => {
  uploadError.value = "";
  fileInputRef.value?.click();
};

</script>

<style scoped>
.crawl-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.input-group--responsive {
  flex-wrap: wrap;
  gap: 12px;
  align-items: stretch;
}

.input-group--responsive > * {
  flex: 1 1 auto;
}

/* .input-group--responsive .form-control {
  min-width: 220px;
} */

.input-group--responsive .el-input-number {
  max-width: 180px;
}

.upload-dropzone {
  border: 2px dashed #c0c4cc;
  border-radius: 12px;
  padding: 32px 16px;
  text-align: center;
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.upload-dropzone--active {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.upload-dropzone__icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 8px;
}

.upload-file-input {
  display: none;
}

@media (max-width: 768px) {
  .input-group--responsive {
    flex-direction: column;
  }

  .input-group--responsive > * {
    width: 100% !important;
    flex: 1 1 auto;
  }

  .input-group--responsive button,
  .input-group--responsive .el-input-number {
    width: 100%;
  }

  .upload-dropzone {
    padding: 24px 12px;
  }
}
</style>
