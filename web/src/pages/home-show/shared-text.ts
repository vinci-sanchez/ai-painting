import { ref } from "vue";

export const sharedText = ref("");

export function setSharedText(value: string) {
  sharedText.value = value;
}
