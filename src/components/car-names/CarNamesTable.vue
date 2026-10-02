<template>
  <table class="data-table">
    <thead>
      <tr>
        <th>Car Name</th>
        <th>Brand</th>
        <th>Notes</th>
        <th>Big Car</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody>
      <tr v-if="loading">
        <td colspan="5" class="table-status">Loading car names…</td>
      </tr>
      <tr v-else-if="error">
        <td colspan="5" class="table-status table-status-error">
          {{ error }}
          <button class="btn retry-btn" @click="$emit('retry')">Retry</button>
        </td>
      </tr>
      <tr v-else-if="totalCount === 0">
        <td colspan="5" class="table-status">No car names yet.</td>
      </tr>
      <tr v-else-if="carNames.length === 0">
        <td colspan="5" class="table-status">No results match your search.</td>
      </tr>
      <tr v-for="cn in carNames" v-else :key="cn.id">
        <td>{{ cn.car_name }}</td>
        <td>
          <span v-if="cn.brand">{{ cn.brand }}</span>
          <span v-else class="no-logo" title="This car name has no brand assigned">No brand</span>
        </td>
        <td>{{ cn.notes }}</td>
        <td>{{ cn.is_big_car ? 'Yes' : 'No' }}</td>
        <td>
          <button class="btn edit-btn" @click="$emit('edit', cn)">Edit</button>
          <button class="btn media-btn" @click="$emit('open-media', cn)">Photos/Videos</button>
          <button v-if="isAdmin" class="btn delete-btn" :disabled="deleting" @click="$emit('delete', cn)">Delete</button>
        </td>
      </tr>
    </tbody>
  </table>
</template>
<script setup>
defineProps({ carNames: { type: Array, default: () => [] }, loading: { type: Boolean, default: false }, error: { type: String, default: '' }, isAdmin: { type: Boolean, default: false }, deleting: { type: Boolean, default: false }, totalCount: { type: Number, default: 0 } })
defineEmits(['edit', 'delete', 'open-media', 'retry'])
</script>
<style scoped>
.data-table{width:100%;border-collapse:collapse;margin-top:20px}
.data-table th,.data-table td{padding:12px;text-align:left;border-bottom:1px solid #ddd}
.data-table th{background-color:#f8f9fa;font-weight:600}
.data-table tbody tr:hover{background-color:#f5f5f5}
.table-status{text-align:center;padding:20px;color:#6b7280}
.table-status-error{color:#b91c1c}
.retry-btn{margin-left:8px;padding:4px 8px;border:none;border-radius:4px;background:#3b82f6;color:#fff;cursor:pointer}
.no-logo{color:#9ca3af;font-style:italic;font-size:.875rem}
.btn{padding:6px 12px;border:none;border-radius:4px;margin-right:8px;cursor:pointer}
.edit-btn{background:#3b82f6;color:#fff}
.edit-btn:hover{background:#2563eb}
.delete-btn{background:#ef4444;color:#fff}
.delete-btn:hover{background:#dc2626}
.media-btn{background:#8b5cf6;color:#fff}
.media-btn:hover{background:#7c3aed}
</style>
