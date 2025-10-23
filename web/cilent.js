import config from "./config.js";
const API_KEY = config.API_KEY;
const BASE_URL = config.BASE_URL;
const TEXT_MODEL = config.TEXT_MODEL;
const IMAGE_MODEL = config.IMAGE_MODEL;
document
  .querySelector(".btn.btn-secondary.me-2")
  .addEventListener("click", segmentText);
document
  .querySelector(".btn.btn-info.me-2")
  .addEventListener("click", extractKeywords);
document.querySelector(".btn-primary").addEventListener("click", generateComic);
// 新增: 处理提示词标签逻辑
document.addEventListener("DOMContentLoaded", () => {
  const promptInput = document.getElementById("promptInput");
  const promptTags = document.getElementById("promptTags");

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

// 分段并显示
async function segmentText() {
    const inputText = document.getElementById('textInput').value.trim();
    if (!inputText) return alert('请输入文本');

    const segmentsList = document.getElementById('segmentsList');
    segmentsList.innerHTML = '<p>分段中...</p>';

    try {
        // 按换行符分割
        const rawSegments = inputText.split(/\n+/).map(s => s.trim()).filter(Boolean);
        const segments = [];

        let currentSegment = '';
        for (let part of rawSegments) {
            if ((currentSegment + part).length >= 100) {
                segments.push(currentSegment + part);
                currentSegment = '';
            } else {
                currentSegment += part + ' ';
            }
        }
        if (currentSegment.trim()) segments.push(currentSegment.trim());

        // 渲染结果
        segmentsList.innerHTML = segments.map((seg, index) => `
            <div class="segment-item">
                <input type="checkbox" class="segment-checkbox" data-index="${index}" checked>
                <span class="segment-text" data-index="${index}">
                    <b>第 ${index + 1} 段：</b> ${seg}
                </span>
            </div>
        `).join('');

    } catch (error) {
        segmentsList.innerHTML = `<p class="text-danger">分段失败: ${error.message}</p>`;
    }
}


// 提取关键词并高亮
async function extractKeywords() {
  const segmentsList = document.getElementById("segmentsList");
  const checkboxes = segmentsList.querySelectorAll(".segment-checkbox:checked");
  if (checkboxes.length === 0) return alert("请至少选择一个段落");

  const selectedIndices = Array.from(checkboxes).map((cb) =>
    parseInt(cb.dataset.index)
  );
  const inputText = Array.from(segmentsList.querySelectorAll(".segment-text"))
    .filter((_, i) => selectedIndices.includes(i))
    .map((span) => span.textContent)
    .join("\n\n");

  try {
    const segments = await segmentAndExtractKeywords(inputText);
    segments.forEach((seg, i) => {
      const span = segmentsList.querySelector(
        `.segment-text[data-index="${selectedIndices[i]}"]`
      );
      let highlightedText = seg.text;
      seg.keywords.forEach((kw) => {
        const regex = new RegExp(
          `(${kw.keyword.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, "\\$&")})`,
          "g"
        );
        highlightedText = highlightedText.replace(
          regex,
          `<span class="highlight">$1</span>`
        );
      });
      span.innerHTML = highlightedText;
    });
  } catch (error) {
    alert(`关键词提取失败: ${error.message}`);
  }
}

// 生成漫画
async function generateComic() {
  const segmentsList = document.getElementById("segmentsList");
  const checkboxes = segmentsList.querySelectorAll(".segment-checkbox:checked");
  if (checkboxes.length === 0) return alert("请至少选择一个段落");

  const selectedIndices = Array.from(checkboxes).map((cb) =>
    parseInt(cb.dataset.index)
  );
  const segments = Array.from(segmentsList.querySelectorAll(".segment-text"))
    .filter((_, i) => selectedIndices.includes(i))
    .map((span) => span.textContent);

  const promptTags = document.getElementById("promptTags");
  const prompts = Array.from(promptTags.querySelectorAll(".badge")).map((tag) =>
    tag.textContent.trim().replace(/×$/, "").trim()
  );

  const panelsDiv = document.getElementById("panels");
  panelsDiv.innerHTML = "<p>生成中...</p>";

  for (let i = 0; i < Math.min(segments.length, 4); i++) {
    const segment = segments[i];
    const prompt = `${segment}。${
      prompts.length > 0 ? `附加提示词: ${prompts.join(", ")}` : ""
    }`;
    const storyboard = await generateStoryboard(prompt);
    const panelDesc =
      storyboard.split("\n").filter((p) => p.trim())[i % 4] || segment;
    const imageUrl = await generateImage(
      panelDesc,
      `漫画风格，面板${i + 1}`,
      prompts
    );
    panelsDiv.innerHTML += `
            <div class="comic-panel">
                <h6>面板 ${i + 1}</h6>
                <img src="${imageUrl}" alt="面板${i + 1}">
                <p>${panelDesc}</p>
            </div>`;
  }
}

// API 调用：分段和关键词提取
async function segmentAndExtractKeywords(text) {
  const response = await fetch("http://localhost:3000/api/segment_keywords", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text }),
  });
  if (!response.ok) throw new Error("分段和关键词提取失败");
  const data = await response.json();
  console.log("response:", data);
  return data;
}

// API 调用：生成故事板
async function generateStoryboard(text) {
  const response = await fetch("http://localhost:3000/api/text", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      model: TEXT_MODEL,
      messages: [
        {
          role: "system",
          content:
            "你是一个漫画创作者。将输入文本转换为1个图片的故事板,每行描述一个分镜,包括场景、角色和简单对话。用中文回复,将所有分镜合并为一个故事板。",
        },
        { role: "user", content: text },
      ],
      max_tokens: 300,
      temperature: 0.7,
    }),
  });
  const data = await response.json();
  return data.choices[0].message.content;
}

// API 调用：生成图像
async function generateImage(prompt, style, prompts = []) {
  const extraPrompts = prompts.length > 0 ? `, ${prompts.join(", ")}` : "";
  const fullPrompt = `${prompt}。${style}，黑白线稿漫画风格，详细插图${extraPrompts}`;
  const response = await fetch("http://localhost:3000/api/image", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      model: IMAGE_MODEL,
      prompt: fullPrompt,
      n: 1,
      size: "512x512",
      response_format: "url",
    }),
  });
  const data = await response.json();
  return data.data[0].url;
}
