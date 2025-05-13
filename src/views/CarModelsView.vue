<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error } = useApi()
const showAddBrandDialog = ref(false)
const showEditBrandDialog = ref(false)
const showAddCarNameDialog = ref(false)
const showEditCarNameDialog = ref(false)
const brands = ref([])
const carNames = ref([])
const editingBrand = ref(null)
const editingCarName = ref(null)
const user = ref(null)

const isAdmin = computed(() => user.value?.role_id === 1)

const newBrand = ref({
  brand: ''
})

const newCarName = ref({
  car_name: '',
  notes: '',
  is_big_car: false
})

const fetchBrands = async () => {
  const result = await callApi({
    query: 'SELECT * FROM brands ORDER BY brand ASC',
    params: []
  })
  if (result.success) {
    brands.value = result.data
  }
}

const fetchCarNames = async () => {
  const result = await callApi({
    query: 'SELECT * FROM cars_names ORDER BY car_name ASC',
    params: []
  })
  if (result.success) {
    carNames.value = result.data
  }
}

const addBrand = async () => {
  const result = await callApi({
    query: 'INSERT INTO brands (brand) VALUES (?)',
    params: [newBrand.value.brand]
  })
  if (result.success) {
    showAddBrandDialog.value = false
    newBrand.value = { brand: '' }
    await fetchBrands()
  }
}

const addCarName = async () => {
  const result = await callApi({
    query: 'INSERT INTO cars_names (car_name, notes, is_big_car) VALUES (?, ?, ?)',
    params: [newCarName.value.car_name, newCarName.value.notes, newCarName.value.is_big_car ? 1 : 0]
  })
  if (result.success) {
    showAddCarNameDialog.value = false
    newCarName.value = { car_name: '', notes: '', is_big_car: false }
    await fetchCarNames()
  }
}

const editBrand = (brand) => {
  editingBrand.value = { ...brand }
  showEditBrandDialog.value = true
}

const editCarName = (carName) => {
  editingCarName.value = { ...carName }
  showEditCarNameDialog.value = true
}

const updateBrand = async () => {
  const result = await callApi({
    query: 'UPDATE brands SET brand = ? WHERE id = ?',
    params: [editingBrand.value.brand, editingBrand.value.id]
  })
  if (result.success) {
    showEditBrandDialog.value = false
    editingBrand.value = null
    await fetchBrands()
  }
}

const updateCarName = async () => {
  const result = await callApi({
    query: 'UPDATE cars_names SET car_name = ?, notes = ?, is_big_car = ? WHERE id = ?',
    params: [
      editingCarName.value.car_name,
      editingCarName.value.notes,
      editingCarName.value.is_big_car ? 1 : 0,
      editingCarName.value.id
    ]
  })
  if (result.success) {
    showEditCarNameDialog.value = false
    editingCarName.value = null
    await fetchCarNames()
  }
}

const deleteBrand = async (brand) => {
  if (confirm('Are you sure you want to delete this brand?')) {
    const result = await callApi({
      query: 'DELETE FROM brands WHERE id = ?',
      params: [brand.id]
    })
    if (result.success) {
      await fetchBrands()
    }
  }
}

const deleteCarName = async (carName) => {
  if (confirm('Are you sure you want to delete this car name?')) {
    const result = await callApi({
      query: 'DELETE FROM cars_names WHERE id = ?',
      params: [carName.id]
    })
    if (result.success) {
      await fetchCarNames()
    }
  }
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchBrands()
    fetchCarNames()
  }
})
</script>

<template>
  <div class="models-view">
    <div class="section">
      <div class="header">
        <h2>Brands Management</h2>
        <button @click="showAddBrandDialog = true" class="add-btn">Add Brand</button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Brand</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="brand in brands" :key="brand.id">
            <td>{{ brand.brand }}</td>
            <td>
              <button @click="editBrand(brand)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click="deleteBrand(brand)" 
                class="btn delete-btn"
              >Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="section">
      <div class="header">
        <h2>Car Names Management</h2>
        <button @click="showAddCarNameDialog = true" class="add-btn">Add Car Name</button>
      </div>
      <table class="data-table">
        <thead>
          <tr>
            <th>Car Name</th>
            <th>Notes</th>
            <th>Big Car</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="carName in carNames" :key="carName.id">
            <td>{{ carName.car_name }}</td>
            <td>{{ carName.notes }}</td>
            <td>{{ carName.is_big_car ? 'Yes' : 'No' }}</td>
            <td>
              <button @click="editCarName(carName)" class="btn edit-btn">Edit</button>
              <button 
                v-if="isAdmin"
                @click="deleteCarName(carName)" 
                class="btn delete-btn"
              >Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Brand Dialog -->
    <div v-if="showAddBrandDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Brand</h3>
        <div class="form-group">
          <input 
            v-model="newBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="addBrand" class="btn save-btn">Save</button>
          <button @click="showAddBrandDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Brand Dialog -->
    <div v-if="showEditBrandDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Brand</h3>
        <div class="form-group">
          <input 
            v-model="editingBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="dialog-actions">
          <button @click="updateBrand" class="btn save-btn">Save</button>
          <button @click="showEditBrandDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Add Car Name Dialog -->
    <div v-if="showAddCarNameDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Add New Car Name</h3>
        <div class="form-group">
          <input 
            v-model="newCarName.car_name" 
            placeholder="Car Name" 
            class="input-field"
          />
          <textarea 
            v-model="newCarName.notes" 
            placeholder="Notes" 
            class="input-field textarea"
          ></textarea>
          <div class="checkbox-field">
            <input 
              type="checkbox" 
              id="newIsBigCar" 
              v-model="newCarName.is_big_car"
            />
            <label for="newIsBigCar">Is Big Car</label>
          </div>
        </div>
        <div class="dialog-actions">
          <button @click="addCarName" class="btn save-btn">Save</button>
          <button @click="showAddCarNameDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Edit Car Name Dialog -->
    <div v-if="showEditCarNameDialog" class="dialog-overlay">
      <div class="dialog">
        <h3>Edit Car Name</h3>
        <div class="form-group">
          <input 
            v-model="editingCarName.car_name" 
            placeholder="Car Name" 
            class="input-field"
          />
          <textarea 
            v-model="editingCarName.notes" 
            placeholder="Notes" 
            class="input-field textarea"
          ></textarea>
          <div class="checkbox-field">
            <input 
              type="checkbox" 
              id="editIsBigCar" 
              v-model="editingCarName.is_big_car"
            />
            <label for="editIsBigCar">Is Big Car</label>
          </div>
        </div>
        <div class="dialog-actions">
          <button @click="updateCarName" class="btn save-btn">Save</button>
          <button @click="showEditCarNameDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.models-view {
  padding: 20px;
}

.section {
  margin-bottom: 40px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.data-table th {
  background-color: #f8f9fa;
  font-weight: 600;
}

.data-table tbody tr:hover {
  background-color: #f5f5f5;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.dialog {
  background-color: white;
  padding: 20px;
  border-radius: 8px;
  min-width: 400px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 20px;
}

.input-field {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.textarea {
  min-height: 100px;
  resize: vertical;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.checkbox-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.checkbox-field input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.checkbox-field label {
  cursor: pointer;
}
</style>