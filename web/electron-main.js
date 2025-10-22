//const { app, BrowserWindow } = require('electron');
import{ app , BrowserWindow } from 'electron';
//const path = require('path');
import path from 'path';
import { fileURLToPath } from 'url';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function createWindow() {
    const win = new BrowserWindow({
        width: 800,
        height: 600,
        webPreferences: {
            preload: path.join(__dirname, 'preload.js'),  // 如果需要预加载脚本，可添加
            nodeIntegration: true,  // 启用 Node.js 集成（根据需要）
            contextIsolation: false  // 根据安全需求调整
        }
    });

    // 在开发模式下加载 Vite dev server
    if (process.env.NODE_ENV === 'development') {
        win.loadURL('http://localhost:5173');  // Vite 默认端口 5173
    } else {
        // 在生产模式下加载构建后的 index.html
        win.loadFile(path.join(__dirname, 'dist/index.html'));
    }
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
    }
});