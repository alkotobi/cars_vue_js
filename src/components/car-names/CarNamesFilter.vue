<template>
  <div class="filter-box">
    <div class="search-box">
      <i class="fas fa-search"></i>
      <input v-model="local" class="search-input" type="text" :placeholder="placeholder" />
      <button v-if="local" type="button" class="clear-btn" @click="clear">×</button>
    </div>
    <div class="filter-status" v-if="totalCount >= 0">
      <span v-if="brandId">Showing {{ resultCount }} of {{ totalCount }} for <strong>{{ brandName }}</strong></span>
      <span v-else>Showing {{ resultCount }} of {{ totalCount }}</span>
      <button v-if="brandId" type="button" class="clear-filter" @click="$emit('clear-pin')">Show all</button>
    </div>
  </div>
</template>
<script setup>
import { computed, watch } from 'vue'
const props = defineProps({ modelValue: { type: String, default: '' }, resultCount: { type: Number, default: 0 }, totalCount: { type: Number, default: 0 }, brandId: { type: [Number, String, null], default: null }, brandName: { type: String, default: '' }, placeholder: { type: String, default: 'Search car names or brands...' } })
const emit = defineEmits(['update:modelValue', 'clear', 'clear-pin'])
const local = computed({ get: () => props.modelValue, set: (v) => emit('update:modelValue', v) })
function clear() { emit('update:modelValue', ''); emit('clear') }
watch(() => props.brandId, () => { /* no-op */ })
</script>
<style scoped>
.filter-box{margin-bottom:16px;display:flex;flex-direction:column;gap:8px}
.search-box{position:relative;display:flex;align-items:center}
.search-box i{position:absolute;left:12px;color:#9ca3af}
.search-input{width:100%;padding:8px 36px 8px 36px;border:1px solid #e5e7eb;border-radius:6px}
.clear-btn{position:absolute;right:8px;background:transparent;border:none;font-size:18px;cursor:pointer;color:#6b7280}
.filter-status{display:flex;align-items:center;gap:10px;font-size:.9rem;color:#374151}
.clear-filter{background:transparent;border:none;color:#2563eb;text-decoration:underline;cursor:pointer;padding:0}
</style>
