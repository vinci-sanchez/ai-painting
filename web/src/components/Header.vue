<template>
  <header>
    <div id="navigation"></div>
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
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
              <router-link to="/"><a class="nav-link">仪表盘</a></router-link>
            </li>
            <li class="nav-item">
              <router-link to="/login">
                <a class="nav-link">登录</a>
              </router-link>
            </li>
            <li class="nav-item">
              <router-link to="/signup">
                <a class="nav-link">注册</a>
              </router-link>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  </header>
</template>
<script setup></script>
