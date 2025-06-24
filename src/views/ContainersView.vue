<template>
  <div class="containers-view">
    <div class="header">
      <h2>
        <i class="fas fa-box"></i>
        Containers Management
      </h2>
    </div>

    <div class="content">
      <ContainersCrud />
    </div>
  </div>
</template>

<script setup>
import ContainersCrud from '../components/containers/ContainersCrud.vue'
</script>

<style scoped>
.containers-view {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0;
  color: #2c3e50;
}

.header h2 i {
  color: #3498db;
}

.content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>
