<template>
  <div>
    <div>
      <el-tabs  v-model="activeName">
        <!-- 爬取小说 -->
        <el-tab-pane label="爬取小说" name="first">
          <section id="crawl-section" class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">小说爬取</h2>
              <p class="text-muted text-center">
                支持从笔趣看（https://www.blqukk.cc/）爬取小说章节
              </p>
              <div class="input-group mb-3">
                <span class="input-group-text">
                  <i class="fas fa-link"></i>
                </span>
                <input
                  type="text"
                  id="novelUrl"
                  class="form-control"
                  placeholder="请输入小说目录URL"
                  aria-label="小说目录URL"
                />
                <el-input-number
                  v-model="num"
                  :min="1"
                  type="number"
                  id="index"
                  style="width: 200px"
                  placeholder="请输入小说章节索引"
                >
                  <template #suffix> <span>索引</span></template>
                </el-input-number>
                <button v-on:click="crawl" class="btn btn-primary" id="crawl">
                  <i class="fas fa-spider"></i> 开始爬取
                </button>
              </div>
              <div
                v-show="crawl_result"
                id="result"
                class="mt-4 p-3 bg-light rounded"
                style="max-height: 200px; overflow-y: auto"
              ></div>
            </div>
          </section>
        </el-tab-pane>
        <!-- 复制小说 -->
        <el-tab-pane label="复制小说" name="second">
          <section id="input-section" class="card mb-5 shadow-lg">
            <div class="card-body">
              <h2 class="card-title text-center mb-4">小说文本输入</h2>
              <p class="text-muted text-center">输入小说文本，转换为漫画故事</p>
              <div class="mb-3">
                <label for="textInput" class="form-label">小说文本</label>
                <textarea
                  class="form-control"
                  id="textInput"
                  @input="handleTextInput"
                  rows="8"
                  placeholder="请输入小说文本，例如：&#10;第一章 森林冒险&#10;小猫咪咪在森林里迷路了..."
                  aria-label="小说文本"
                >
                </textarea>
              </div>
            </div>
          </section>
        </el-tab-pane>
        <el-tab-pane label="上传文件" name="third">
          <section id="input-section" class="card mb-5 shadow-lg">
            <div class="card-body">没做呢</div>
          </section>
          </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";

import config from "../../config.json";
const activeName = ref("first");
const crawl_result = ref(true);

const num = ref(1);


import bus from "../eventBus.js";
const isDone = ref(false);

function finishTask(input) {
  isDone.value = true;
  const info = {
    status: "done",
    time: new Date().toLocaleTimeString(),
    message: input,
  };

  console.log("A组件发送任务完成事件：", info);

  // 向全局广播消息
  bus.emit("task-finished", info);
}

//爬取小说

//let nextIndex = 0;
async function crawl() {
  const url = document.getElementById("novelUrl").value;
  const resultDiv = document.getElementById("result");
  const errorMessage = document.getElementById("error-message");

  if (!url) {
    ElMessage("请输入有效的URL");
    return;
  }
  try {
    let index = num.value - 1;
    const response = await fetch(`${config.BAKE_URL}/api/crawl`, {
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
      resultDiv.innerHTML += `<h3>HTTP错误：</h3><p>状态码 ${response.status}爬取章节内容失败: 未找到对应的章节链接，可能站点结构变化或URL无效</p><hr>`;
      throw new Error(`HTTP错误：状态码 ${response.status}`);
    }

    const data = await response.json();
    console.log("API响应：", data);

    if (data.error) {
      if (errorMessage) {
        errorMessage.textContent = `错误：${data.error}`;
        errorMessage.classList.remove("d-none");
      }
      return;
    }

    if (!data.chapters || !Array.isArray(data.chapters)) {
      if (errorMessage) {
        errorMessage.textContent = "错误：未收到有效章节数据";
        errorMessage.classList.remove("d-none");
      }
      return;
    }

    // 更新 num
    num.value = data.next_index;
    console.log("num", num.value);
    console.log("data.next_index", data.next_index);

    // 清空 resultDiv 并显示当前爬取的章节
    resultDiv.innerHTML = "";
    data.chapters.forEach((ch) => {
      resultDiv.innerHTML += `<h3>${ch.title}</h3><p>${ch.content}</p><hr>`;
    });

    // 添加“继续爬取”按钮，仅在有更多章节时显示
    if (num.value < data.total_chapters) {
      resultDiv.innerHTML += `<button id="continue-crawl" class="btn btn-primary">继续爬取</button>`;
      // 为“继续爬取”按钮绑定事件
      const continueBtn = document.getElementById("continue-crawl");
      if (continueBtn) {
        continueBtn.addEventListener("click", crawl);
      }
    } else {
      resultDiv.innerHTML += "<p>已爬取所有章节</p>";
    }

    // 将爬取内容自动填入文本框
    const textInput = document.getElementById("textInput");
    if (textInput) {
      textInput.value = data.chapters
        .map((ch) => `${ch.title}\n${ch.content}`)
        .join("\n");
      finishTask(textInput.value);
    }
  } catch (error) {
    console.error("爬取错误：", error);
    if (errorMessage) {
      errorMessage.textContent = `爬取失败：${error.message}`;
      errorMessage.classList.remove("d-none");
    }
  }
}
</script>
