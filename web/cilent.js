import config from "./config.js";
const API_KEY = config.API_KEY;
const BASE_URL = config.BASE_URL;
const TEXT_MODEL = config.TEXT_MODEL;
const IMAGE_MODEL = config.IMAGE_MODEL;

// 全局变量跟踪当前段落
let currentSegmentIndex = -1;
let segments = [];
let pendingData = null; // 存储待确认的 prompt 和 storyboard

// 绑定按钮事件
document
  .querySelector(".btn.btn-secondary.me-2")
  .addEventListener("click", segmentText);
document
  .querySelector(".btn.btn-info.me-2")
  .addEventListener("click", extractStoryboard);
document.querySelector(".btn.btn-primary").addEventListener("click", generateComic);

// 新增: 提示词标签逻辑
document.addEventListener("DOMContentLoaded", () => {
  const promptInput = document.getElementById("promptInput");
  const promptTags = document.getElementById("promptTags");

  if (!promptInput || !promptTags) {
    console.error("未找到 promptInput 或 promptTags 元素");
    return;
  }

  promptInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && promptInput.value.trim() !== "") {
      event.preventDefault();
      addTag(promptInput.value.trim());
      promptInput.value = "";
    }
  });

  function addTag(text) {
    const tag = document.createElement("span");
    tag.className = "badge bg-secondary me-2 mb-2";
    tag.innerHTML = `${text} <button type="button" class="btn-close btn-close-white ms-1" aria-label="Close"></button>`;
    tag
      .querySelector(".btn-close")
      .addEventListener("click", () => tag.remove());
    promptTags.appendChild(tag);
  }
});

// 分段并显示当前段落（本地分割，不发送请求）
async function segmentText() {
  const inputText = document.getElementById("textInput").value.trim();
  if (!inputText) return alert("请输入文本");

  const segmentsList = document.getElementById("segmentsList");
  segmentsList.innerHTML = "<p>分段中...</p>";

  try {
    // 按换行符分割
    const rawSegments = inputText.split(/\n+/).map((s) => s.trim()).filter(Boolean);
    segments = [];

    let currentSegment = "";
    for (let part of rawSegments) {
      if ((currentSegment + part).length >= 100) {
        segments.push(currentSegment + part);
        currentSegment = "";
      } else {
        currentSegment += part + " ";
      }
    }
    if (currentSegment.trim()) segments.push(currentSegment.trim());

    // 重置当前段落索引
    currentSegmentIndex = -1;

    // 显示第一段
    await showNextSegment();
  } catch (error) {
    segmentsList.innerHTML = `<p class="text-danger">分段失败: ${error.message}</p>`;
  }
}

// 显示下一段文本
async function showNextSegment() {
  const segmentsList = document.getElementById("segmentsList");
  if (currentSegmentIndex + 1 >= segments.length) {
    segmentsList.innerHTML = "<p>已显示所有段落</p>";
    return;
  }

  currentSegmentIndex++;
  const currentSegment = segments[currentSegmentIndex];

  segmentsList.innerHTML = `
    <div class="segment-item">
      <span class="segment-text" data-index="${currentSegmentIndex}">
        <b>第 ${currentSegmentIndex + 1} 段文本：</b> ${currentSegment}
      </span>
    </div>
  `;
}

// 提取分镜和内容
async function extractStoryboard() {
  const segmentsList = document.getElementById("segmentsList");
  if (currentSegmentIndex < 0 || currentSegmentIndex >= segments.length) {
    return alert("请先进行分段");
  }

  const currentSegment = segments[currentSegmentIndex];
  try {
    const response = await generateStoryboard(currentSegment);
    pendingData = response.data; // 存储 prompt 和 storyboard
    segmentsList.innerHTML = `
      <div class="segment-item">
        <span class="segment-text" data-index="${currentSegmentIndex}">
          <b>第 ${currentSegmentIndex + 1} 段：</b><br>
          <b>提示词：</b> ${response.data.prompt}<br>
          <b>分镜描述：</b> ${response.data.storyboard}
        </span>
        <button class="btn btn-success btn-sm mt-2 confirm-storyboard" data-index="${currentSegmentIndex}">确认并生成图像</button>
      </div>
    `;

    // 绑定确认按钮事件
    document.querySelectorAll(".confirm-storyboard").forEach((btn) => {
      btn.addEventListener("click", async () => {
        await generateComicFromConfirmed();
      });
    });
  } catch (error) {
    segmentsList.innerHTML = `<p class="text-danger">分镜提取失败: ${error.message}</p>`;
  }
}

// 从确认的故事板生成漫画
async function generateComicFromConfirmed() {
  const panelsDiv = document.getElementById("panels");
  if (!pendingData) {
    panelsDiv.innerHTML = "<p class='text-danger'>没有待确认的故事板</p>";
    return;
  }

  panelsDiv.innerHTML = "<p>生成中...</p>";

  try {
    const imageData = await generateImage(pendingData.prompt, pendingData.storyboard);
    panelsDiv.innerHTML = `
      <div class="comic-panel">
        <h6>面板 ${currentSegmentIndex + 1}</h6>
        <img src="data:image/png;base64,${imageData.image_base64}" alt="面板${currentSegmentIndex + 1}">
        <p>${pendingData.storyboard}</p>
      </div>
    `;
    // 显示下一段文本
    await showNextSegment();
  } catch (error) {
    panelsDiv.innerHTML = `<p class="text-danger">漫画生成失败: ${error.message}</p>`;
  }
}

// 生成漫画
async function generateComic() {
  if (currentSegmentIndex < 0 || currentSegmentIndex >= segments.length) {
    return alert("请先进行分段");
  }

  if (!pendingData) {
    return alert("请先提取分镜和内容");
  }

  await generateComicFromConfirmed();
}

// API 调用：生成分镜和内容
async function generateStoryboard(text) {
  const promptTags = document.getElementById("promptTags");
  if (!promptTags) {
    console.error("未找到 promptTags 元素");
    return { data: { prompt: text, storyboard: text } }; // 回退逻辑
  }

  const prompts = Array.from(promptTags.querySelectorAll(".badge")).map((tag) =>
    tag.textContent.trim().replace(/×$/, "").trim()
  );
  const extraPrompts = prompts.length > 0 ? `, 附加提示词: ${prompts.join(", ")}` : "";
  console.log("发送 /api/text 请求体:", JSON.stringify({
    model: TEXT_MODEL,
    messages: [
      {
        role: "system",
        content: "你是一个漫画创作者。将输入文本转换为1个图片的故事板，每行描述一个分镜，包括场景、角色和简单对话。用中文回复，将所有分镜合并为一个故事板，并生成对应的提示词，最后输出对应的英文。",
      },
      { role: "user", content: text + extraPrompts },
    ],
    max_tokens: 300,
    temperature: 0.7,
  }));

  const response = await fetch("http://localhost:3000/api/text", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      model: TEXT_MODEL,
      messages: [
        {
          role: "system",
          content: "你是一个漫画创作者。将输入文本转换为1个图片的故事板，每行描述一个分镜，包括场景、角色和简单对话。用中文回复，将所有分镜合并为一个故事板，并生成对应的提示词，最后输出对应的英文。",
        },
        { role: "user", content: text + extraPrompts },
      ],
      max_tokens: 300,
      temperature: 0.7,
    }),
  });
  if (!response.ok) {
    const errorData = await response.json();
    console.error("错误响应 /api/text:", errorData);
    throw new Error(`分镜生成失败: ${errorData.error || response.statusText}`);
  }
  const data = await response.json();
  console.log("收到 /api/text 响应:", data);
  return data; // 期望 { message, data: { prompt, storyboard } }
}

// API 调用：生成图像
async function generateImage(prompt, storyboard) {
  console.log("发送 /generate_image 请求");
  const response = await fetch("http://localhost:3000/generate_image", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({}),
  });
  if (!response.ok) {
    const errorData = await response.json();
    console.error("错误响应 /generate_image:", errorData);
    throw new Error(`图像生成失败: ${errorData.error || response.statusText}`);
  }
  const data = await response.json();
  console.log("收到 /generate_image 响应:", data);
  return data; // 返回 { image_base64: "..." }
}