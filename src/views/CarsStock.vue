<script setup>
import { ref } from 'vue'

const cars = ref([])
</script>

<template>
  <div class="cars-stock">
    <h2>Cars Stock Management</h2>
    <div class="content">
      <!-- Cars stock table will go here -->
      <table class="cars-table">
        <thead>
          <tr>
            <th>Car Name</th>
            <th>Brand</th>
            <th>Color</th>
            <th>VIN</th>
            <th>Price Buy</th>
            <th>Price Sell</th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <!-- Car rows will go here -->
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.cars-stock {
  padding: 20px;
}

.content {
  margin-top: 20px;
}

.cars-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.cars-table th,
.cars-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.cars-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.cars-table tbody tr:hover {
  background-color: #f5f5f5;
}
</style>