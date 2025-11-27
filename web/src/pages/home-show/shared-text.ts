import { ref } from "vue";

export const sharedText = ref("");

export function setSharedText(value: string) {
  sharedText.value = value;
  
}
export const Storyboard = ref("");

export function setStoryboard(value: string) {
  Storyboard.value = value;

}
export const comicText = ref("");

export function setComicText(value: string) {
  comicText.value = value;
}

export const comicimage = ref<string[]>([]);

export function setComicImage(value: string[]) {
  comicimage.value = value;
}

export const comicpage = ref<string[]>([]);

export function setComicPage(value: string[]) {
  comicpage.value = value;
}

export type ParameterDraft = {
  title?: string;
  scene?: string;
  prompt?: string;
  dialogue?: string;
  character?: string;
  style?: string;
  color?: string;
  customTags?: string[];
};

export const parameterDraft = ref<ParameterDraft | null>(null);

export function setParameterDraft(value: ParameterDraft | null) {
  parameterDraft.value = value;
}
