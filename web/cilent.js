// import config from "./config.js";

// const API_KEY = config.API_KEY;
// const BASE_URL = config.BASE_URL;
// const TEXT_MODEL = config.TEXT_MODEL;
// const IMAGE_MODEL = config.IMAGE_MODEL;
// const BAKE_URL = config.BAKE_URL;

// // 全局变量
// let currentSegmentIndex = -1;
// let segments = [];
// let pendingData = null;
// let last_image = null;
// // 全局错误处理
// // window.addEventListener("error", (e) =>
// //   console.error("全局错误:", e.error || e.message, e)
// // );
// // window.addEventListener("unhandledrejection", (e) =>
// //   console.error("未处理 Promise 拒绝:", e.reason)
// // );

// // 等待 DOM 加载完成
// // document.addEventListener("DOMContentLoaded", () => {
// //   const segBtn = document.querySelector(".btn.btn-outline-secondary");
// //   const storyboardBtn = document.querySelector(".btn.btn-outline-info");
// //   const generateBtn = document.querySelector(".btn.btn-primary");
// //   const crawlBtn = document.getElementById("crawl");

// //   // 绑定“分段并显示”按钮
// //   if (segBtn) {
// //     segBtn.addEventListener("click", (event) => {
// //       event.preventDefault();
// //       segmentText();
// //     });
// //   } else {
// //     console.error("未找到 '分段并显示' 按钮");
// //   }

// //   // 绑定“提取分镜和内容”按钮
// //   if (storyboardBtn) {
// //     storyboardBtn.addEventListener("click", (event) => {
// //       event.preventDefault();
// //       extractStoryboard();
// //     });
// //   } else {
// //     console.error("未找到 '提取分镜和内容' 按钮");
// //   }

// //   // 绑定“生成漫画”按钮
// //   if (generateBtn) {
// //     generateBtn.addEventListener("click", (event) => {
// //       event.preventDefault();
// //       generateComic();
// //     });
// //   } else {
// //     console.error("未找到 '生成漫画' 按钮");
// //   }

// //   // 绑定“开始爬取”按钮
// //   if (crawlBtn) {
// //     crawlBtn.addEventListener("click", crawl);
// //   } else {
// //     console.error("未找到 '开始爬取' 按钮");
// //   }

// //   // 事件代理：处理动态生成的确认按钮
// //   const segmentsList = document.getElementById("segmentsList");
// //   if (segmentsList) {
// //     segmentsList.addEventListener("click", async (e) => {
// //       if (e.target.matches(".confirm-storyboard")) {
// //         e.preventDefault();
// //         await generateComicFromConfirmed();
// //       }
// //     });
// //   } else {
// //     console.error("未找到 segmentsList 元素");
// //   }
// // });

// // 分段并显示当前段落（本地分割，不发送请求）
// async function segmentText() {
//   const inputText = document.getElementById("textInput").value.trim();
//   const errorMessage = document.getElementById("error-message");
//   const segmentsList = document.getElementById("segmentsList");

//   if (!inputText) {
//     if (errorMessage) {
//       errorMessage.textContent = "请输入小说文本";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (errorMessage) errorMessage.classList.add("d-none");
//   segmentsList.innerHTML = "<p>分段中...</p>";

//   try {
//     // 预处理：移除多余换行符，规范化标点
//     let cleanedText = inputText
//       .replace(/\n+/g, " ") // 替换换行符为空格
//       .replace(/\s+/g, " ") // 合并多个空格
//       .trim();

//     // 按中英文标点符号分割，支持中文标点（。！？）和英文标点（.!?）
//     const rawSegments = cleanedText
//       .split(/(?<=[.!?。！？])[\s]+|(?<=[\'"])[\s]+/)
//       .map((s) => s.trim())
//       .filter(Boolean);

//     segments = [];
//     let currentSegment = "";
//     const minSegmentLength = 50; // 最小段落长度
//     const maxSegmentLength = 200; // 最大段落长度，防止段落过长

//     for (let part of rawSegments) {
//       if ((currentSegment + part).length >= minSegmentLength) {
//         segments.push((currentSegment + part).substring(0, maxSegmentLength));
//         currentSegment = "";
//       } else {
//         currentSegment += part + " ";
//       }
//     }
//     if (currentSegment.trim()) {
//       segments.push(currentSegment.trim());
//     }

//     // 如果没有分出段落，强制将整个文本作为一段
//     if (segments.length === 0 && cleanedText) {
//       segments.push(cleanedText.substring(0, maxSegmentLength));
//     }

//     // 重置当前段落索引
//     currentSegmentIndex = -1;

//     // 显示第一段
//     await showNextSegment();
//   } catch (error) {
//     console.error("分段失败:", error);
//     segmentsList.innerHTML = `<p class="text-danger">分段失败: ${error.message}</p>`;
//   }
// }
// // 显示下一段文本
// async function showNextSegment() {
//   const segmentsList = document.getElementById("segmentsList");
//   if (currentSegmentIndex + 1 >= segments.length) {
//     segmentsList.innerHTML = "<p>已显示所有段落</p>";
//     return;
//   }

//   currentSegmentIndex++;
//   const currentSegment = segments[currentSegmentIndex];

//   segmentsList.innerHTML = `
//     <div class="segment-item">
//       <span class="segment-text" data-index="${currentSegmentIndex}">
//         <b>第 ${currentSegmentIndex + 1} 段文本：</b> ${currentSegment}
//       </span>
//     </div>
//   `;
// }

// // 提取分镜和内容
// async function extractStoryboard() {
//   const segmentsList = document.getElementById("segmentsList");
//   const errorMessage = document.getElementById("error-message");

//   if (currentSegmentIndex < 0 || currentSegmentIndex >= segments.length) {
//     if (errorMessage) {
//       errorMessage.textContent = "请先进行分段";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (errorMessage) errorMessage.classList.add("d-none");
//   const currentSegment = segments[currentSegmentIndex];
//   segmentsList.innerHTML = "<p>提取分镜中...</p>";

//   try {
//     const response = await generateStoryboard(currentSegment);
//     pendingData = response.data; // 存储解析后的 scene, prompt, character
//     segmentsList.innerHTML = `
//       <div class="segment-item">
//         <span class="segment-text" data-index="${currentSegmentIndex}">
//           <b>第 ${currentSegmentIndex + 1} 段：</b><br>
//           <b>场景：</b> ${pendingData.scene}<br>
//           <b>提示词描述：</b> ${pendingData.prompt}<br>
//           <b>人物描述：</b> ${pendingData.character || "无"}
//         </span>
//         <button type="button" class="btn btn-success btn-sm mt-2 confirm-storyboard" data-index="${currentSegmentIndex}">确认并生成图像</button>
//       </div>
//     `;
//   } catch (error) {
//     console.error("提取分镜失败:", error);
//     segmentsList.innerHTML = `<p class="text-danger">分镜提取失败: ${error.message}</p>`;
//   }
// }

// // 从确认的故事板生成漫画
// async function generateComicFromConfirmed() {
//   const panelsDiv = document.getElementById("panels");
//   const errorMessage = document.getElementById("error-message");

//   if (!pendingData) {
//     if (errorMessage) {
//       errorMessage.textContent = "没有待确认的故事板";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (errorMessage) errorMessage.classList.add("d-none");
//   panelsDiv.innerHTML = "<p>生成中...</p>";

//   try {
//     const imageData = await generateImage(pendingData);
//     panelsDiv.innerHTML = `
//       <div class="comic-panel">
//         <h6>面板 ${currentSegmentIndex + 1}</h6>
//         <img src="${imageData.image_url}" alt="面板${currentSegmentIndex + 1}">
//         <p>${pendingData.scene}</p>
//         ${
//           pendingData.character
//             ? `<p><b>人物：</b> ${pendingData.character}</p>`
//             : ""
//         }
//       </div>
//     `;
//     // 只有在成功生成漫画后才显示下一段
//     await showNextSegment();
//   } catch (error) {
//     console.error("漫画生成失败:", error);
//     panelsDiv.innerHTML = `<p class="text-danger">漫画生成失败: ${error.message}</p>`;
//   }
// }

// // 生成漫画
// async function generateComic() {
//   const errorMessage = document.getElementById("error-message");

//   if (currentSegmentIndex < 0 || currentSegmentIndex >= segments.length) {
//     if (errorMessage) {
//       errorMessage.textContent = "请先进行分段";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (!pendingData) {
//     if (errorMessage) {
//       errorMessage.textContent = "请先提取分镜和内容";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (errorMessage) errorMessage.classList.add("d-none");
//   await generateComicFromConfirmed();
// }

// // API 调用：生成分镜和内容
// async function generateStoryboard(text) {
//   const promptTags = document.getElementById("promptTags");
//   const prompts = promptTags
//     ? Array.from(promptTags.querySelectorAll(".badge")).map((tag) =>
//         tag.textContent.trim().replace(/×$/, "").trim()
//       )
//     : [];
//   const extraPrompts =
//     prompts.length > 0 ? `, 附加提示词: ${prompts.join(", ")}` : "";
//   const requestBody = {
//     model: TEXT_MODEL,
//     messages: [
//       {
//         role: "system",
//         content:
//           "你是一个漫画创作者。将输入文本转换为1个图片的漫画，要求含有场景，提示词和人物(若有的话依照人物名称生成符合该人物的服饰，长相，没有就返回null)，输出格式为纯文本，字段按以下格式：scene:...;prompt:...;character:...;字段间用分号分隔，不包含换行符，用中文回答我。",
//       },
//       { role: "user", content: text + extraPrompts },
//     ],
//     max_tokens: 300,
//     temperature: 0.7,
//   };
//   console.log("发送 /api/text 请求体:", JSON.stringify(requestBody, null, 2));

//   try {
//     const controller = new AbortController();
//     const timeoutId = setTimeout(() => controller.abort(), 30000); // 30秒超时
//     const response = await fetch("http://localhost:3000/api/text", {
//       method: "POST",
//       headers: { "Content-Type": "application/json" },
//       body: JSON.stringify(requestBody),
//       signal: controller.signal,
//     });
//     clearTimeout(timeoutId);
//     if (!response.ok) {
//       const errorData = await response.json().catch(() => ({}));
//       console.error("错误响应 /api/text:", errorData);
//       throw new Error(
//         `分镜生成失败: ${errorData.error || response.statusText}`
//       );
//     }
//     const data = await response.json();
//     console.log("收到 /api/text 响应:", JSON.stringify(data, null, 2));

//     if (!data.data || !data.data.message) {
//       throw new Error("后端响应格式错误，缺少 message 字段");
//     }

//     // 解析纯文本
//     const textContent = data.data.message;
//     const parsedData = {
//       scene: "默认场景",
//       prompt: "默认提示词",
//       character: "",
//     };

//     const sceneMatch = textContent.match(/scene:([\s\S]*?)(?=;prompt:|$)/);
//     const promptMatch = textContent.match(/prompt:([\s\S]*?)(?=;character:|$)/);
//     const characterMatch = textContent.match(/character:([\s\S]*?)(?=;|$)/);

//     console.log("解析结果:", {
//       scene: sceneMatch ? sceneMatch[1] : null,
//       prompt: promptMatch ? promptMatch[1] : null,
//       character: characterMatch ? characterMatch[1] : null,
//     });
//     if (sceneMatch && sceneMatch[1]) {
//       parsedData.scene = sceneMatch[1].trim();
//     }
//     if (promptMatch && promptMatch[1]) {
//       parsedData.prompt =
//         promptMatch[1].trim() +
//         "Output in a style similar to that of Japanese anime";
//     }
//     if (characterMatch && characterMatch[1] && characterMatch[1] !== "null") {
//       parsedData.character = characterMatch[1].trim();
//     }
//     if (last_image !== null) {
//       parsedData.init_image = last_image;
//     }
//     return { data: parsedData };
//   } catch (error) {
//     console.error("generateStoryboard 错误:", error);
//     throw error;
//   }
// }

// // API 调用：生成图像
// async function generateImage(data) {
//   console.log("发送 /api/image 请求");
//   console.log("图像生成输入数据:", data);
//   const requestBody = {
//     model: "ep-20251021153509-xh86n",
//     role: "user",
//     prompt: data.prompt,
//     storyboard: data.scene,
//   };
//   const response = await fetch("http://localhost:3000/api/image", {
//     method: "POST",
//     headers: { "Content-Type": "application/json" },
//     body: JSON.stringify(requestBody),
//   });
//   if (!response.ok) {
//     const errorData = await response.json();
//     console.error("错误响应 /api/image:", errorData);
//     throw new Error(`图像生成失败: ${errorData.error || response.statusText}`);
//   }
//   const dataResponse = await response.json();
//   console.log("收到 /api/image 响应:", dataResponse);

//   if (
//     !dataResponse.data ||
//     !dataResponse.data[0] ||
//     !dataResponse.data[0].url
//   ) {
//     throw new Error("图像生成失败：响应中缺少有效的图像 URL");
//   }
//   last_image = dataResponse.data[0].url;
//   console.log("生成的图像 URL:", last_image);
//   return { image_url: dataResponse.data[0].url };
// }

// 爬取小说章节
// let nextIndex = 0;
// async function crawl() {
//   const url = document.getElementById("novelUrl").value;
//   const resultDiv = document.getElementById("result");
//   const errorMessage = document.getElementById("error-message");

//   if (!url) {
//     if (errorMessage) {
//       errorMessage.textContent = "请输入有效的URL";
//       errorMessage.classList.remove("d-none");
//     }
//     return;
//   }

//   if (errorMessage) errorMessage.classList.add("d-none");

//   try {
//     const response = await fetch("http://localhost:3000/api/crawl", {
//       method: "POST",
//       headers: { "Content-Type": "application/json" },
//       body: JSON.stringify({
//         novel_url: url,
//         start_index: nextIndex,
//         limit: 5,
//       }),
//     });

//     if (!response.ok) {
//       throw new Error(`HTTP错误：状态码 ${response.status}`);
//     }

//     const data = await response.json();
//     console.log("API响应：", data);

//     if (data.error) {
//       if (errorMessage) {
//         errorMessage.textContent = `错误：${data.error}`;
//         errorMessage.classList.remove("d-none");
//       }
//       return;
//     }

//     if (!data.chapters || !Array.isArray(data.chapters)) {
//       if (errorMessage) {
//         errorMessage.textContent = "错误：未收到有效章节数据";
//         errorMessage.classList.remove("d-none");
//       }
//       return;
//     }

//     // 更新 nextIndex
//     nextIndex = data.next_index;

//     // 清空 resultDiv 并显示当前爬取的章节
//     resultDiv.innerHTML = "";
//     data.chapters.forEach((ch) => {
//       resultDiv.innerHTML += `<h3>${ch.title}</h3><p>${ch.content}</p><hr>`;
//     });

//     // 添加“继续爬取”按钮，仅在有更多章节时显示
//     if (nextIndex < data.total_chapters) {
//       resultDiv.innerHTML += `<button id="continue-crawl" class="btn btn-primary">继续爬取</button>`;
//       // 为“继续爬取”按钮绑定事件
//       const continueBtn = document.getElementById("continue-crawl");
//       if (continueBtn) {
//         continueBtn.addEventListener("click", crawl);
//       }
//     } else {
//       resultDiv.innerHTML += "<p>已爬取所有章节</p>";
//     }

//     // 将爬取内容自动填入文本框
//     const textInput = document.getElementById("textInput");
//     textInput.value = data.chapters
//       .map((ch) => `${ch.title}\n${ch.content}`)
//       .join("\n");
//   } catch (error) {
//     console.error("爬取错误：", error);
//     if (errorMessage) {
//       errorMessage.textContent = `爬取失败：${error.message}`;
//       errorMessage.classList.remove("d-none");
//     }
//   }
// }

///激活render
import { ElMessage, ElMessageBox } from "element-plus";
async function ping() {
  try {
    await fetch("https://ai-painting.onrender.com/");
    console.log("Ping success");
    ElMessage({
      message: "服务器冷启动成功",
      type: "success",
    });
  } catch (err) {
    console.log("Ping failed:", err);
    ElMessage({
      message: "服务器冷启动失败，请联系管理员",
      type: "error",
    });
  }
}

// 每 10 分钟 ping 一次
setInterval(ping, 10 * 60 * 1000);

// 页面打开后立即 ping 一次
ping();

///背景文字
import config from "./config.js";
let cycle = [true, true, true, true, true, true];
show_all_bgdtxt();

async function show_all_bgdtxt() {
  for (let groupIndex = 1; groupIndex < 7; groupIndex++) {
    // console.log("i=", groupIndex);
    const bgdtxt = document.getElementById("bgdtxt_" + groupIndex);
    bgdtxt.setAttribute("display", "true");
    show_one_bgdtxt(bgdtxt, groupIndex);
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
}
//更新位置
function updatabgdtxt(bgdtxt) {
  try {
    bgdtxt.style.fontSize = getRandomInt(10, 22) + "px";
    let rand_txt = getRandomInt(0, config.bgdtxt_world.length - 1);
    bgdtxt.textContent = config.bgdtxt_world[rand_txt];
    bgdtxt.style.top = getRandomInt(3, 99) + "%";
    bgdtxt.style.left = getRandomInt(3, 99) + "%";
    bgdtxt.style.transform = `rotate(${getRandomInt(-90, 90)}deg)`;
  } catch (error) {
    console.error("Error updating background text:", error);
  }
}
// 获取 min（包含）到 max（包含）的随机整数
function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min + 1)) + min;
}
function show_one_bgdtxt(bgdtxt, groupIndex) {
  try {
    if (!cycle[groupIndex]) return; // 防止动画未完成
    cycle[groupIndex] = false;
    updatabgdtxt(bgdtxt);
    const text = bgdtxt.textContent.trim(); //获取文本内容并去除首尾空格
    const show_time = 200; //ms
    bgdtxt.innerHTML = "";
    bgdtxt.style.display = "block";
    text.split("").forEach((char, index) => {
      const span = document.createElement("span");
      span.className = "char";
      span.style.animation = `showAndHide 3s forwards`; // 动态设置文字显示持续时间
      span.style.animationDelay = `${(index * show_time) / 1000}s`; // 动态设置延迟（每字符 0.2 秒）
      span.textContent = char;
      bgdtxt.appendChild(span);
    });
    setTimeout(() => {
      // updatabgdtxt(bgdtxt);
      cycle[groupIndex] = true;
      show_one_bgdtxt(bgdtxt, groupIndex);
    }, 2000 + text.length * show_time + getRandomInt(0, 500)); //200不知道干嘛的
  } catch (error) {
    console.error("Error showing background text:", error);
  }
}
