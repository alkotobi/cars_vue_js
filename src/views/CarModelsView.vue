<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, uploadFile, getFileUrl, error } = useApi()
const showAddBrandDialog = ref(false)
const showEditBrandDialog = ref(false)
const showAddCarNameDialog = ref(false)
const showEditCarNameDialog = ref(false)
const brands = ref([])
const carNames = ref([])
const editingBrand = ref(null)
const editingCarName = ref(null)
const user = ref(null)
const showMediaDialog = ref(false)
const selectedCarNameForMedia = ref(null)
const carNameMedia = ref([])
const uploadingMedia = ref(false)
const selectedMediaFiles = ref([])

const isAdmin = computed(() => user.value?.role_id === 1)

const newBrand = ref({
  brand: '',
  logoFile: null
})

const editingBrandLogoFile = ref(null)

const newCarName = ref({
  car_name: '',
  notes: '',
  is_big_car: false,
  id_brand: null
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
    query: `
      SELECT cn.*, b.brand 
      FROM cars_names cn
      LEFT JOIN brands b ON cn.id_brand = b.id 
      ORDER BY cn.car_name ASC
    `,
    params: []
  })
  if (result.success) {
    carNames.value = result.data
  }
}

const openMediaDialog = async (carName) => {
  selectedCarNameForMedia.value = carName
  showMediaDialog.value = true
  await fetchCarNameMedia(carName.id)
}

const fetchCarNameMedia = async (carNameId) => {
  const result = await callApi({
    query: `
      SELECT m.*, u.username as uploaded_by_username
      FROM car_name_media m
      LEFT JOIN users u ON m.uploaded_by = u.id
      WHERE m.car_name_id = ? AND m.is_active = 1
      ORDER BY m.uploaded_at DESC
    `,
    params: [carNameId]
  })
  if (result.success) {
    carNameMedia.value = result.data
  }
}

const handleMediaFileSelect = (event) => {
  const files = Array.from(event.target.files || [])
  selectedMediaFiles.value = files
}

const uploadCarNameMedia = async () => {
  if (!selectedMediaFiles.value.length || !selectedCarNameForMedia.value) return
  if (!user.value || !user.value.id) {
    alert('User not logged in')
    return
  }
  
  uploadingMedia.value = true
  const errors = []
  const successCount = []
  
  try {
    for (const file of selectedMediaFiles.value) {
      try {
        // Determine media type
        const mediaType = file.type.startsWith('image/') ? 'photo' : 'video'
        
        // Generate filename
        const fileExtension = file.name.split('.').pop().toLowerCase()
        const sanitizedCarName = selectedCarNameForMedia.value.car_name.toLowerCase().replace(/\s+/g, '_').replace(/[^a-z0-9_]/g, '')
        const filename = `${sanitizedCarName}_${mediaType}_${Date.now()}_${Math.random().toString(36).substring(7)}.${fileExtension}`
        
        // Upload file
        const uploadResult = await uploadFile(file, `car_names/${selectedCarNameForMedia.value.id}`, filename)
        
        if (uploadResult.success) {
          // Insert media record
          const insertResult = await callApi({
            query: `
              INSERT INTO car_name_media (car_name_id, file_path, file_name, file_size, file_type, media_type, uploaded_by)
              VALUES (?, ?, ?, ?, ?, ?, ?)
            `,
            params: [
              selectedCarNameForMedia.value.id,
              uploadResult.relativePath,
              file.name,
              file.size,
              file.type,
              mediaType,
              user.value.id
            ]
          })
          
          if (insertResult.success) {
            successCount.push(file.name)
          } else {
            errors.push(`${file.name}: ${insertResult.error}`)
          }
        } else {
          errors.push(`${file.name}: Upload failed`)
        }
      } catch (err) {
        errors.push(`${file.name}: ${err.message}`)
      }
    }
    
    // Refresh media list
    await fetchCarNameMedia(selectedCarNameForMedia.value.id)
    
    // Clear selected files
    selectedMediaFiles.value = []
    const fileInput = document.getElementById('media-file-input')
    if (fileInput) fileInput.value = ''
    
    // Show result message
    if (successCount.length > 0 && errors.length === 0) {
      alert(`Successfully uploaded ${successCount.length} file(s)`)
    } else if (successCount.length > 0) {
      alert(`Uploaded ${successCount.length} file(s), ${errors.length} failed: ${errors.join(', ')}`)
    } else {
      alert(`Upload failed: ${errors.join(', ')}`)
    }
  } finally {
    uploadingMedia.value = false
  }
}

const deleteCarNameMedia = async (media) => {
  if (!confirm(`Are you sure you want to delete ${media.file_name}?`)) return
  
  const result = await callApi({
    query: 'UPDATE car_name_media SET is_active = 0 WHERE id = ?',
    params: [media.id]
  })
  
  if (result.success) {
    await fetchCarNameMedia(selectedCarNameForMedia.value.id)
  } else {
    alert('Failed to delete media: ' + result.error)
  }
}

const closeMediaDialog = () => {
  showMediaDialog.value = false
  selectedCarNameForMedia.value = null
  carNameMedia.value = []
  selectedMediaFiles.value = []
}

const addBrand = async () => {
  let logoPath = null
  
  // Upload logo if provided
  if (newBrand.value.logoFile) {
    try {
      const fileExtension = newBrand.value.logoFile.name.split('.').pop().toLowerCase()
      const filename = `brand_${newBrand.value.brand.toLowerCase().replace(/\s+/g, '_')}_${Date.now()}.${fileExtension}`
      const uploadResult = await uploadFile(newBrand.value.logoFile, 'brands', filename)
      if (uploadResult.success) {
        logoPath = uploadResult.relativePath
      }
    } catch (err) {
      console.error('Error uploading logo:', err)
      alert('Failed to upload logo: ' + err.message)
      return
    }
  }
  
  // Build INSERT query - try with logo_path first, fallback if column doesn't exist
  let query = 'INSERT INTO brands (brand, logo_path) VALUES (?, ?)'
  let params = [newBrand.value.brand, logoPath]
  
  const result = await callApi({ query, params })
  
  // If the insert failed due to missing logo_path column, try without it
  if (!result.success && result.error && result.error.includes('logo_path')) {
    query = 'INSERT INTO brands (brand) VALUES (?)'
    params = [newBrand.value.brand]
    const retryResult = await callApi({ query, params })
    if (retryResult.success) {
      // If logo was uploaded but column doesn't exist, warn user
      if (logoPath) {
        console.warn('Logo uploaded but logo_path column does not exist in database. Please run migration 012_add_brand_logos.sql')
        alert('Brand added successfully, but logo could not be saved. Please run migration 012_add_brand_logos.sql to enable logo support.')
      }
      showAddBrandDialog.value = false
      newBrand.value = { brand: '', logoFile: null }
      await fetchBrands()
      return
    } else {
      // Both attempts failed
      console.log(retryResult.error)
      alert('Failed to add brand: ' + (retryResult.error || 'Unknown error'))
      return
    }
  }
  
  if (result.success) {
    showAddBrandDialog.value = false
    newBrand.value = { brand: '', logoFile: null }
    await fetchBrands()
  } else {
    console.log(result.error)
    alert('Failed to add brand: ' + (result.error || 'Unknown error'))
  }
}

const addCarName = async () => {
  const result = await callApi({
    query: 'INSERT INTO cars_names (car_name, notes, is_big_car, id_brand) VALUES (?, ?, ?, ?)',
    params: [
      newCarName.value.car_name,
      newCarName.value.notes,
      newCarName.value.is_big_car ? 1 : 0,
      newCarName.value.id_brand
    ]
  })
  if (result.success) {
    showAddCarNameDialog.value = false
    newCarName.value = { car_name: '', notes: '', is_big_car: false, id_brand: null }
    await fetchCarNames()
  }
  else{
    console.log(result.error)
  }
}

const editBrand = (brand) => {
  editingBrand.value = { ...brand }
  editingBrandLogoFile.value = null
  showEditBrandDialog.value = true
}

const editCarName = (carName) => {
  editingCarName.value = { ...carName }
  showEditCarNameDialog.value = true
}

const updateBrand = async () => {
  let logoPath = editingBrand.value.logo_path
  
  // Upload new logo if provided
  if (editingBrandLogoFile.value) {
    try {
      const fileExtension = editingBrandLogoFile.value.name.split('.').pop().toLowerCase()
      const filename = `brand_${editingBrand.value.brand.toLowerCase().replace(/\s+/g, '_')}_${Date.now()}.${fileExtension}`
      const uploadResult = await uploadFile(editingBrandLogoFile.value, 'brands', filename)
      if (uploadResult.success) {
        logoPath = uploadResult.relativePath
      }
    } catch (err) {
      console.error('Error uploading logo:', err)
      alert('Failed to upload logo: ' + err.message)
      return
    }
  }
  
  // Build UPDATE query - try with logo_path first, fallback if column doesn't exist
  let query = 'UPDATE brands SET brand = ?, logo_path = ? WHERE id = ?'
  let params = [editingBrand.value.brand, logoPath, editingBrand.value.id]
  
  const result = await callApi({ query, params })
  
  // If the update failed due to missing logo_path column, try without it
  if (!result.success && result.error && result.error.includes('logo_path')) {
    query = 'UPDATE brands SET brand = ? WHERE id = ?'
    params = [editingBrand.value.brand, editingBrand.value.id]
    const retryResult = await callApi({ query, params })
    if (retryResult.success) {
      // If logo was uploaded but column doesn't exist, warn user
      if (logoPath && editingBrandLogoFile.value) {
        console.warn('Logo uploaded but logo_path column does not exist in database. Please run migration 012_add_brand_logos.sql')
        alert('Brand updated successfully, but logo could not be saved. Please run migration 012_add_brand_logos.sql to enable logo support.')
      }
      showEditBrandDialog.value = false
      editingBrand.value = null
      editingBrandLogoFile.value = null
      await fetchBrands()
      return
    } else {
      // Both attempts failed
      console.log(retryResult.error)
      alert('Failed to update brand: ' + (retryResult.error || 'Unknown error'))
      return
    }
  }
  
  if (result.success) {
    showEditBrandDialog.value = false
    editingBrand.value = null
    editingBrandLogoFile.value = null
    await fetchBrands()
  } else {
    console.log(result.error)
  }
}

const updateCarName = async () => {
  const result = await callApi({
    query: 'UPDATE cars_names SET car_name = ?, notes = ?, is_big_car = ?, id_brand = ? WHERE id = ?',
    params: [
      editingCarName.value.car_name,
      editingCarName.value.notes,
      editingCarName.value.is_big_car ? 1 : 0,
      editingCarName.value.id_brand,
      editingCarName.value.id
    ]
  })
  if (result.success) {
    showEditCarNameDialog.value = false
    editingCarName.value = null
    await fetchCarNames()
  }
  else{
    console.log(result.error)
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

const formatFileSize = (bytes) => {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString()
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
            <th>Logo</th>
            <th>Brand</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="brand in brands" :key="brand.id">
            <td>
              <img 
                v-if="brand.logo_path" 
                :src="getFileUrl(brand.logo_path)" 
                :alt="brand.brand"
                class="brand-logo"
              />
              <span v-else class="no-logo">No logo</span>
            </td>
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
            <th>Brand</th>
            <th>Notes</th>
            <th>Big Car</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="carName in carNames" :key="carName.id">
            <td>{{ carName.car_name }}</td>
            <td>{{ carName.brand }}</td>
            <td>{{ carName.notes }}</td>
            <td>{{ carName.is_big_car ? 'Yes' : 'No' }}</td>
            <td>
              <button @click="editCarName(carName)" class="btn edit-btn">Edit</button>
              <button @click="openMediaDialog(carName)" class="btn media-btn">Photos/Videos</button>
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
          <label>Brand Name</label>
          <input 
            v-model="newBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label>Brand Logo</label>
          <input 
            type="file"
            accept="image/*"
            @change="(e) => newBrand.logoFile = e.target.files?.[0] || null"
            class="input-field"
          />
          <small class="form-hint">Upload a logo image (PNG, JPG, etc.)</small>
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
          <label>Brand Name</label>
          <input 
            v-model="editingBrand.brand" 
            placeholder="Brand Name" 
            class="input-field"
          />
        </div>
        <div class="form-group">
          <label>Current Logo</label>
          <div v-if="editingBrand.logo_path" class="logo-preview">
            <img 
              :src="getFileUrl(editingBrand.logo_path)" 
              :alt="editingBrand.brand"
              class="brand-logo-preview"
            />
          </div>
          <span v-else class="no-logo">No logo uploaded</span>
        </div>
        <div class="form-group">
          <label>Upload New Logo (optional)</label>
          <input 
            type="file"
            accept="image/*"
            @change="(e) => editingBrandLogoFile = e.target.files?.[0] || null"
            class="input-field"
          />
          <small class="form-hint">Upload a new logo to replace the current one</small>
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
          >
        </div>
        <div class="form-group">
          <select v-model="newCarName.id_brand">
            <option value="">Select Brand</option>
            <option v-for="brand in brands" :key="brand.id" :value="brand.id">
              {{ brand.brand }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <input 
            v-model="newCarName.notes" 
            placeholder="Notes"
          >
        </div>
        <div class="form-group">
          <label>
            <input 
              type="checkbox" 
              v-model="newCarName.is_big_car"
            > Big Car
          </label>
        </div>
        <div class="dialog-buttons">
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
          >
        </div>
        <div class="form-group">
          <select v-model="editingCarName.id_brand">
            <option value="">Select Brand</option>
            <option v-for="brand in brands" :key="brand.id" :value="brand.id">
              {{ brand.brand }}
            </option>
          </select>
        </div>
        <div class="form-group">
          <input 
            v-model="editingCarName.notes" 
            placeholder="Notes"
          >
        </div>
        <div class="form-group">
          <label>
            <input 
              type="checkbox" 
              v-model="editingCarName.is_big_car"
            > Big Car
          </label>
        </div>
        <div class="dialog-buttons">
          <button @click="updateCarName" class="btn save-btn">Save</button>
          <button @click="showEditCarNameDialog = false" class="btn cancel-btn">Cancel</button>
        </div>
      </div>
    </div>

    <!-- Media Dialog -->
    <div v-if="showMediaDialog" class="dialog-overlay" @click.self="closeMediaDialog">
      <div class="dialog media-dialog">
        <div class="dialog-header">
          <h3>Photos & Videos - {{ selectedCarNameForMedia?.car_name }}</h3>
          <button @click="closeMediaDialog" class="close-btn">&times;</button>
        </div>
        
        <div class="media-upload-section">
          <div class="form-group">
            <label>Upload Photos/Videos</label>
            <input
              id="media-file-input"
              type="file"
              accept="image/*,video/*"
              multiple
              @change="handleMediaFileSelect"
              class="input-field"
              :disabled="uploadingMedia"
            />
            <small class="form-hint">You can select multiple files at once</small>
          </div>
          <button 
            @click="uploadCarNameMedia" 
            class="btn save-btn"
            :disabled="!selectedMediaFiles.length || uploadingMedia"
          >
            <span v-if="uploadingMedia">Uploading...</span>
            <span v-else>Upload {{ selectedMediaFiles.length }} file(s)</span>
          </button>
        </div>

        <div class="media-gallery">
          <h4>Media Gallery ({{ carNameMedia.length }})</h4>
          <div v-if="carNameMedia.length === 0" class="no-media">
            No photos or videos uploaded yet
          </div>
          <div v-else class="media-grid">
            <div v-for="media in carNameMedia" :key="media.id" class="media-item">
              <div v-if="media.media_type === 'photo'" class="media-photo">
                <img :src="getFileUrl(media.file_path)" :alt="media.file_name" />
              </div>
              <div v-else class="media-video">
                <video controls :src="getFileUrl(media.file_path)"></video>
              </div>
              <div class="media-info">
                <div class="media-name">{{ media.file_name }}</div>
                <div class="media-meta">
                  <span>{{ media.media_type }}</span>
                  <span v-if="media.file_size">{{ formatFileSize(media.file_size) }}</span>
                  <span>{{ formatDate(media.uploaded_at) }}</span>
                </div>
              </div>
              <button @click="deleteCarNameMedia(media)" class="btn delete-btn-small" title="Delete">
                <i class="fas fa-trash"></i>
              </button>
            </div>
          </div>
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

.brand-logo {
  width: 40px;
  height: 40px;
  object-fit: contain;
  border-radius: 4px;
}

.brand-logo-preview {
  width: 100px;
  height: 100px;
  object-fit: contain;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 8px;
  background: #f9f9f9;
}

.logo-preview {
  margin-bottom: 12px;
}

.no-logo {
  color: #9ca3af;
  font-style: italic;
  font-size: 0.875rem;
}

.form-group label {
  font-weight: 500;
  margin-bottom: 4px;
  display: block;
  color: #374151;
}

.form-hint {
  display: block;
  color: #6b7280;
  font-size: 0.75rem;
  margin-top: 4px;
}

.media-btn {
  background-color: #8b5cf6;
  color: white;
}

.media-dialog {
  max-width: 900px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  color: #6b7280;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.media-upload-section {
  margin-bottom: 30px;
  padding: 20px;
  background-color: #f9fafb;
  border-radius: 8px;
}

.media-gallery h4 {
  margin-bottom: 16px;
  color: #374151;
}

.no-media {
  text-align: center;
  padding: 40px;
  color: #9ca3af;
  font-style: italic;
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.media-item {
  position: relative;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: white;
}

.media-photo img,
.media-video video {
  width: 100%;
  height: 200px;
  object-fit: cover;
  display: block;
}

.media-info {
  padding: 12px;
}

.media-name {
  font-weight: 500;
  font-size: 0.875rem;
  color: #374151;
  margin-bottom: 8px;
  word-break: break-word;
}

.media-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  font-size: 0.75rem;
  color: #6b7280;
}

.media-meta span {
  padding: 2px 8px;
  background-color: #f3f4f6;
  border-radius: 4px;
  text-transform: capitalize;
}

.delete-btn-small {
  position: absolute;
  top: 8px;
  right: 8px;
  padding: 6px 10px;
  background-color: rgba(239, 68, 68, 0.9);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
}

.delete-btn-small:hover {
  background-color: #ef4444;
}

</style>