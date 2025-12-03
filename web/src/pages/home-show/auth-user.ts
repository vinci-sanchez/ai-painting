import { readonly, ref } from "vue";

export type AuthUser = {
  username: string;
};

const STORAGE_KEY = "ttm-current-user";
const currentUserRef = ref<AuthUser | null>(null);

const hydrateFromStorage = () => {
  currentUserRef.value = readStoredUser();
};

if (typeof window !== "undefined") {
  hydrateFromStorage();
  window.addEventListener("storage", (event) => {
    if (event.key === STORAGE_KEY) {
      hydrateFromStorage();
    }
  });
}

export const currentUser = readonly(currentUserRef);

export const setAuthUser = (user: AuthUser | null) => {
  currentUserRef.value = user;
  if (typeof window === "undefined") {
    return;
  }
  if (user) {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(user));
  } else {
    window.localStorage.removeItem(STORAGE_KEY);
  }
};

export const clearAuthUser = () => setAuthUser(null);

function readStoredUser(): AuthUser | null {
  if (typeof window === "undefined") {
    return null;
  }
  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed.username === "string") {
      return { username: parsed.username };
    }
  } catch {
    // ignore corrupted storage content
  }
  window.localStorage.removeItem(STORAGE_KEY);
  return null;
}
