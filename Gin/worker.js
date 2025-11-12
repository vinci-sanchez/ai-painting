const DEFAULT_BASE_URL = "https://ark.cn-beijing.volces.com/api/v3";
const DEFAULT_COLAB_ENDPOINT = "https://articular-proportionable-alverta.ngrok-free.dev/";
const DEFAULT_MODEL_ID = "ep-20251026170710-8pgpm";

const CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET,POST,OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type,Authorization",
};

export default {
  async fetch(request, env) {
    if (request.method === "OPTIONS") {
      return new Response(null, { status: 204, headers: withCors() });
    }

    const url = new URL(request.url);

    try {
      if (url.pathname === "/api/hello" && request.method === "GET") {
        return jsonResponse({ message: "hello" });
      }

      if (url.pathname === "/api/text" && request.method === "POST") {
        return handleText(request, env);
      }

      if (url.pathname === "/api/image" && request.method === "POST") {
        return handleArkImage(request, env);
      }

      if (url.pathname === "/generate_image" && request.method === "POST") {
        return handleColabImage(request, env);
      }

      if (url.pathname === "/api/crawl" && request.method === "POST") {
        return handleCrawl(request, env);
      }

      return jsonResponse({ error: "Not Found" }, 404);
    } catch (error) {
      console.error("Worker execution failed", error);
      return errorResponse(error.message || "Internal Server Error");
    }
  },
};

function withCors(headers = {}) {
  return { ...CORS_HEADERS, ...headers };
}

function jsonResponse(payload, status = 200, headers = {}) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: withCors({ "Content-Type": "application/json", ...headers }),
  });
}

function errorResponse(message, status = 500) {
  return jsonResponse({ error: message }, status);
}

async function handleText(request, env) {
  const apiKey = env?.API_KEY;
  if (!apiKey) {
    return errorResponse("API_KEY is not configured", 500);
  }

  const baseURL = env?.BASE_URL || DEFAULT_BASE_URL;
  const body = await request.text();
  if (!body) {
    return errorResponse("Request body is required", 400);
  }

  const upstream = await fetch(joinURL(baseURL, "/chat/completions"), {
    method: "POST",
    headers: {
      Authorization: `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body,
  });

  const text = await upstream.text();
  let payload;
  try {
    payload = JSON.parse(text);
  } catch {
    payload = null;
  }

  if (!upstream.ok) {
    const message =
      payload?.error?.message ||
      `Upstream text API returned status ${upstream.status}`;
    return errorResponse(message, 502);
  }

  const message = payload?.choices?.[0]?.message?.content;
  if (typeof message === "undefined") {
    return errorResponse("Upstream text API response is missing choices", 500);
  }

  return jsonResponse({ data: { message } });
}

async function handleArkImage(request, env) {
  const apiKey = env?.API_KEY;
  if (!apiKey) {
    return errorResponse("API_KEY is not configured", 500);
  }

  let requestData;
  try {
    requestData = await request.json();
  } catch {
    return errorResponse("Request body must be valid JSON", 400);
  }

  if (!requestData?.prompt) {
    return errorResponse("prompt is required", 400);
  }

  const baseURL = env?.BASE_URL || DEFAULT_BASE_URL;
  const model = env?.ARK_MODEL_ID || DEFAULT_MODEL_ID;
  const arkPayload = {
    model,
    prompt: requestData.prompt,
    role: requestData.role || "",
    storyboard: requestData.storyboard || "",
  };

  const upstream = await fetch(joinURL(baseURL, "/images/generations"), {
    method: "POST",
    headers: {
      Authorization: `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify(arkPayload),
  });

  const text = await upstream.text();
  let payload;
  try {
    payload = JSON.parse(text);
  } catch {
    payload = null;
  }

  if (!upstream.ok) {
    const message =
      payload?.error?.message ||
      `Ark image API returned status ${upstream.status}`;
    return errorResponse(message, 502);
  }

  if (!payload?.data?.length) {
    return errorResponse("Ark image API returned no data", 500);
  }

  return new Response(text, {
    status: upstream.status,
    headers: withCors({ "Content-Type": "application/json" }),
  });
}

async function handleColabImage(request, env) {
  let requestData;
  try {
    requestData = await request.json();
  } catch {
    return errorResponse("Request body must be valid JSON", 400);
  }

  if (!requestData?.prompt) {
    return errorResponse("prompt is required", 400);
  }

  const colabEndpoint = env?.COLAB_ENDPOINT || DEFAULT_COLAB_ENDPOINT;
  const payload = {
    role: requestData.role || "",
    prompt: requestData.prompt,
    storyboard: requestData.storyboard || "",
  };

  const upstream = await fetch(joinURL(colabEndpoint, "/generate"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  const text = await upstream.text();

  if (!upstream.ok) {
    let errorBody;
    try {
      errorBody = JSON.parse(text);
    } catch {
      errorBody = null;
    }
    const message =
      errorBody?.error || `Colab endpoint returned status ${upstream.status}`;
    return errorResponse(message, 502);
  }

  return new Response(text, {
    status: upstream.status,
    headers: withCors({
      "Content-Type": upstream.headers.get("content-type") || "application/json",
    }),
  });
}

async function handleCrawl(request, env) {
  const crawlerEndpoint = (env?.CRAWLER_ENDPOINT || "").trim();
  if (!crawlerEndpoint) {
    return errorResponse(
      "CRAWLER_ENDPOINT is not configured, crawler requests cannot be handled",
      503
    );
  }

  const body = await request.text();
  const url =
    crawlerEndpoint.endsWith("/api/crawl") || crawlerEndpoint.includes("?")
      ? crawlerEndpoint
      : `${crawlerEndpoint.replace(/\/$/, "")}/api/crawl`;

  const upstream = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body,
  });

  const text = await upstream.text();
  return new Response(text, {
    status: upstream.status,
    headers: withCors({
      "Content-Type": upstream.headers.get("content-type") || "application/json",
    }),
  });
}

function joinURL(base, path) {
  try {
    return new URL(path, base).toString();
  } catch {
    const normalizedBase = (base || "").replace(/\/$/, "");
    const normalizedPath = path.startsWith("/") ? path : `/${path}`;
    return `${normalizedBase}${normalizedPath}`;
  }
}
