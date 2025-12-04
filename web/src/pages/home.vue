<template>
  <div class="home-layout">
    <aside class="home-layout__nav">
      <LeftNavigation />
    </aside>
    <main class="home-layout__content">
      <RouterView v-slot="{ Component }">
        <KeepAlive>
          <component :is="Component" />
        </KeepAlive>
      </RouterView>
    </main>
  </div>
</template>

<script setup>
import LeftNavigation from "../components/LeftNavigation.vue";
</script>

<style scoped>
.home-layout {
  display: flex;
  min-height: calc(100vh - 72px);
  height: 100%;
}

.home-layout__nav {
  width: auto;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  /* align-self: flex-start; */
  /* align-self: stretch; height: auto; */
}

.home-layout__content {
  flex: 1;
  overflow: auto;
  padding: 16px;
}

@media (max-width: 960px) {
  .home-layout {
    flex-direction: row;
    align-items: stretch;
  }

  .home-layout__nav {
    width: auto;
    position: sticky;
    top: 0;
  }

  .home-layout__content {
    padding: 12px 8px 24px;
  }
}
</style>
