<template>
  <div class="toolbar">
    <h2>Car Names</h2>
    <div class="actions">
      <button class="add-btn" @click="$emit('add-car')">Add Car</button>
      <button class="brands-btn" :class="{ active: brandsOpen }" @click="$emit('toggle-brands')">
        Brands
        <span class="toolbar-count">{{ brandsCount }}</span>
      </button>
    </div>
  </div>
</template>
<script setup>
defineProps({ carNamesCount: { type: Number, default: 0 }, brandsCount: { type: Number, default: 0 }, brandsOpen: { type: Boolean, default: false } })
defineEmits(['add-car', 'toggle-brands'])
</script>
<style scoped>
.toolbar{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}
.actions{display:flex;gap:8px}
.add-btn{padding:8px 16px;border:none;border-radius:4px;cursor:pointer;background-color:#10b981;color:#fff}
.add-btn:hover{background-color:#059669}
.brands-btn{display:inline-flex;align-items:center;gap:6px;padding:8px 16px;border:none;border-radius:4px;cursor:pointer;background-color:#6366f1;color:#fff;transition:.15s}
.brands-btn:hover{background-color:#4f46e5}
.brands-btn.active{background-color:#4f46e5;box-shadow:0 1px 2px rgba(0,0,0,.06)}
.toolbar-count{display:inline-block;padding:0 6px;border-radius:10px;background:rgba(255,255,255,.25);font-size:.85em;line-height:1.5}
</style>
