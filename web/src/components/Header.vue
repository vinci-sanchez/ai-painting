<template>
  <header>
    <div id="navigation"></div>
    <nav class="navbar navbar-expand-lg navbar-dark header-gradient">
      <div class="container">
        <a class="navbar-brand" href="#">Text-to-manga Studio</a>
        <span
          style="
            font-size: 13px;
            margin-bottom: -10px;
            user-select: none;
            color: #d0bcff;
          "
          >文生画</span
        >
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarNav"
          aria-controls="navbarNav"
          aria-expanded="false"
          aria-label="切换导航"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav ms-auto">
            <li class="nav-item">
              <router-link class="nav-link nav-link--button" to="/"
                >仪表盘</router-link
              >
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
          </ul>
        </div>
      </div>
    </nav>
  </header>
</template>
<script setup></script>
<style scoped>
.header-gradient {
  background: linear-gradient(90deg, #84baea 0%, #a9bbda 55%, #689cf5 100%);
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
</style>
