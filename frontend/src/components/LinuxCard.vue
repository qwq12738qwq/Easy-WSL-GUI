<script setup>

const getImageUrl = (name) => {
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

</script>

<template>
  <div class="card hover-lift" @click="$emit('action')">
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
      <button class="btn btn-primary card-btn" @click.stop="$emit('action')">
        一键安装
      </button>
    </div>
  </div>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: column;
  padding: var(--spacing-lg);
  height: 100%;
  cursor: pointer;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-normal);
  position: relative;
  overflow: hidden;
}

.card:hover {
  border-color: var(--color-brand);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-md);
}

.icon-container {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-hover);
  border-radius: var(--radius-md);
  padding: 8px;
}

.local-icon {
  width: 100%;  
  height: 100%;
  object-fit: contain;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-text-tertiary);
  transition: all var(--transition-normal);
}

.status-dot.online { 
  background: var(--color-success); 
  box-shadow: 0 0 8px var(--color-success); 
}

.card-body h4 {
  margin: var(--spacing-sm) 0 var(--spacing-xs) 0;
  font-size: var(--font-size-lg);
  color: var(--color-text-primary);
}

.card-body p {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.card-footer {
  margin-top: auto;
  padding-top: var(--spacing-md);
}

.card-btn {
  width: 100%;
  opacity: 0.8;
  transform: translateY(5px);
  transition: all var(--transition-normal);
}

.card:hover .card-btn {
  opacity: 1;
  transform: translateY(0);
}
</style>