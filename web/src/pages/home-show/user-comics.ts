import { readonly, ref } from "vue";
import config from "../config.json";
import { currentUser } from "./auth-user";

const BACK_URL =
  (config as Record<string, string | undefined>).BACK_URL ??
  "http://localhost:3000";

export type StoredComic = {
  id: number;
  title: string;
  pageNumber: number;
  imageBase64: string;
  metadata: Record<string, unknown> | null;
  createdAt: string;
  isShared: boolean;
  shareMessage: string;
  likesCount: number;
  commentsCount: number;
};

const userComicsRef = ref<StoredComic[]>([]);
const userComicsLoading = ref(false);

export const userComics = readonly(userComicsRef);
export const isUserComicsLoading = readonly(userComicsLoading);

type RawComicResponse = {
  id: number;
  title: string;
  page_number?: number;
  image_base64: string;
  metadata?: unknown;
  created_at: string;
  is_shared?: boolean;
  share_message?: string;
  likes_count?: number;
  comments_count?: number;
};

const parseMetadata = (
  value: unknown
): Record<string, unknown> | null => {
  if (!value) return null;
  if (typeof value === "string") {
    try {
      return JSON.parse(value);
    } catch {
      return null;
    }
  }
  if (typeof value === "object") {
    return value as Record<string, unknown>;
  }
  return null;
};

const normalizeComic = (payload: RawComicResponse): StoredComic => {
  return {
    id: payload.id,
    title: payload.title,
    pageNumber:
      typeof payload.page_number === "number" && payload.page_number > 0
        ? payload.page_number
        : payload.id,
    imageBase64: payload.image_base64,
    metadata: parseMetadata(payload.metadata),
    createdAt: payload.created_at,
    isShared: Boolean(payload.is_shared),
    shareMessage:
      typeof payload.share_message === "string"
        ? payload.share_message
        : "",
    likesCount:
      typeof payload.likes_count === "number" && payload.likes_count >= 0
        ? payload.likes_count
        : 0,
    commentsCount:
      typeof payload.comments_count === "number" && payload.comments_count >= 0
        ? payload.comments_count
        : 0,
  };
};

export const clearUserComics = () => {
  userComicsRef.value = [];
};

export const refreshUserComics = async (username?: string) => {
  const targetName = (username ?? currentUser.value?.username)?.trim();
  if (!targetName) {
    clearUserComics();
    return;
  }
  userComicsLoading.value = true;
  try {
    const response = await fetch(
      `${BACK_URL}/api/users/${encodeURIComponent(targetName)}/comics`
    );
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      console.warn("无法同步历史漫画", data?.error);
      userComicsRef.value = [];
      return;
    }
    const comicsArray = Array.isArray(data?.comics) ? data.comics : [];
    userComicsRef.value = comicsArray.map((item: RawComicResponse) =>
      normalizeComic(item)
    );
  } catch (error) {
    console.error("拉取历史漫画失败", error);
    userComicsRef.value = [];
  } finally {
    userComicsLoading.value = false;
  }
};

type SaveComicPayload = {
  title: string;
  pageNumber: number;
  imageBase64?: string;
  imageUrl?: string;
  metadata?: Record<string, unknown>;
  isShared?: boolean;
  shareMessage?: string;
};

export const saveComicForCurrentUser = async (
  payload: SaveComicPayload
): Promise<StoredComic | null> => {
  const username = currentUser.value?.username?.trim();
  if (!username) {
    return null;
  }
  if (!payload.imageBase64 && !payload.imageUrl) {
    throw new Error("缺少漫画图片数据");
  }
  const response = await fetch(
    `${BACK_URL}/api/users/${encodeURIComponent(username)}/comics`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: payload.title,
        page_number: payload.pageNumber,
        image_base64: payload.imageBase64 ?? "",
        image_url: payload.imageUrl ?? "",
        metadata: payload.metadata ?? {},
        is_shared: payload.isShared ?? false,
        share_message: payload.shareMessage ?? "",
      }),
    }
  );
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data?.comic) {
    throw new Error(data?.error || "保存漫画失败");
  }
  const normalized = normalizeComic(data.comic as RawComicResponse);
  userComicsRef.value = [normalized, ...userComicsRef.value];
  return normalized;
};

export const deleteComicForCurrentUser = async (comicId: number) => {
  const username = currentUser.value?.username?.trim();
  if (!username) {
    throw new Error("用户未登录");
  }
  if (!comicId || comicId <= 0) {
    throw new Error("漫画ID无效");
  }

  const response = await fetch(
    `${BACK_URL}/api/users/${encodeURIComponent(username)}/comics/${comicId}`,
    { method: "DELETE" }
  );
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data?.error || "删除漫画失败");
  }
  userComicsRef.value = userComicsRef.value.filter(
    (comic) => comic.id !== comicId
  );
};

type ShareUpdatePayload = {
  comicId: number;
  isShared: boolean;
  shareMessage: string;
};

export const updateComicShareForCurrentUser = async (
  payload: ShareUpdatePayload
) => {
  const username = currentUser.value?.username?.trim();
  if (!username) {
    throw new Error("用户未登录");
  }
  const response = await fetch(
    `${BACK_URL}/api/users/${encodeURIComponent(username)}/comics/${
      payload.comicId
    }/share`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        is_shared: payload.isShared,
        share_message: payload.shareMessage,
      }),
    }
  );
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data?.comic) {
    throw new Error(data?.error || "更新分享设置失败");
  }
  const updated = normalizeComic(data.comic as RawComicResponse);
  userComicsRef.value = userComicsRef.value.map((comic) =>
    comic.id === updated.id ? updated : comic
  );
  return updated;
};

export const likeComic = async (comicId: number) => {
  if (!comicId) {
    throw new Error("无效的漫画ID");
  }
  const response = await fetch(`${BACK_URL}/api/comics/${comicId}/like`, {
    method: "POST",
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data?.error || "点赞失败");
  }
  return typeof data.likes === "number" ? data.likes : 0;
};

export type ComicComment = {
  id: number;
  comicId: number;
  author: string;
  content: string;
  createdAt: string;
};

export const fetchComicComments = async (
  comicId: number
): Promise<ComicComment[]> => {
  if (!comicId) {
    return [];
  }
  const response = await fetch(`${BACK_URL}/api/comics/${comicId}/comments`);
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data?.error || "无法获取留言");
  }
  const list = Array.isArray(data?.comments) ? data.comments : [];
  return list.map((item: Record<string, unknown>) => ({
    id: Number(item.id) || Date.now(),
    comicId: Number(item.comic_id) || comicId,
    author: typeof item.author === "string" ? item.author : "游客",
    content: typeof item.content === "string" ? item.content : "",
    createdAt: typeof item.created_at === "string" ? item.created_at : "",
  }));
};

export const addComicComment = async (
  comicId: number,
  payload: { author?: string; content: string }
): Promise<ComicComment> => {
  if (!comicId) {
    throw new Error("无效的漫画ID");
  }
  const response = await fetch(`${BACK_URL}/api/comics/${comicId}/comments`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      author: payload.author ?? "游客",
      content: payload.content,
    }),
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data?.comment) {
    throw new Error(data?.error || "留言失败");
  }
  const item = data.comment as Record<string, unknown>;
  return {
    id: Number(item.id) || Date.now(),
    comicId: Number(item.comic_id) || comicId,
    author: typeof item.author === "string" ? item.author : "游客",
    content: typeof item.content === "string" ? item.content : "",
    createdAt: typeof item.created_at === "string" ? item.created_at : "",
  };
};

export const fetchFeaturedComics = async (
  limit = 5
): Promise<StoredComic[]> => {
  const response = await fetch(
    `${BACK_URL}/api/comics/shared/featured?limit=${limit}`
  );
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data?.error || "无法获取分享漫画");
  }
  const comics = Array.isArray(data?.comics) ? data.comics : [];
  return comics.map((item: RawComicResponse) => normalizeComic(item));
};
