<template>
  <div class="main" :class="{ redBg: isRed, blueBg: isBlue, greenBg: isGreen }">
    <div>这里放小说</div>
  </div>
  <div class="bottom-side">
    <div>这里放按钮-来改变背景颜色</div>
    <button @click="changeColor('red')" class="color1">红色背景</button>
    <button @click="changeColor('blue')" class="color2">蓝色背景</button>
    <button @click="changeColor('green')" class="color3">绿色背景</button>
    <button @click="changeColor('default')" class="color4">默认背景</button>
    <div>这里来放改变字体大小</div>
    <button @click="changeFontSize" class="font1">改变字体大小</button>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const isRed = ref(false)
const isBlue = ref(false)
const isGreen = ref(false)
const fontSize = ref('16px')

// const bgColor = ref('default') // 跟踪当前背景颜色

const changeColor = (color) => {
    // 重置所有颜色状态
    isRed.value = false
    isBlue.value = false
    isGreen.value = false

    if(color === 'red') {
      isRed.value = true
    } else if(color === 'blue') {
      isBlue.value = true
    } else if(color === 'green') {
      isGreen.value = true
    }
  
}

const changeFontSize = () => {
    // 切换字体大小，在16px和24px之间交替变化
  fontSize.value = fontSize.value === '16px' ? '24px' : '16px'
}
</script>

<style>
.main {
  background-color: rgb(120, 120, 120);
  font-size: v-bind(fontSize);
}

.main.redBg {
  background-color: red;
}
.main.blueBg {
  background-color: blue;
}
.main.greenBg {
  background-color: green;
}


.bottom-side {
  background-color: yellow;
}

.color1 {
  background-color: red;
}
.color2 {
  background-color: blue;
}
.color3 {
  background-color: green;
}
.color4 {
  background-color: rgb(120, 120, 120);
}

.font1 {
  background-color: rgb(243, 244, 245);
}
</style>