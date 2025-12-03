<template>
  <div>
    <Header />
    <router-view />
  </div>
</template>

<script setup>
import { watch } from "vue";
import Header from "./components/Header.vue";
import { currentUser } from "./pages/home-show/auth-user";
import {
  clearUserComics,
  refreshUserComics,
} from "./pages/home-show/user-comics";

watch(
  () => currentUser.value?.username,
  (username) => {
    if (username) {
      refreshUserComics(username);
    } else {
      clearUserComics();
    }
  },
  { immediate: true }
);
</script>

<!-- <style>
body {
  font-family: Arial, sans-serif;
  margin: 0;
  padding: 0;
}
</style> -->
