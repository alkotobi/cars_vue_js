<template>
  <div class="task-form-example">
    <h2>Task Form Examples</h2>

    <div class="example-buttons">
      <button @click="openCarTask" class="btn-example">
        <i class="fas fa-car"></i>
        Create Car Task
      </button>

      <button @click="openClientTask" class="btn-example">
        <i class="fas fa-user"></i>
        Create Client Task
      </button>

      <button @click="openSupplierTask" class="btn-example">
        <i class="fas fa-truck"></i>
        Create Supplier Task
      </button>

      <button @click="openSellTask" class="btn-example">
        <i class="fas fa-handshake"></i>
        Create Sell Task
      </button>

      <button @click="openBuyTask" class="btn-example">
        <i class="fas fa-shopping-cart"></i>
        Create Buy Task
      </button>

      <button @click="openGeneralTask" class="btn-example">
        <i class="fas fa-tasks"></i>
        Create General Task
      </button>
    </div>

    <!-- General Purpose Task Form -->
    <TaskForm
      :entity-data="currentEntityData"
      :entity-type="currentEntityType"
      :is-visible="showTaskForm"
      @save="handleTaskSave"
      @cancel="handleTaskCancel"
    />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import TaskForm from './car-stock/TaskForm.vue'

const showTaskForm = ref(false)
const currentEntityType = ref('general')
const currentEntityData = ref({})

// Example data for different entity types
const exampleData = {
  car: {
    id: 123,
    car_name: 'Toyota Camry 2023',
  },
  client: {
    id: 456,
    name: 'John Doe',
  },
  supplier: {
    id: 789,
    name: 'ABC Motors',
  },
  sell: {
    id: 101,
    bill_number: 'SELL-2024-001',
  },
  buy: {
    id: 202,
    bill_number: 'BUY-2024-001',
  },
  general: {},
}

const openCarTask = () => {
  currentEntityType.value = 'car'
  currentEntityData.value = exampleData.car
  showTaskForm.value = true
}

const openClientTask = () => {
  currentEntityType.value = 'client'
  currentEntityData.value = exampleData.client
  showTaskForm.value = true
}

const openSupplierTask = () => {
  currentEntityType.value = 'supplier'
  currentEntityData.value = exampleData.supplier
  showTaskForm.value = true
}

const openSellTask = () => {
  currentEntityType.value = 'sell'
  currentEntityData.value = exampleData.sell
  showTaskForm.value = true
}

const openBuyTask = () => {
  currentEntityType.value = 'buy'
  currentEntityData.value = exampleData.buy
  showTaskForm.value = true
}

const openGeneralTask = () => {
  currentEntityType.value = 'general'
  currentEntityData.value = exampleData.general
  showTaskForm.value = true
}

const handleTaskSave = (result) => {
  console.log('Task saved:', result)
  showTaskForm.value = false
  // You can add success notification here
}

const handleTaskCancel = () => {
  showTaskForm.value = false
}
</script>

<style scoped>
.task-form-example {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}

.task-form-example h2 {
  text-align: center;
  margin-bottom: 30px;
  color: #1f2937;
}

.example-buttons {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 30px;
}

.btn-example {
  padding: 16px 20px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-size: 1rem;
  font-weight: 500;
}

.btn-example:hover {
  background-color: #45a049;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.btn-example i {
  font-size: 1.1rem;
}

@media (max-width: 640px) {
  .example-buttons {
    grid-template-columns: 1fr;
  }

  .btn-example {
    padding: 14px 16px;
    font-size: 0.9rem;
  }
}
</style>
