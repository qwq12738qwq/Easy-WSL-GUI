<script setup>

const getImageUrl = (name) => {
  // 打印到浏览器控制台 (F12)
  console.log("子组件收到的图片名:", name); 
  
  // URL 参数必须是静态字符串拼接，方便 Vite 在编译时识别
  return new URL(`../assets/icons/${name}.png`, import.meta.url).href
}

const props = defineProps({
  title: String,
  description: String,
  status: String,
  iconName: {
    type: String,
    default: 'UbuntuCoF' // 对应图片的文件名（不含后缀）
  }
})

defineEmits(['action'])

// 处理 Vite 动态资源路径的核心函数

</script>

<template>
  <div class="card" @click="$emit('action')">
    <div class="card-header">
      <div class="icon-container">
        <img :src="getImageUrl(iconName)" class="local-icon" alt="icon" />
      </div>
      <span class="status-dot" :class="status"></span>
    </div>
    <div class="card-body">
      <h4>{{ title }}</h4>
      <p>{{ description }}</p>
    </div>
    <div class="card-footer">

      <button
      class="card-btn" 
      @click.stop="$emit('action')">
      一键安装
    </button>

    </div>
  </div>
</template>

<style scoped>
/* 保持你原有的 .card 样式不变 */

.icon-container {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.local-icon {
  width: 48px;  
  height: 48px;
  object-fit: contain; /* 保持图片比例 */
}

/* 状态灯样式保持... */
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #d9d9d9;
}
.status-dot.online { background: #52c41a; box-shadow: 0 0 5px #52c41a; }

/* 其他样式参考你提供的代码... */
.card-btn {
  width: 100%;
  padding: 6px;
  border: none;
  background: #f5f5f5;
  border-radius: 6px;
  color: #595959;
  cursor: pointer;
}
.card:hover .card-btn { background: #1890ff; color: white; }

.card {
  display: flex;
  flex-direction: column;
  padding: 20px;
  height: 100%;
  cursor: pointer;
  background: #252526; /* 基础底色 */
  box-sizing: border-box;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.online { background: #4af626; box-shadow: 0 0 8px #4af626; }
.status-dot.offline { background: #888; }

.card-body h4 {
  margin: 12px 0 4px 0;
  font-size: 16px;
  color: #fff;
}

.card-body p {
  font-size: 13px;
  color: #888;
  line-height: 1.4;
}

.card-footer {
  margin-top: auto; /* 按钮始终在底部 */
  padding-top: 15px;
}
</style>