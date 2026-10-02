<template>
  <table class="data-table">
    <thead>
      <tr>
        <th>Logo</th>
        <th>Brand</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody>
      <tr v-if="loading">
        <td colspan="3" class="table-status">Loading brands…</td>
      </tr>
      <tr v-else-if="error">
        <td colspan="3" class="table-status table-status-error">
          {{ error }}
          <button class="btn retry-btn" @click="$emit('retry')">Retry</button>
        </td>
      </tr>
      <tr v-else-if="brands.length === 0">
        <td colspan="3" class="table-status">No brands yet.</td>
      </tr>
      <tr v-for="brand in brands" v-else :key="brand.id" class="brand-row" tabindex="0" role="button" :aria-label="`View models of ${brand.brand}`" @click="$emit('select', brand)" @keydown.enter.prevent="$emit('select', brand)" @keydown.space.prevent="$emit('select', brand)">
        <td>
          <img v-if="brand.logo_path" :src="getFileUrl(brand.logo_path)" :alt="brand.brand" class="brand-logo" />
          <span v-else class="no-logo">No logo</span>
        </td>
        <td>{{ brand.brand }}</td>
        <td>
          <button class="btn edit-btn" @click.stop="$emit('edit', brand)">Edit</button>
          <button v-if="isAdmin" class="btn delete-btn" :disabled="deleting" @click.stop="$emit('delete', brand)">Delete</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
<script setup>
import { useApi } from '../../composables/useApi'
const { getFileUrl } = useApi()
defineProps({ brands: { type: Array, default: () => [] }, loading: { type: Boolean, default: false }, error: { type: String, default: '' }, isAdmin: { type: Boolean, default: false }, deleting: { type: Boolean, default: false } })
defineEmits(['select', 'edit', 'delete', 'retry'])
</script>
<style scoped>
.data-table{width:100%;border-collapse:collapse;margin-top:20px}
.data-table th,.data-table td{padding:12px;text-align:left;border-bottom:1px solid #ddd}
.data-table th{background-color:#f8f9fa;font-weight:600}
.data-table tbody tr:hover{background-color:#f5f5f5}
.table-status{text-align:center;padding:20px;color:#6b7280}
.table-status-error{color:#b91c1c}
.retry-btn{margin-left:8px;padding:4px 8px;border:none;border-radius:4px;background:#3b82f6;color:#fff;cursor:pointer}
.brand-row{cursor:pointer}
.brand-row:focus-visible{outline:2px solid #3b82f6;outline-offset:-2px}
.brand-logo{width:40px;height:40px;object-fit:contain;border-radius:4px}
.no-logo{color:#9ca3af;font-style:italic;font-size:.875rem}
.btn{padding:6px 12px;border:none;border-radius:4px;margin-right:8px;cursor:pointer}
.edit-btn{background:#3b82f6;color:#fff}
.edit-btn:hover{background:#2563eb}
.delete-btn{background:#ef4444;color:#fff}
.delete-btn:hover{background:#dc2626}
</style>
