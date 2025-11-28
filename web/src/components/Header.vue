<template>
  <header>
    <div id="navigation"></div>
    <nav class="navbar  navbar-dark header-gradient">
      <div class="container">
        <a class="navbar-brand" href="#">Text-to-manga Studio</a>
        <!-- <span
          style="
            font-size: 13px;
            margin-bottom: -10px;
            user-select: none;
            color: #d0bcff;
          "
          >文生漫</span
        > -->

        <!-- 让汉堡 + 下拉面板包在一个容器里，方便 hover -->
        <div
          class="toggler-wrapper"
          @mouseenter="menuOpen = true"
          @mouseleave="menuOpen = false"
        >
          <button
            class="navbar-toggler"
            type="button"
            aria-label="切换导航"
          >
            <span class="navbar-toggler-icon"></span>
          </button>

          <!-- hover 下拉菜单（向下弹出） -->
          <transition name="fade-slide">
            <div v-show="menuOpen" class="hover-menu">
              <div class="hover-menu__item">
                <button class="hover-menu__btn" @click="go({ name: 'home-welcome' })">
                  仪表盘
                </button>
              </div>
              <div class="hover-menu__item">
                <button class="hover-menu__btn" @click="go('/login')">
                  登录
                </button>
              </div>
              <div class="hover-menu__item">
                <button class="hover-menu__btn" @click="go('/signup')">
                  注册
                </button>
              </div>
            </div>
          </transition>
        </div>

        <div class="collapse navbar-collapse" id="navbarNav">
          <!-- <ul class="navbar-nav ms-auto">
            <li class="nav-item">
              <router-link
                class="nav-link nav-link--button"
                :to="{ name: 'home-welcome' }"
              >
                仪表盘
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link nav-link--button" to="/login"
                >登录</router-link
              >
            </li>
            <li class="nav-item">
              <router-link class="nav-link nav-link--button" to="/signup" style=""
                >注册</router-link
              >
            </li>
          </ul> -->
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup>
import { ref } from "vue"
import { useRouter } from "vue-router"

const menuOpen = ref(false)
const router = useRouter()

function go(target) {
  menuOpen.value = false
  router.push(target)
}
</script>

<style scoped>
.header-gradient {
  background: linear-gradient(90deg, #84baea 0%, #a9bbda 55%, #689cf5 100%);
  height: 72px;
  display: flex;
  align-items: center;
}

.header-gradient .container {
  display: flex;
  align-items: center;
  flex-wrap: nowrap;
  height: 100%;
}

.header-gradient .navbar-brand,
.header-gradient .nav-link {
  color: #ffffff;
}

.header-gradient .nav-link:hover,
.header-gradient .navbar-brand:hover {
  color: #d0e6ff;
}

.nav-link--button {
  text-decoration: none;
}

/* ===== 新增：汉堡 hover 下拉 ===== */
.toggler-wrapper {
  position: relative;
  margin-left: auto; /* 保持汉堡在右侧（不影响你其它布局） */
}

.hover-menu {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 160px;
  background: rgba(255, 255, 255, 0.98);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 8px;
  z-index: 9999;
}

.hover-menu__item + .hover-menu__item {
  margin-top: 6px;
}

.hover-menu__btn {
  width: 100%;
  border: none;
  background: transparent;
  padding: 10px 12px;
  border-radius: 8px;
  text-align: left;
  cursor: pointer;
  font-size: 14px;
  color: #333;
}

.hover-menu__btn:hover {
  background: #eff5ff;
  color: #1f5eff;
}

/* 小动画 */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.15s ease;
}
.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
