import config from "./config.js";
const API_KEY = config.API_KEY;
const BASE_URL = config.BASE_URL;
const TEXT_MODEL = config.TEXT_MODEL;
const IMAGE_MODEL = config.IMAGE_MODEL;

// 全局变量跟踪当前段落
let currentSegmentIndex = -1;
let segments = [];
let pendingData = null; // 存储待确认的 prompt 和 storyboard
window.addEventListener("error", (e) =>
  console.error("GLOBAL ERROR:", e.error || e.message, e)
);
window.addEventListener("unhandledrejection", (e) =>
  console.error("GLOBAL UNHANDLED REJECTION:", e.reason)
);

// 把 segment 按钮也加上 preventDefault（以防放在 form 中）
const segBtn = document.querySelector(".btn.btn-secondary.me-2");
if (segBtn) {
  segBtn.addEventListener("click", (event) => {
    event.preventDefault();
    segmentText();
  });
}

// 事件代理：处理动态生成的确认按钮，避免遗漏 binding
const segmentsList = document.getElementById("segmentsList");
if (segmentsList) {
  segmentsList.addEventListener("click", async (e) => {
    if (e.target && e.target.matches(".confirm-storyboard")) {
      e.preventDefault();
      await generateComicFromConfirmed();
    }
  });
}
window.addEventListener("error", (e) => {
  console.error("捕获到 error:", e.error || e.message, e);
});
window.addEventListener("unhandledrejection", (e) => {
  console.error("捕获到 unhandledrejection:", e.reason);
});

// 绑定按钮事件
document
  .querySelector(".btn.btn-secondary.me-2")
  .addEventListener("click", (event) => {
    event.preventDefault();
    segmentText();
  });

document
  .querySelector(".btn.btn-info.me-2")
  .addEventListener("click", (event) => {
    event.preventDefault(); // 阻止默认行为
    extractStoryboard();
  });

document
  .querySelector(".btn.btn-primary")
  .addEventListener("click", (event) => {
    event.preventDefault(); // 阻止默认行为
    generateComic();
  });

document.querySelectorAll(".confirm-storyboard").forEach((btn) => {
  btn.addEventListener("click", async (event) => {
    event.preventDefault(); // 阻止默认行为
    await generateComicFromConfirmed();
  });
});

// 新增: 提示词标签逻辑
// document.addEventListener("DOMContentLoaded", () => {
//   const promptInput = document.getElementById("promptInput");
//   const promptTags = document.getElementById("promptTags");

//   if (!promptInput || !promptTags) {
//     console.error("未找到 promptInput 或 promptTags 元素");
//     return;
//   }

//   promptInput.addEventListener("keydown", (event) => {
//     if (event.key === "Enter" && promptInput.value.trim() !== "") {
//       event.preventDefault();
//       addTag(promptInput.value.trim());
//       promptInput.value = "";
//     }
//   });

//   function addTag(text) {
//     const tag = document.createElement("span");
//     tag.className = "badge bg-secondary me-2 mb-2";
//     tag.innerHTML = `${text} <button type="button" class="btn-close btn-close-white ms-1" aria-label="Close"></button>`;
//     tag
//       .querySelector(".btn-close")
//       .addEventListener("click", () => tag.remove());
//     promptTags.appendChild(tag);
//   }
// });

// 分段并显示当前段落（本地分割，不发送请求）
async function segmentText() {
  const inputText = document.getElementById("textInput").value.trim();
  if (!inputText) return alert("请输入文本");

  const segmentsList = document.getElementById("segmentsList");
  segmentsList.innerHTML = "<p>分段中...</p>";

  try {
    // 按换行符分割
    const rawSegments = inputText
      .split(/\n+/)
      .map((s) => s.trim())
      .filter(Boolean);
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
        segmentsList.innerHTML = "<p class='text-danger'>请先进行分段</p>";
        return;
    }

    const currentSegment = segments[currentSegmentIndex];
    segmentsList.innerHTML = "<p>提取分镜中...</p>";
    try {
        const response = await generateStoryboard(currentSegment);
        pendingData = response.data; // 存储解析后的 scene, prompt, character
        segmentsList.innerHTML = `
            <div class="segment-item">
                <span class="segment-text" data-index="${currentSegmentIndex}">
                    <b>第 ${currentSegmentIndex + 1} 段：</b><br>
                    <b>场景：</b> ${pendingData.scene}<br>
                    <b>提示词描述：</b> ${pendingData.prompt}<br>
                    <b>人物描述：</b> ${pendingData.character || '无'}
                </span>
                <button type="button" class="btn btn-success btn-sm mt-2 confirm-storyboard" data-index="${currentSegmentIndex}">确认并生成图像</button>
            </div>
        `;
        // 动态绑定事件
        segmentsList.querySelectorAll(".confirm-storyboard").forEach((btn) => {
            btn.addEventListener("click", async (event) => {
                event.preventDefault();
                await generateComicFromConfirmed();
            });
        });
    } catch (error) {
        console.error("提取分镜失败:", error);
        segmentsList.innerHTML = `<p class="text-danger">分镜提取失败: ${error.message}</p>`;
    }
}
///////////////////////////////////////////////////////////////////base64
// async function generateComicFromConfirmed() {
//     const panelsDiv = document.getElementById("panels");
//     if (!pendingData) {
//         panelsDiv.innerHTML = "<p class='text-danger'>没有待确认的故事板</p>";
//         return;
//     }

//     panelsDiv.innerHTML = "<p>生成中...</p>";
//     try {
//         const imageData = await generateImage(pendingData);
//         panelsDiv.innerHTML = `
//             <div class="comic-panel">
//                 <h6>面板 ${currentSegmentIndex + 1}</h6>
//                 <img src="data:image/png;base64,${imageData.image_base64}" alt="面板${currentSegmentIndex + 1}">
//                 <p>${pendingData.scene}</p>
//                 ${pendingData.character ? `<p><b>人物：</b> ${pendingData.character}</p>` : ''}
//             </div>
//         `;
//         await showNextSegment();
//     } catch (error) {
//         console.error("漫画生成失败:", error);
//         panelsDiv.innerHTML = `<p class="text-danger">漫画生成失败: ${error.message}</p>`;
//     }
// }

// 从确认的故事板生成漫画
// async function generateComicFromConfirmed() {
//   const panelsDiv = document.getElementById("panels");
//   if (!pendingData) {
//     panelsDiv.innerHTML = "<p class='text-danger'>没有待确认的故事板</p>";
//     return;
//   }

//   panelsDiv.innerHTML = "<p>生成中...</p>";

//   try {
//     const imageData = await generateImage(
//       pendingData.prompt,
//       pendingData.storyboard
//     );
//     panelsDiv.innerHTML = `
//       <div class="comic-panel">
//         <h6>面板 ${currentSegmentIndex + 1}</h6>
//         <img src="data:image/png;base64,${imageData.image_base64}" alt="面板${
//       currentSegmentIndex + 1
//     }">
//         <p>${pendingData.storyboard}</p>
//       </div>
//     `;
//     // 显示下一段文本
//     await showNextSegment();
//   } catch (error) {
//     panelsDiv.innerHTML = `<p class="text-danger">漫画生成失败: ${error.message}</p>`;
//   }
// }

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
                    "你是一个漫画创作者。将输入文本转换为1个图片的漫画，要求含有场景，提示词和人物(若有的话依照人物名称生成符合该人物的服饰，长相，没有就返回null)，输出格式为纯文本，字段按以下格式：scene:...,prompt:...,character:...，字段间用逗号分隔，不包含换行符，用英文回答我。",
            },
            { role: "user", content: text + extraPrompts },
        ],
        max_tokens: 300,
        temperature: 0.7,
    };
    console.log("发送 /api/text 请求体:", JSON.stringify(requestBody, null, 2));

    try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 10000); // 10秒超时
        const response = await fetch("http://localhost:3000/api/text", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(requestBody),
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
        const parsedData = { scene: "默认场景", prompt: "默认提示词", character: "" };

        // 使用正则表达式提取 scene, prompt, character
const sceneMatch = textContent.match(/scene:([^.]*\.)|$/);
        const promptMatch = textContent.match(/prompt:([^.]*\.)|$/);
        const characterMatch = textContent.match(/character:([^.]*\.)|$/);

        if (sceneMatch && sceneMatch[1]) {
            parsedData.scene = sceneMatch[1].trim();
        }
        if (promptMatch && promptMatch[1]) {
            parsedData.prompt = promptMatch[1].trim() + "Output in a style similar to that of Japanese anime";
        }
        if (characterMatch && characterMatch[1] && characterMatch[1] !== "null") {
            parsedData.character = characterMatch[1].trim();
        }

        return { data: parsedData };
    } catch (error) {
        console.error("generateStoryboard 错误:", error);
        throw error;
    }
}

// API 调用：生成图像
// API 调用：生成图像
async function generateImage(data) {
    console.log("发送 /api/image 请求");
    const requestBody = {
        model: "ep-20251021153509-xh86n",
        role: "user",
        prompt: data.prompt,
        storyboard: data.scene
    };
    const response = await fetch("http://localhost:3000/api/image", {
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

    // 检查响应是否包含有效的图像 URL
    if (!dataResponse.data || !dataResponse.data[0] || !dataResponse.data[0].url) {
        throw new Error("图像生成失败：响应中缺少有效的图像 URL");
    }

    return { image_url: dataResponse.data[0].url }; // 返回图像 URL
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
        const imageData = await generateImage(pendingData);
        panelsDiv.innerHTML = `
            <div class="comic-panel">
                <h6>面板 ${currentSegmentIndex + 1}</h6>
                <img src="${imageData.image_url}" alt="面板${currentSegmentIndex + 1}">
                <p>${pendingData.scene}</p>
                ${pendingData.character ? `<p><b>人物：</b> ${pendingData.character}</p>` : ''}
            </div>
        `;
        await showNextSegment();
    } catch (error) {
        console.error("漫画生成失败:", error);
        panelsDiv.innerHTML = `<p class="text-danger">漫画生成失败: ${error.message}</p>`;
    }
}