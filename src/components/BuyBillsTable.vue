<script setup>
import { defineProps, defineEmits } from 'vue'

const props = defineProps({
  buyBills: {
    type: Array,
    required: true
  },
  selectedBill: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['select-bill'])

const selectBill = (bill) => {
  emit('select-bill', bill)
}
</script>

<template>
  <div class="master-section">
    <div v-if="selectedBill" class="selected-id">
      Selected Purchase ID: {{ selectedBill.id }}
    </div>
    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>Date</th>
          <th>Supplier</th>
          <th>Amount</th>
          <th>Paid</th>
          <th>Balance</th>
        </tr>
      </thead>
      <tbody>
        <tr 
          v-for="bill in buyBills" 
          :key="bill.id"
          :class="{ 'selected': selectedBill?.id === bill.id }"
          @click="selectBill(bill)"
        >
          <td>{{ bill.id }}</td>
          <td>{{ new Date(bill.date_buy).toLocaleDateString() }}</td>
          <td>{{ bill.supplier_name }}</td>
          <td>{{ bill.amount }}</td>
          <td>{{ bill.payed }}</td>
          <td>{{ bill.amount - bill.payed }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.master-section {
  width: 100%;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.data-table tbody tr {
  cursor: pointer;
  transition: background-color 0.2s;
}

.data-table tbody tr:hover {
  background-color: #f3f4f6;
}

.data-table tbody tr.selected {
  background-color: #e5e7eb;
}

.selected-id {
  background-color: #f0fdf4;
  color: #047857;
  padding: 8px 16px;
  border-radius: 4px;
  margin-bottom: 16px;
  border: 1px solid #10b981;
  font-weight: 500;
}
</style>