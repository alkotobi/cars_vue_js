<template>
  <div class="car-names-view">
    <CarNamesToolbar
      :brands-count="brands.length"
      :car-names-count="carNames.length"
      :brands-open="showBrandsModal"
      @add-car="openAddCarName"
      @toggle-brands="toggleBrandsModal"
    />
    <CarNamesFilter
      v-model="filterQuery"
      :result-count="visibleCarNames.length"
      :total-count="carNames.length"
      :brand-id="brandPinId"
      :brand-name="brandPinName"
      @clear-pin="clearBrandPin"
    />
    <CarNamesTable
      :car-names="visibleCarNames"
      :loading="loadingCarNames"
      :error="carNamesError"
      :is-admin="isAdmin"
      :deleting="isBusy('delete-car-name')"
      :total-count="carNames.length"
      @edit="openEditCarName"
      @delete="deleteCarName"
      @open-media="openMedia"
      @retry="fetchCarNames"
    />
    <BrandsModal
      v-if="showBrandsModal"
      :brands="brands"
      :loading="loadingBrands"
      :error="brandsError"
      :is-admin="isAdmin"
      :deleting="isBusy('delete-brand')"
      @add="openAddBrand"
      @edit="openEditBrand"
      @delete="deleteBrand"
      @select="pinBrand"
      @close="closeBrandsModal"
      @retry="fetchBrands"
    />
    <BrandFormDialog
      v-if="showBrandForm"
      :brand="editingBrand"
      :saving="isBusy('add-brand') || isBusy('update-brand')"
      @save="handleBrandSave"
      @close="closeBrandForm"
    />
    <CarNameFormDialog
      v-if="showCarNameForm"
      :car-name="editingCarName"
      :brands="brands"
      :saving="isBusy('add-car-name') || isBusy('update-car-name')"
      @save="handleCarNameSave"
      @close="closeCarNameForm"
    />
    <CarNameMediaDialog
      v-if="mediaCarName"
      :car-name="mediaCarName"
      @close="closeMedia"
    />
    <MessageBox v-bind="msgBox" @confirm="onConfirm" @cancel="onCancel" @close="onClose" />
  </div>
</template>
<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import CarNamesToolbar from '../components/car-names/CarNamesToolbar.vue'
import CarNamesFilter from '../components/car-names/CarNamesFilter.vue'
import CarNamesTable from '../components/car-names/CarNamesTable.vue'
import BrandsModal from '../components/car-names/BrandsModal.vue'
import BrandFormDialog from '../components/car-names/BrandFormDialog.vue'
import CarNameFormDialog from '../components/car-names/CarNameFormDialog.vue'
import CarNameMediaDialog from '../components/car-names/CarNameMediaDialog.vue'
import MessageBox from '../components/MessageBox.vue'
import { useNotify } from '../composables/useNotify'
import { useSafeApi } from '../composables/useSafeApi'
import { useBrands } from '../composables/useBrands'
import { useCarNames } from '../composables/useCarNames'
import { filterCarNames } from '../lib/carNamesFilter'
import { useSubmitGuard } from '../composables/useSubmitGuard'

const { msgBox, notify, notifyError, onConfirm, onCancel, onClose } = useNotify()
const { safeApi, uploadFile, getFileUrl } = useSafeApi()
const { guard, isBusy } = useSubmitGuard()
const brandsApiFactory = useBrands({ notify, onBrandsChanged: fetchCarNames })
const brandsApi = brandsApiFactory({ safeApi, uploadFile, getFileUrl })
const brands = brandsApi.brands
const loadingBrands = brandsApi.loading
const brandsError = brandsApi.error
const fetchBrands = brandsApi.fetchBrands
const carNamesApiFactory = useCarNames({ notify })
const carNamesApi = carNamesApiFactory({ safeApi })
const carNames = carNamesApi.carNames
const loadingCarNames = carNamesApi.loading
const carNamesError = carNamesApi.error
const fetchCarNames = carNamesApi.fetchCarNames

async function fetchBoth() {
  await Promise.all([fetchBrands(), fetchCarNames()])
}

const showBrandsModal = ref(false)
const showBrandForm = ref(false)
const showCarNameForm = ref(false)
const editingBrand = ref(null)
const editingCarName = ref(null)
const mediaCarName = ref(null)
const filterQuery = ref('')
const brandPinId = ref(null)
const brandPinName = ref('')

const user = ref(null)
const isAdmin = computed(() => user.value?.role_id === 1)

const visibleCarNames = computed(() =>
  filterCarNames(carNames.value, filterQuery.value, { brandId: brandPinId.value })
)

function toggleBrandsModal() {
  showBrandsModal.value = !showBrandsModal.value
}
function closeBrandsModal() {
  showBrandsModal.value = false
}
function openAddBrand() {
  editingBrand.value = null
  showBrandForm.value = true
}
function openEditBrand(b) {
  editingBrand.value = { ...b }
  showBrandForm.value = true
}
function closeBrandForm() {
  showBrandForm.value = false
  editingBrand.value = null
}
function openAddCarName() {
  editingCarName.value = null
  showCarNameForm.value = true
}
function openEditCarName(cn) {
  editingCarName.value = { ...cn }
  showCarNameForm.value = true
}
function closeCarNameForm() {
  showCarNameForm.value = false
  editingCarName.value = null
}
function openMedia(cn) {
  mediaCarName.value = { ...cn }
}
function closeMedia() {
  mediaCarName.value = null
}
function pinBrand(b) {
  brandPinId.value = b.id
  brandPinName.value = b.brand || ''
  filterQuery.value = ''
}
function clearBrandPin() {
  brandPinId.value = null
  brandPinName.value = ''
}

const addBrand = guard('add-brand', async (p) => {
  const r = await brandsApi.addBrand(p)
  if (r.success) closeBrandForm()
  return r
})
const updateBrand = guard('update-brand', async (p) => {
  const r = await brandsApi.updateBrand({ ...p, editingBrandLogoFile: p.logoFile })
  if (r.success) closeBrandForm()
  return r
})
const deleteBrand = guard('delete-brand', async (b) => {
  await brandsApi.deleteBrand(b)
})
const addCarName = guard('add-car-name', async (p) => {
  const r = await carNamesApi.addCarName(p)
  if (r.success) closeCarNameForm()
  return r
})
const updateCarName = guard('update-car-name', async (p) => {
  const r = await carNamesApi.updateCarName(p)
  if (r.success) closeCarNameForm()
  return r
})
const deleteCarName = guard('delete-car-name', async (cn) => {
  await carNamesApi.deleteCarName(cn)
})

function handleBrandSave(p) {
  if (editingBrand.value && editingBrand.value.id) {
    updateBrand({ ...p, id: editingBrand.value.id })
  } else {
    addBrand(p)
  }
}
function handleCarNameSave(p) {
  if (editingCarName.value && editingCarName.value.id) {
    updateCarName({ ...p, id: editingCarName.value.id })
  } else {
    addCarName(p)
  }
}

function onKeydown(e) {
  if (msgBox.value.show) return
  if (mediaCarName.value) return
  if (showCarNameForm.value) return
  if (showBrandForm.value) return
  if (showBrandsModal.value) {
    if (e.key === 'Escape') {
      e.preventDefault()
      closeBrandsModal()
    }
    return
  }
  if (e.key === 'Escape') {
    if (brandPinId.value) {
      e.preventDefault()
      clearBrandPin()
    }
  }
}

onMounted(async () => {
  window.addEventListener('keydown', onKeydown)
  const userStr = localStorage.getItem('user')
  if (!userStr) return
  try {
    user.value = JSON.parse(userStr)
  } catch (err) {
    await notifyError('Session problem', 'Saved login could not be read.', 'Log out and log in again.')
    return
  }
  await fetchBoth()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeydown)
})
</script>
<style scoped>
.car-names-view {
  padding: 20px;
}
</style>
