import config from './config.js';
const API_KEY = config.API_KEY;
const BASE_URL = config.BASE_URL;
const TEXT_MODEL = config.TEXT_MODEL;
const IMAGE_MODEL = config.IMAGE_MODEL;

// 新增: 处理提示词标签逻辑
document.addEventListener('DOMContentLoaded', () => {
    const promptInput = document.getElementById('promptInput');
    const promptTags = document.getElementById('promptTags');

    promptInput.addEventListener('keydown', (event) => {
        if (event.key === 'Enter' && promptInput.value.trim() !== '') {
            event.preventDefault();
            addTag(promptInput.value.trim());
            promptInput.value = '';
        }
    });

    function addTag(text) {
        const tag = document.createElement('span');
        tag.className = 'badge bg-secondary me-2 mb-2';
        tag.innerHTML = `${text} <button type="button" class="btn-close btn-close-white ms-1" aria-label="Close"></button>`;
        tag.querySelector('.btn-close').addEventListener('click', () => tag.remove());
        promptTags.appendChild(tag);
    }

    document.querySelector('.btn-primary').addEventListener('click', generateComic);
});

async function generateComic() {
    const inputText = document.getElementById('textInput').value;
    if (!inputText) return alert('请输入文本');

    // 收集提示词标签
    const promptTags = document.getElementById('promptTags');
    const prompts = Array.from(promptTags.querySelectorAll('.badge')).map(tag => tag.textContent.trim().replace(/×$/, '').trim());

    // 步骤1: 分段和提取关键词
    const segments = await segmentAndExtractKeywords(inputText);
    const panelsDiv = document.getElementById('panels');
    panelsDiv.innerHTML = '<p>生成中...</p>';

    // 步骤2: 为每段生成故事板和图像
    for (let i = 0; i < Math.min(segments.length, 4); i++) { // 限制最多4个面板
        const segment = segments[i];
        const prompt = `${segment.text}。关键词：${segment.keywords.map(k => k.keyword).join(', ')}。${prompts.length > 0 ? `附加提示词: ${prompts.join(', ')}` : ''}`;
        
        // 生成故事板描述
        const storyboard = await generateStoryboard(prompt);
        const panelDesc = storyboard.split('\n').filter(p => p.trim())[i % 4] || segment.text;

        // 生成图像
        const imageUrl = await generateImage(panelDesc, `漫画风格，面板${i + 1}`, prompts);
        panelsDiv.innerHTML += `
            <div class="comic-panel">
                <h6>面板 ${i + 1}</h6>
                <img src="${imageUrl}" alt="面板${i + 1}">
                <p>${panelDesc}<br><b>关键词:</b> ${segment.keywords.map(k => k.keyword).join(', ')}</p>
            </div>`;
    }
}

// API 调用：分段和关键词提取
async function segmentAndExtractKeywords(text) {
    const response = await fetch('http://localhost:3000/api/segment_keywords', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text })
    });
    if (!response.ok) throw new Error('分段和关键词提取失败');
    return await response.json();
}

// API 调用：生成故事板
async function generateStoryboard(text) {
    const response = await fetch('http://localhost:3000/api/text', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            model: TEXT_MODEL,
            messages: [
                { role: 'system', content: '你是一个漫画创作者。将输入文本转换为3-4个面板的故事板，每行描述一个面板，包括场景、角色和简单对话。用中文回复，每行一个面板。' },
                { role: 'user', content: text }
            ],
            max_tokens: 300,
            temperature: 0.7
        })
    });
    const data = await response.json();
    return data.choices[0].message.content;
}

// API 调用：生成图像
async function generateImage(prompt, style, prompts = []) {
    const extraPrompts = prompts.length > 0 ? `, ${prompts.join(', ')}` : '';
    const fullPrompt = `${prompt}。${style}，黑白线稿漫画风格，详细插图${extraPrompts}`;
    const response = await fetch('http://localhost:3000/api/image', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            model: IMAGE_MODEL,
            prompt: fullPrompt,
            n: 1,
            size: '512x512',
            response_format: 'url'
        })
    });
    const data = await response.json();
    return data.data[0].url;
}