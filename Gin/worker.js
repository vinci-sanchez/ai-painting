import wasm from "./main.wasm";
import "./wasm_exec.js";

// 载入 Go runtime
const go = new Go();

export default {
  async fetch(request) {
    // 实例化 WebAssembly
    const { instance } = await WebAssembly.instantiateStreaming(fetch(wasm), go.importObject);

    go.run(instance);

    // 你可以在这里处理请求转发或交互
    const url = new URL(request.url);

    if (url.pathname === "/api/hello") {
      return new Response("Hello from Go (WASM in Cloudflare Worker)!", {
        headers: { "Content-Type": "text/plain" },
      });
    }

    return new Response("Go backend running in Cloudflare!", { status: 200 });
  },
};
