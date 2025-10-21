import config from './config.js';
        const API_KEY = config.API_KEY;
        const BASE_URL = config.BASE_URL;
        const TEXT_MODEL = config.TEXT_MODEL; 
        const IMAGE_MODEL = config.IMAGE_MODEL;  

        
        async function generateComic() {
            const inputText = document.getElementById('textInput').value;
            if (!inputText) return alert('请输入文本');

            // 步骤1: 生成故事板（3-4个面板的描述）
            const storyboard = await generateStoryboard(inputText);
            const panels = storyboard.split('\n').filter(p => p.trim());  // 按面板分割

            // 步骤2: 为每个面板生成图像
            const panelsDiv = document.getElementById('panels');
            panelsDiv.innerHTML = '<p>生成中...</p>';
            for (let i = 0; i < panels.length; i++) {
                const imageUrl = await generateImage(panels[i], `漫画风格，面板${i + 1}`);
                panelsDiv.innerHTML += `<div class="comic-panel"><h6>面板 ${i + 1}</h6><img src="${imageUrl}" alt="面板${i + 1}"><p>${panels[i]}</p></div>`;
            }
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
        async function generateImage(prompt, style) {
            const fullPrompt = `${prompt}。${style}，黑白线稿漫画风格，详细插图`;
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