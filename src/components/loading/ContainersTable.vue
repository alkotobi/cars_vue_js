<template>
  <div class="containers-table-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-box"></i>
        {{ t('loading.containers') }}
        <span v-if="hasActiveFilters" class="filter-indicator">
          <i class="fas fa-filter"></i>
          {{ t('loading.filtered') }}
        </span>
      </h3>
      <div class="header-actions" v-if="selectedLoadingId">
        <button
          @click="openAddDialog"
          class="add-btn"
          :disabled="loading || !props.selectedLoadingId"
          :title="t('loading.add_new_container')"
        >
          <i class="fas fa-plus"></i>
          {{ t('loading.add_container') }}
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
          :title="t('loading.refresh_containers')"
        >
          <i class="fas fa-sync-alt"></i>
          <span>{{ t('loading.refresh') }}</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <div v-if="!selectedLoadingId" class="empty-state">
      <i class="fas fa-mouse-pointer"></i>
      <p>{{ t('loading.click_loading_record_to_view_containers') }}</p>
    </div>

    <div v-else class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ t('loading.loading') }}</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="!props.selectedLoadingId" class="empty-state">
        <i class="fas fa-mouse-pointer fa-2x"></i>
        <p>{{ t('loading.please_select_loading_record') }}</p>
      </div>

      <div v-else-if="containers.length === 0" class="empty-state">
        <i class="fas fa-box-open fa-2x"></i>
        <p>{{ t('loading.no_containers_found_for_loading_record') }}</p>
        <button @click="openAddDialog" class="add-first-btn">
          <i class="fas fa-plus"></i>
          {{ t('loading.add_first_container') }}
        </button>
      </div>

      <div v-else class="table-content">
        <table class="containers-table">
          <thead>
            <tr>
              <th @click="sortByColumn('id')" class="sortable-header">
                {{ t('loading.id') }}
                <i
                  v-if="sortBy === 'id'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('container_name')" class="sortable-header">
                {{ t('loading.container') }}
                <i
                  v-if="sortBy === 'container_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('ref_container')" class="sortable-header">
                {{ t('loading.reference') }}
                <i
                  v-if="sortBy === 'ref_container'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('so')" class="sortable-header">
                {{ t('loading.so') }}
                <i
                  v-if="sortBy === 'so'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('date_departed')" class="sortable-header">
                {{ t('loading.date_departed') }}
                <i
                  v-if="sortBy === 'date_departed'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('date_loaded')" class="sortable-header">
                {{ t('loading.date_loaded') }}
                <i
                  v-if="sortBy === 'date_loaded'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('date_on_board')" class="sortable-header">
                {{ t('loading.date_on_board') }}
                <i
                  v-if="sortBy === 'date_on_board'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('is_released')" class="sortable-header">
                {{ t('loading.released') }}
                <i
                  v-if="sortBy === 'is_released'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('note')" class="sortable-header">
                {{ t('loading.notes') }}
                <i
                  v-if="sortBy === 'note'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th>{{ t('loading.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="container in containers"
              :key="container.id"
              class="table-row"
              :class="{ 'selected-row': props.selectedContainerId === container.id }"
              @click="handleContainerClick(container)"
            >
              <td class="id-cell">#{{ container.id }}</td>
              <td class="container-cell">{{ container.container_name || '-' }}</td>
              <td class="ref-cell">
                <div class="ref-content">
                  <span v-if="container.ref_container && container.ref_container.trim()">
                    {{ container.ref_container }}
                  </span>
                  <span v-else class="no-ref-error">
                    <i class="fas fa-exclamation-triangle"></i>
                    {{ t('loading.no_reference') }}
                  </span>
                </div>
              </td>
              <td class="so-cell">{{ container.so || '-' }}</td>
              <td class="date-cell">{{ formatDate(container.date_departed) }}</td>
              <td class="date-cell">{{ formatDate(container.date_loaded) }}</td>
              <td class="date-cell">{{ formatDate(container.date_on_board) }}</td>
              <td class="released-cell">
                <span
                  :class="{
                    'status-released': container.is_released,
                    'status-not-released': !container.is_released,
                  }"
                >
                  <i :class="[container.is_released ? 'fas fa-check-circle' : 'fas fa-clock']"></i>
                  {{ container.is_released ? t('loading.released') : t('loading.not_released') }}
                </span>
              </td>
              <td class="note-cell">{{ truncateText(container.note, 30) }}</td>
              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click.stop="editContainer(container)"
                    class="action-btn edit-btn"
                    :title="t('loading.edit_container')"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    @click.stop="setOnBoard(container)"
                    class="action-btn onboard-btn"
                    :title="getOnBoardButtonTitle(container)"
                    :disabled="container.date_on_board || container.assigned_cars_count === 0"
                  >
                    <i class="fas fa-ship"></i>
                  </button>
                  <button
                    @click.stop="deleteContainer(container)"
                    class="action-btn delete-btn"
                    :title="t('loading.delete_container')"
                  >
                    <i class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Table Footer with Container Count -->
      <div class="table-footer" v-if="containers.length > 0">
        <span class="container-count">
          {{ t('loading.showing_containers', { count: containers.length }) }}
          <span v-if="hasActiveFilters" class="filtered-note">
            {{ t('loading.filtered_by_criteria') }}
          </span>
        </span>
      </div>
    </div>

    <!-- Add/Edit Dialog -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-box"></i>
            {{
              isEditing
                ? t('loading.edit_container_assignment')
                : t('loading.add_container_to_loading')
            }}
          </h3>
          <button class="close-btn" @click="closeDialog" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveContainer" class="dialog-content">
          <div class="form-row">
            <div class="form-group">
              <label for="id_container">
                <i class="fas fa-box"></i>
                {{ t('loading.container') }}
              </label>
              <div class="input-with-button">
                <select
                  id="id_container"
                  v-model="formData.id_container"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">{{ t('loading.select_container') }}</option>
                  <option
                    v-for="container in availableContainers"
                    :key="container.id"
                    :value="container.id"
                  >
                    {{ container.name }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddContainerDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  :title="t('loading.add_new_container')"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="ref_container">
                <i class="fas fa-tag"></i>
                {{ t('loading.reference') }}
              </label>
              <input
                type="text"
                id="ref_container"
                v-model="formData.ref_container"
                :disabled="isSubmitting"
                :placeholder="t('loading.container_reference')"
                maxlength="255"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="so">
                <i class="fas fa-file-invoice"></i>
                {{ t('loading.so') }}
              </label>
              <input
                type="text"
                id="so"
                v-model="formData.so"
                :disabled="isSubmitting"
                :placeholder="t('loading.so_number')"
                maxlength="255"
              />
            </div>

            <div class="form-group">
              <label for="date_departed">
                <i class="fas fa-calendar"></i>
                {{ t('loading.date_departed') }}
              </label>
              <input
                type="date"
                id="date_departed"
                v-model="formData.date_departed"
                :disabled="isSubmitting"
                min="1900-01-01"
                max="2100-12-31"
              />
            </div>
          </div>

          <div class="form-group">
            <label for="date_loaded">
              <i class="fas fa-calendar"></i>
              {{ t('loading.date_loaded') }}
            </label>
            <input
              type="date"
              id="date_loaded"
              v-model="formData.date_loaded"
              :disabled="isSubmitting"
              min="1900-01-01"
              max="2100-12-31"
            />
          </div>

          <div class="form-group">
            <label for="date_on_board">
              <i class="fas fa-calendar"></i>
              {{ t('loading.date_on_board') }}
            </label>
            <input
              type="date"
              id="date_on_board"
              v-model="formData.date_on_board"
              :disabled="isSubmitting"
              min="1900-01-01"
              max="2100-12-31"
            />
          </div>

          <div class="form-group">
            <label for="is_released">
              <i class="fas fa-check-circle"></i>
              {{ t('loading.released') }}
            </label>
            <input
              type="checkbox"
              id="is_released"
              v-model="formData.is_released"
              :disabled="isSubmitting"
            />
          </div>

          <div class="form-group">
            <label for="note">
              <i class="fas fa-sticky-note"></i>
              {{ t('loading.notes') }}
            </label>
            <textarea
              id="note"
              v-model="formData.note"
              :disabled="isSubmitting"
              :placeholder="t('loading.additional_notes')"
              rows="3"
            ></textarea>
          </div>

          <div class="form-actions">
            <button type="button" @click="closeDialog" class="cancel-btn" :disabled="isSubmitting">
              {{ t('loading.cancel') }}
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isSubmitting"
              :class="{ processing: isSubmitting }"
            >
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
              <span>{{ isEditing ? t('loading.update') : t('loading.add') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Add Container Dialog -->
    <div v-if="showAddContainerDialog" class="dialog-overlay" @click.self="closeAddContainerDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-box"></i>
            {{ t('loading.add_new_container') }}
          </h3>
          <button class="close-btn" @click="closeAddContainerDialog" :disabled="isAddingContainer">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveNewContainer" class="dialog-content">
          <div class="form-group">
            <label for="new_container_name">
              <i class="fas fa-tag"></i>
              {{ t('loading.container_name') }}
            </label>
            <input
              type="text"
              id="new_container_name"
              v-model="newContainerData.name"
              required
              :disabled="isAddingContainer"
              maxlength="30"
              :placeholder="t('loading.enter_container_name')"
            />
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="closeAddContainerDialog"
              class="cancel-btn"
              :disabled="isAddingContainer"
            >
              {{ t('loading.cancel') }}
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isAddingContainer"
              :class="{ processing: isAddingContainer }"
            >
              <i v-if="isAddingContainer" class="fas fa-spinner fa-spin"></i>
              <span>{{ t('loading.create') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <div v-if="showDeleteDialog" class="dialog-overlay" @click.self="closeDeleteDialog">
      <div class="dialog delete-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-exclamation-triangle"></i>
            {{ t('loading.confirm_delete') }}
          </h3>
          <button class="close-btn" @click="closeDeleteDialog" :disabled="isDeleting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="delete-message">
            <p>
              {{
                t('loading.confirm_delete_container_message', {
                  containerName: containerToDelete?.container_name,
                })
              }}
            </p>
          </div>

          <div class="form-actions">
            <button
              type="button"
              @click="closeDeleteDialog"
              class="cancel-btn"
              :disabled="isDeleting"
            >
              {{ t('loading.cancel') }}
            </button>
            <button
              type="button"
              @click="confirmDelete"
              class="delete-confirm-btn"
              :disabled="isDeleting"
              :class="{ processing: isDeleting }"
            >
              <i v-if="isDeleting" class="fas fa-spinner fa-spin"></i>
              <span>{{ t('loading.delete') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- On Board Dialog -->
    <div v-if="showOnBoardDialog" class="dialog-overlay" @click.self="closeOnBoardDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-ship"></i>
            {{ t('loading.set_container_on_board') }}
          </h3>
          <button class="close-btn" @click="closeOnBoardDialog" :disabled="isSubmittingOnBoard">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="onboard_date">
              <i class="fas fa-calendar"></i>
              {{ t('loading.date_on_board') }}
            </label>
            <input
              type="date"
              id="onboard_date"
              v-model="onBoardData.date_on_board"
              required
              :disabled="isSubmittingOnBoard"
              min="1900-01-01"
              max="2100-12-31"
            />
          </div>

          <div class="form-group">
            <label>
              <i class="fas fa-info-circle"></i>
              {{ t('loading.container_information') }}
            </label>
            <div class="info-display">
              <p>
                <strong>{{ t('loading.container') }}:</strong> {{ onBoardData.container_name }}
              </p>
              <p>
                <strong>{{ t('loading.reference') }}:</strong>
                {{ onBoardData.ref_container || t('loading.na') }}
              </p>
              <p>
                <strong>{{ t('loading.assigned_cars') }}:</strong>
                {{ onBoardData.assigned_cars_count || 0 }}
              </p>
            </div>
          </div>

          <div class="warning-message">
            <i class="fas fa-exclamation-triangle"></i>
            <p>
              <strong>{{ t('loading.warning') }}:</strong>{{ t('loading.warning_message') }}
            </p>
          </div>
        </div>

        <div class="dialog-footer">
          <button
            type="button"
            @click="closeOnBoardDialog"
            class="cancel-btn"
            :disabled="isSubmittingOnBoard"
          >
            {{ t('loading.cancel') }}
          </button>
          <button
            type="button"
            @click="confirmOnBoard"
            class="confirm-btn"
            :disabled="isSubmittingOnBoard"
            :class="{ processing: isSubmittingOnBoard }"
          >
            <i v-if="isSubmittingOnBoard" class="fas fa-spinner fa-spin"></i>
            <span>{{ t('loading.set_on_board') }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, defineExpose, onMounted, computed } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '@/composables/useApi'

const { t } = useEnhancedI18n()
const props = defineProps({
  selectedLoadingId: {
    type: [Number, String],
    default: null,
  },
  selectedContainerId: {
    type: [Number, String],
    default: null,
  },
  soldDateFrom: {
    type: String,
    default: '',
  },
  soldDateTo: {
    type: String,
    default: '',
  },
  // New car and client filters
  carNameFilter: {
    type: String,
    default: '',
  },
  vinFilter: {
    type: String,
    default: '',
  },
  clientNameFilter: {
    type: String,
    default: '',
  },
  clientIdFilter: {
    type: String,
    default: '',
  },
  // Container reference filter
  containerRefFilter: {
    type: String,
    default: '',
  },
})

const emit = defineEmits(['container-click', 'refresh-unassigned-cars', 'container-created'])

const { callApi } = useApi()

// Computed property to check if any filters are active
const hasActiveFilters = computed(() => {
  return !!(
    props.carNameFilter ||
    props.vinFilter ||
    props.clientNameFilter ||
    props.clientIdFilter ||
    props.soldDateFrom ||
    props.soldDateTo ||
    props.containerRefFilter
  )
})

const containers = ref([])
const availableContainers = ref([])
const loading = ref(false)
const error = ref(null)
const showDialog = ref(false)
const showAddContainerDialog = ref(false)
const showDeleteDialog = ref(false)
const showOnBoardDialog = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const isAddingContainer = ref(false)
const isDeleting = ref(false)
const isSubmittingOnBoard = ref(false)
const containerToDelete = ref(null)
const sortBy = ref('id')
const sortOrder = ref('desc')

const formData = ref({
  id_container: '',
  ref_container: '',
  so: '',
  date_departed: '',
  date_loaded: '',
  date_on_board: '',
  is_released: false,
  note: '',
})

const newContainerData = ref({
  name: '',
})

const onBoardData = ref({
  id: null,
  date_on_board: '',
  container_name: '',
  ref_container: '',
  assigned_cars_count: 0,
})

const fetchContainers = async () => {
  if (!props.selectedLoadingId) {
    containers.value = []
    return
  }

  console.log('Fetching containers for loading ID:', props.selectedLoadingId)

  loading.value = true
  error.value = null

  try {
    // Build the ORDER BY clause based on sorting
    let orderByClause = 'ORDER BY lc.id DESC' // Default sorting

    if (sortBy.value) {
      const sortField =
        sortBy.value === 'id'
          ? 'lc.id'
          : sortBy.value === 'container_name'
            ? 'c.name'
            : sortBy.value === 'ref_container'
              ? 'lc.ref_container'
              : sortBy.value === 'so'
                ? 'lc.so'
                : sortBy.value === 'date_departed'
                  ? 'lc.date_departed'
                  : sortBy.value === 'date_loaded'
                    ? 'lc.date_loaded'
                    : sortBy.value === 'date_on_board'
                      ? 'lc.date_on_board'
                      : sortBy.value === 'is_released'
                        ? 'lc.is_released'
                        : sortBy.value === 'note'
                          ? 'lc.note'
                          : 'lc.id'

      orderByClause = `ORDER BY ${sortField} ${sortOrder.value.toUpperCase()}`
    }

    let query = `
      SELECT 
        lc.id,
        lc.id_container,
        c.name as container_name,
        lc.ref_container,
        lc.so,
        lc.date_departed,
        lc.date_loaded,
        lc.date_on_board,
        lc.is_released,
        lc.note,
        COUNT(cs.id) as assigned_cars_count,
        GROUP_CONCAT(DISTINCT cn.car_name) as car_names,
        GROUP_CONCAT(DISTINCT cs.vin) as vins,
        GROUP_CONCAT(DISTINCT cl.name) as client_names,
        GROUP_CONCAT(DISTINCT cl.id_no) as client_ids
      FROM loaded_containers lc
      LEFT JOIN containers c ON lc.id_container = c.id
      LEFT JOIN cars_stock cs ON lc.id = cs.id_loaded_container
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN clients cl ON cs.id_client = cl.id
      WHERE lc.id_loading = ?
    `

    let params = [props.selectedLoadingId]

    // Add sold date filter if specified
    if (props.soldDateFrom || props.soldDateTo) {
      query += ` AND cs.id_sell IS NOT NULL`
      query += ` AND EXISTS (
        SELECT 1 FROM sell_bill sb 
        WHERE sb.id = cs.id_sell 
        AND sb.date_sell IS NOT NULL
      `

      if (props.soldDateFrom) {
        query += ` AND DATE(sb.date_sell) >= ?`
        params.push(props.soldDateFrom)
      }

      if (props.soldDateTo) {
        query += ` AND DATE(sb.date_sell) <= ?`
        params.push(props.soldDateTo)
      }

      query += ` )`
    }

    // Add car and client filters
    if (props.carNameFilter) {
      query += ` AND EXISTS (
        SELECT 1 FROM cars_stock cs2 
        LEFT JOIN buy_details bd2 ON cs2.id_buy_details = bd2.id
        LEFT JOIN cars_names cn2 ON bd2.id_car_name = cn2.id
        WHERE cs2.id_loaded_container = lc.id 
        AND cn2.car_name LIKE ?
      )`
      params.push(`%${props.carNameFilter}%`)
    }

    if (props.vinFilter) {
      query += ` AND EXISTS (
        SELECT 1 FROM cars_stock cs2 
        WHERE cs2.id_loaded_container = lc.id 
        AND cs2.vin LIKE ?
      )`
      params.push(`%${props.vinFilter}%`)
    }

    if (props.clientNameFilter) {
      query += ` AND EXISTS (
        SELECT 1 FROM cars_stock cs2 
        LEFT JOIN clients cl2 ON cs2.id_client = cl2.id
        WHERE cs2.id_loaded_container = lc.id 
        AND cl2.name LIKE ?
      )`
      params.push(`%${props.clientNameFilter}%`)
    }

    if (props.clientIdFilter) {
      query += ` AND EXISTS (
        SELECT 1 FROM cars_stock cs2 
        LEFT JOIN clients cl2 ON cs2.id_client = cl2.id
        WHERE cs2.id_loaded_container = lc.id 
        AND cl2.id_no LIKE ?
      )`
      params.push(`%${props.clientIdFilter}%`)
    }

    if (props.containerRefFilter) {
      query += ` AND lc.ref_container LIKE ?`
      params.push(`%${props.containerRefFilter}%`)
    }

    query += ` GROUP BY lc.id ${orderByClause}`

    const result = await callApi({
      query,
      params,
    })

    console.log('Containers Query:', query)
    console.log('Containers Params:', params)
    console.log('Sold Date Filters for Containers:', {
      from: props.soldDateFrom,
      to: props.soldDateTo,
    })
    console.log('Car and Client Filters for Containers:', {
      carName: props.carNameFilter,
      vin: props.vinFilter,
      clientName: props.clientNameFilter,
      clientId: props.clientIdFilter,
      containerRef: props.containerRefFilter,
    })
    console.log('Containers query result:', result)

    if (result.success) {
      containers.value = result.data || []
      console.log('Containers data:', containers.value)
    } else {
      error.value = result.error || 'Failed to load containers'
      console.error('Containers query failed:', result.error)
    }
  } catch (err) {
    error.value = 'Error loading containers: ' + err.message
    console.error('Containers fetch error:', err)
  } finally {
    loading.value = false
  }
}

const fetchAvailableContainers = async () => {
  try {
    const result = await callApi({
      query: 'SELECT id, name FROM containers ORDER BY name ASC',
    })

    if (result.success) {
      availableContainers.value = result.data || []
    }
  } catch (err) {
    console.error('Error fetching available containers:', err)
  }
}

const openAddDialog = async () => {
  isEditing.value = false
  formData.value = {
    id_container: '',
    ref_container: '',
    so: '',
    date_departed: '',
    date_loaded: '',
    date_on_board: '',
    is_released: false,
    note: '',
  }
  // Fetch available containers for the dropdown
  await fetchAvailableContainers()
  showDialog.value = true
}

const editContainer = async (container) => {
  console.log('Editing container:', container)
  isEditing.value = true
  formData.value = {
    id: container.id,
    id_container: container.id_container || '',
    ref_container: container.ref_container || '',
    so: container.so || '',
    date_departed: container.date_departed || '',
    date_loaded: container.date_loaded || '',
    date_on_board: container.date_on_board || '',
    is_released: container.is_released,
    note: container.note || '',
  }
  console.log('Form data set to:', formData.value)
  // Fetch available containers for the dropdown
  await fetchAvailableContainers()
  showDialog.value = true
}

const closeDialog = () => {
  showDialog.value = false
  formData.value = {
    id_container: '',
    ref_container: '',
    so: '',
    date_departed: '',
    date_loaded: '',
    date_on_board: '',
    is_released: false,
    note: '',
  }
  isSubmitting.value = false
}

const saveContainer = async () => {
  if (isSubmitting.value) return

  isSubmitting.value = true

  try {
    // Convert empty date strings to null
    const dateDeparted =
      formData.value.date_departed && formData.value.date_departed.trim()
        ? formData.value.date_departed
        : null
    const dateLoaded =
      formData.value.date_loaded && formData.value.date_loaded.trim()
        ? formData.value.date_loaded
        : null
    const dateOnBoard =
      formData.value.date_on_board && formData.value.date_on_board.trim()
        ? formData.value.date_on_board
        : null

    const newContainerRef =
      formData.value.ref_container && formData.value.ref_container.trim()
        ? formData.value.ref_container
        : null

    const newSoValue = formData.value.so && formData.value.so.trim() ? formData.value.so : null

    const query = isEditing.value
      ? 'UPDATE loaded_containers SET id_container = ?, ref_container = ?, so = ?, date_departed = ?, date_loaded = ?, date_on_board = ?, is_released = ?, note = ? WHERE id = ?'
      : 'INSERT INTO loaded_containers (id_loading, id_container, ref_container, so, date_departed, date_loaded, date_on_board, is_released, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)'

    const params = isEditing.value
      ? [
          formData.value.id_container,
          newContainerRef,
          newSoValue,
          dateDeparted,
          dateLoaded,
          dateOnBoard,
          formData.value.is_released ? 1 : 0,
          formData.value.note && formData.value.note.trim() ? formData.value.note : null,
          formData.value.id,
        ]
      : [
          props.selectedLoadingId,
          formData.value.id_container,
          newContainerRef,
          newSoValue,
          dateDeparted,
          dateLoaded,
          dateOnBoard,
          formData.value.is_released ? 1 : 0,
          formData.value.note && formData.value.note.trim() ? formData.value.note : null,
        ]

    const result = await callApi({ query, params })

    if (result.success) {
      // If this is a new container (not editing), emit the container-created event
      if (!isEditing.value && result.lastInsertId) {
        // Get the complete loaded container data
        const fetchResult = await callApi({
          query: `
            SELECT 
              lc.id,
              lc.id_container,
              c.name as container_name,
              lc.ref_container,
              lc.so,
              lc.date_departed,
              lc.date_loaded,
              lc.date_on_board,
              lc.is_released,
              lc.note,
              COUNT(cs.id) as assigned_cars_count
            FROM loaded_containers lc
            LEFT JOIN containers c ON lc.id_container = c.id
            LEFT JOIN cars_stock cs ON lc.id = cs.id_loaded_container
            WHERE lc.id = ?
            GROUP BY lc.id
          `,
          params: [result.lastInsertId],
        })

        if (fetchResult.success && fetchResult.data.length > 0) {
          const newLoadedContainer = fetchResult.data[0]
          emit('container-created', newLoadedContainer)
        }
      }

      // If editing and container reference changed, update all associated cars
      if (isEditing.value && newContainerRef !== null) {
        try {
          const updateCarsResult = await callApi({
            query: 'UPDATE cars_stock SET container_ref = ? WHERE id_loaded_container = ?',
            params: [newContainerRef, formData.value.id],
          })

          if (!updateCarsResult.success) {
            console.warn(
              'Failed to update container_ref for associated cars:',
              updateCarsResult.error,
            )
          }
        } catch (err) {
          console.warn('Error updating container_ref for associated cars:', err)
        }
      }

      await fetchContainers()
      closeDialog()
    } else {
      error.value = result.error || 'Failed to save container'
    }
  } catch (err) {
    error.value = 'Error saving container: ' + err.message
  } finally {
    isSubmitting.value = false
  }
}

const openAddContainerDialog = () => {
  newContainerData.value = { name: '' }
  showAddContainerDialog.value = true
}

const closeAddContainerDialog = () => {
  showAddContainerDialog.value = false
  newContainerData.value = { name: '' }
  isAddingContainer.value = false
}

const saveNewContainer = async () => {
  if (isAddingContainer.value) return

  isAddingContainer.value = true

  try {
    const result = await callApi({
      query: 'INSERT INTO containers (name) VALUES (?)',
      params: [newContainerData.value.name],
    })

    if (result.success) {
      // Get the newly created container ID
      const newContainerId = parseInt(result.lastInsertId)

      // Refresh the available containers list
      await fetchAvailableContainers()

      // Create a loaded container record for this container
      const loadedContainerResult = await callApi({
        query: 'INSERT INTO loaded_containers (id_loading, id_container) VALUES (?, ?)',
        params: [props.selectedLoadingId, newContainerId],
      })

      if (loadedContainerResult.success) {
        const loadedContainerId = loadedContainerResult.lastInsertId

        // Get the complete loaded container data after creation
        const fetchResult = await callApi({
          query: `
            SELECT 
              lc.id,
              lc.id_container,
              c.name as container_name,
              lc.ref_container,
              lc.so,
              lc.date_departed,
              lc.date_loaded,
              lc.date_on_board,
              lc.is_released,
              lc.note,
              COUNT(cs.id) as assigned_cars_count
            FROM loaded_containers lc
            LEFT JOIN containers c ON lc.id_container = c.id
            LEFT JOIN cars_stock cs ON lc.id = cs.id_loaded_container
            WHERE lc.id = ?
            GROUP BY lc.id
          `,
          params: [loadedContainerId],
        })

        if (fetchResult.success && fetchResult.data.length > 0) {
          const newLoadedContainer = fetchResult.data[0]

          // Refresh the containers table to show the new loaded container
          await fetchContainers()

          // Emit the saved loaded container data to parent component
          emit('container-created', newLoadedContainer)

          // Set the newly created container as selected in the form
          formData.value.id_container = newContainerId.toString()
        } else {
          console.error('Failed to fetch newly created loaded container data')
        }
      } else {
        console.error('Failed to create loaded container record:', loadedContainerResult.error)
        error.value = 'Failed to create loaded container record: ' + loadedContainerResult.error
      }

      closeAddContainerDialog()
    } else {
      error.value = result.error || 'Failed to create container'
    }
  } catch (err) {
    error.value = 'Error creating container: ' + err.message
  } finally {
    isAddingContainer.value = false
  }
}

const deleteContainer = (container) => {
  containerToDelete.value = container
  showDeleteDialog.value = true
}

const closeDeleteDialog = () => {
  showDeleteDialog.value = false
  containerToDelete.value = null
  isDeleting.value = false
}

const confirmDelete = async () => {
  if (isDeleting.value || !containerToDelete.value) return

  isDeleting.value = true

  try {
    // First, unassign all cars from this container and clear their container_ref
    const unassignCarsResult = await callApi({
      query: `
        UPDATE cars_stock 
        SET id_loaded_container = NULL, container_ref = NULL 
        WHERE id_loaded_container = ?
      `,
      params: [containerToDelete.value.id],
    })

    if (!unassignCarsResult.success) {
      throw new Error('Failed to unassign cars from container: ' + unassignCarsResult.error)
    }

    console.log(
      `Unassigned ${unassignCarsResult.affectedRows || 0} cars from container ${containerToDelete.value.id}`,
    )

    // Now delete the container
    const result = await callApi({
      query: 'DELETE FROM loaded_containers WHERE id = ?',
      params: [containerToDelete.value.id],
    })

    if (result.success) {
      await fetchContainers()
      closeDeleteDialog()

      // Show success message with number of cars unassigned
      const carsUnassigned = unassignCarsResult.affectedRows || 0
      if (carsUnassigned > 0) {
        alert(
          `Container deleted successfully. ${carsUnassigned} car(s) have been unassigned from this container.`,
        )
      } else {
        alert('Container deleted successfully.')
      }

      // Emit refresh unassigned cars event
      emit('refresh-unassigned-cars')
    } else {
      error.value = result.error || 'Failed to delete container'
    }
  } catch (err) {
    error.value = 'Error deleting container: ' + err.message
  } finally {
    isDeleting.value = false
  }
}

const refreshData = () => {
  fetchContainers()
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

const truncateText = (text, maxLength) => {
  if (!text) return '-'
  if (text.length <= maxLength) return text
  return text.substring(0, maxLength) + '...'
}

const handleContainerClick = (container) => {
  emit('container-click', container.id)
}

const setOnBoard = (container) => {
  // Check if container reference is null or empty
  if (!container.ref_container || container.ref_container.trim() === '') {
    alert('Cannot set container on board. Please set the container reference first.')
    return
  }

  // Check if container has assigned cars
  if (container.assigned_cars_count === 0) {
    alert('Cannot set container on board. No cars are assigned to this container.')
    return
  }

  // Check if container is already on board
  if (container.date_on_board) {
    alert('Container is already on board.')
    return
  }

  // Set today's date as default
  const today = new Date().toISOString().split('T')[0]

  onBoardData.value = {
    id: container.id,
    date_on_board: today,
    container_name: container.container_name,
    ref_container: container.ref_container,
    assigned_cars_count: container.assigned_cars_count,
  }
  showOnBoardDialog.value = true
}

const closeOnBoardDialog = () => {
  showOnBoardDialog.value = false
  isSubmittingOnBoard.value = false
}

const confirmOnBoard = async () => {
  if (isSubmittingOnBoard.value || !onBoardData.value.date_on_board) return

  if (
    !confirm(
      `Are you sure you want to set container "${onBoardData.value.container_name}" on board for ${onBoardData.value.date_on_board}? This action cannot be undone.`,
    )
  ) {
    return
  }

  isSubmittingOnBoard.value = true

  try {
    const result = await callApi({
      query: 'UPDATE loaded_containers SET date_on_board = ? WHERE id = ?',
      params: [onBoardData.value.date_on_board, onBoardData.value.id],
    })

    if (result.success) {
      await fetchContainers()
      closeOnBoardDialog()
    } else {
      error.value = result.error || 'Failed to set container on board'
    }
  } catch (err) {
    error.value = 'Error setting container on board: ' + err.message
  } finally {
    isSubmittingOnBoard.value = false
  }
}

const getOnBoardButtonTitle = (container) => {
  if (container.date_on_board) {
    return t('loading.container_already_on_board')
  }
  if (!container.ref_container || container.ref_container.trim() === '') {
    return t('loading.set_container_reference_first')
  }
  if (container.assigned_cars_count === 0) {
    return t('loading.no_cars_assigned_to_this_container')
  }
  return t('loading.set_on_board')
}

const sortByColumn = (column) => {
  if (sortBy.value === column) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = column
    sortOrder.value = 'asc'
  }
  fetchContainers()
}

// Watch for changes in selectedLoadingId
watch(
  () => props.selectedLoadingId,
  (newId) => {
    if (newId) {
      fetchContainers()
    } else {
      containers.value = []
    }
  },
  { immediate: true },
)

// Watch for changes in sold date filters
watch(
  () => [props.soldDateFrom, props.soldDateTo],
  () => {
    if (props.selectedLoadingId) {
      fetchContainers()
    }
  },
)

// Watch for changes in car and client filters
watch(
  () => [
    props.carNameFilter,
    props.vinFilter,
    props.clientNameFilter,
    props.clientIdFilter,
    props.containerRefFilter,
  ],
  () => {
    if (props.selectedLoadingId) {
      fetchContainers()
    }
  },
)

// Fetch available containers when component mounts
onMounted(() => {
  fetchAvailableContainers()
})

// Expose methods to parent component
defineExpose({
  fetchContainers,
  fetchAvailableContainers,
})
</script>

<style scoped>
.containers-table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  margin-top: 16px;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
}

.table-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.table-header h3 i {
  color: #3498db;
}

.filter-indicator {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 8px;
  padding: 2px 8px;
  background-color: #e3f2fd;
  color: #1976d2;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
}

.filter-indicator i {
  font-size: 0.7rem;
}

.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: #f8fafc;
  font-size: 0.85rem;
}

.container-count {
  color: #6b7280;
}

.filtered-note {
  color: #1976d2;
  font-style: italic;
  margin-left: 8px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.add-btn,
.refresh-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.add-btn {
  background-color: #3498db;
  color: white;
}

.add-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.refresh-btn {
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.refresh-btn.processing {
  opacity: 0.7;
}

.loading-indicator {
  margin-left: 4px;
}

.table-wrapper {
  position: relative;
  max-height: 400px;
  overflow-y: auto;
}

.loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 20px;
  background-color: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  margin: 16px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #6b7280;
}

.empty-state i {
  margin-bottom: 16px;
  color: #d1d5db;
}

.empty-state p {
  margin: 0 0 16px 0;
  font-size: 1rem;
}

.add-first-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.add-first-btn:hover {
  background-color: #2980b9;
}

.table-content {
  overflow-x: auto;
}

.containers-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.containers-table th {
  background-color: #f8fafc;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 10;
}

.containers-table td {
  padding: 12px 8px;
  border-bottom: 1px solid #f3f4f6;
  color: #374151;
}

.table-row:hover {
  background-color: #f9fafb;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.table-row:active {
  background-color: #e5e7eb;
}

.selected-row {
  background-color: #dbeafe !important;
  border-left: 4px solid #3b82f6;
}

.selected-row:hover {
  background-color: #bfdbfe !important;
}

.id-cell {
  font-weight: 600;
  color: #6b7280;
  width: 60px;
}

.container-cell {
  font-weight: 500;
}

.ref-cell {
  font-family: monospace;
  /* background-color: #f3f4f6; */
  border-radius: 4px;
  padding: 2px 6px;
}

.ref-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.no-ref-error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #dc2626;
  font-size: 0.85rem;
  font-weight: 500;
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 4px;
  padding: 4px 8px;
}

.no-ref-error i {
  color: #dc2626;
  font-size: 0.8rem;
}

.date-cell {
  color: #6b7280;
  white-space: nowrap;
}

.released-cell {
  width: 100px;
}

.note-cell {
  max-width: 200px;
  color: #6b7280;
}

.actions-cell {
  width: 100px;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.8rem;
}

.edit-btn {
  background-color: #fbbf24;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background-color: #f59e0b;
}

.onboard-btn {
  background-color: #3498db;
  color: white;
  border: 1px solid #2980b9;
}

.onboard-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.onboard-btn:disabled {
  background-color: #e5e7eb;
  color: #9ca3af;
  border-color: #d1d5db;
  cursor: not-allowed;
}

.onboard-btn:disabled:hover {
  background-color: #e5e7eb;
  color: #9ca3af;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

/* Dialog Styles */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
}

.dialog-header h3 i {
  color: #3498db;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.dialog-content {
  padding: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.form-group label i {
  color: #3498db;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.form-group input:disabled,
.form-group select:disabled,
.form-group textarea:disabled {
  background-color: #f9fafb;
  color: #6b7280;
}

.input-with-button {
  display: flex;
  gap: 8px;
}

.input-with-button select {
  flex: 1;
}

.add-item-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 42px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.add-item-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}

.cancel-btn,
.save-btn,
.delete-confirm-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 8px;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
}

.save-btn {
  background-color: #3498db;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.save-btn.processing {
  opacity: 0.7;
}

.delete-confirm-btn {
  background-color: #ef4444;
  color: white;
}

.delete-confirm-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.delete-confirm-btn.processing {
  opacity: 0.7;
}

/* Delete Dialog Specific Styles */
.delete-dialog {
  max-width: 400px;
}

.delete-message {
  margin-bottom: 20px;
}

.delete-message p {
  margin: 0 0 12px 0;
  color: #374151;
}

/* On Board Dialog Specific Styles */
.info-display {
  background-color: #f8fafc;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  margin-bottom: 16px;
}

.info-display p {
  margin: 4px 0;
  font-size: 0.9rem;
  color: #374151;
}

.warning-message {
  background-color: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 16px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.warning-message i {
  color: #d97706;
  margin-top: 2px;
}

.warning-message p {
  margin: 0;
  font-size: 0.9rem;
  color: #92400e;
}

.sortable-header {
  cursor: pointer;
}

.sort-inactive {
  opacity: 0.5;
}

/* Status styles */
.status-available,
.status-sold {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9em;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.status-available:hover,
.status-sold:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.status-available {
  color: #10b981;
  background-color: #d1fae5;
  border: 1px solid #a7f3d0;
}

.status-sold {
  color: #ef4444;
  background-color: #fee2e2;
  border: 1px solid #fca5a5;
}

/* Released status styles */
.status-released,
.status-not-released {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.9em;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.status-released {
  color: #10b981;
  background-color: #d1fae5;
  border: 1px solid #a7f3d0;
}

.status-not-released {
  color: #f59e0b;
  background-color: #fef3c7;
  border: 1px solid #fde68a;
}
</style>
