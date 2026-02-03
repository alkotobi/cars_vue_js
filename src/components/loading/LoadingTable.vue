<template>
  <div class="loading-table-container">
    <div class="table-header">
      <h3>
        <i class="fas fa-truck-loading"></i>
        {{ t('loading.recent_loading_records') }}
      </h3>
      <div class="header-actions">
        <button
          @click="toggleFilters"
          class="toggle-filters-btn"
          :class="{ active: showFilters }"
          :title="t('loading.toggle_filters')"
        >
          <i class="fas fa-filter"></i>
          <span
            >{{ showFilters ? t('loading.hide') : t('loading.show') }}
            {{ t('loading.filters') }}</span
          >
        </button>
        <button @click="openAddDialog" class="add-btn" :disabled="loading">
          <i class="fas fa-plus"></i>
          <span>{{ t('loading.add_record') }}</span>
        </button>
        <button
          @click="printLoadingRecord"
          class="print-btn"
          :disabled="!selectedLoadingId || loading || isPrintingRecord"
          :title="t('loading.print_loading_record')"
        >
          <i v-if="isPrintingRecord" class="fas fa-spinner fa-spin"></i>
          <i v-else class="fas fa-print"></i>
          <span>{{ isPrintingRecord ? t('loading.printing') : t('loading.print') }}</span>
        </button>
        <button
          @click="refreshData"
          class="refresh-btn"
          :disabled="loading"
          :class="{ processing: loading }"
        >
          <i class="fas fa-sync-alt"></i>
          <span>{{ t('loading.refresh') }}</span>
          <i v-if="loading" class="fas fa-spinner fa-spin loading-indicator"></i>
        </button>
      </div>
    </div>

    <!-- Filters Section -->
    <div v-if="showFilters" class="filters-section">
      <div class="filters-header">
        <h4>
          <i class="fas fa-filter"></i>
          {{ t('loading.filters') }}
        </h4>
        <button @click="clearFilters" class="clear-filters-btn">
          <i class="fas fa-times"></i>
          {{ t('loading.clear_filters') }}
        </button>
      </div>

      <div class="filters-grid">
        <!-- Search & General Filters Group -->
        <div class="filter-group-category search-filters">
          <div class="category-header">
            <i class="fas fa-search"></i>
            <span>Search & General</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-search"></i>
                {{ t('loading.search') }}:
              </label>
              <input
                type="text"
                v-model="filters.search"
                :placeholder="t('loading.search_placeholder')"
              />
            </div>
          </div>
        </div>

        <!-- Vehicle Information Filters Group -->
        <div class="filter-group-category vehicle-filters">
          <div class="category-header">
            <i class="fas fa-car"></i>
            <span>Vehicle Information</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-car"></i>
                {{ t('loading.car_name') }}:
              </label>
              <input
                type="text"
                v-model="filters.carName"
                :placeholder="t('loading.filter_by_car_name')"
              />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-barcode"></i>
                {{ t('loading.vin') }}:
              </label>
              <input type="text" v-model="filters.vin" :placeholder="t('loading.filter_by_vin')" />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-hashtag"></i>
                Car ID:
              </label>
              <input type="text" v-model="filters.carId" placeholder="Filter by car ID..." />
            </div>
          </div>
        </div>

        <!-- Client & Business Filters Group -->
        <div class="filter-group-category client-filters">
          <div class="category-header">
            <i class="fas fa-users"></i>
            <span>Client & Business</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-user"></i>
                {{ t('loading.client_name') }}:
              </label>
              <input
                type="text"
                v-model="filters.clientName"
                :placeholder="t('loading.filter_by_client_name')"
              />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-id-card"></i>
                {{ t('loading.client_id_no') }}:
              </label>
              <input type="text" v-model="filters.clientId" placeholder="Filter by client ID..." />
            </div>
          </div>
        </div>

        <!-- Container & Storage Filters Group -->
        <div class="filter-group-category container-filters">
          <div class="category-header">
            <i class="fas fa-boxes"></i>
            <span>Container & Storage</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-ship"></i>
                Container ID:
              </label>
              <input
                type="text"
                v-model="filters.container"
                placeholder="Filter by container ID..."
              />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-box"></i>
                {{ t('loading.container_ref') }}:
              </label>
              <input
                type="text"
                v-model="filters.containerRef"
                :placeholder="t('loading.filter_by_container_ref')"
              />
            </div>
          </div>
        </div>

        <!-- Shipping & Logistics Filters Group -->
        <div class="filter-group-category shipping-filters">
          <div class="category-header">
            <i class="fas fa-ship"></i>
            <span>Shipping & Logistics</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-ship"></i>
                {{ t('loading.shipping_line') }}:
              </label>
              <select v-model="filters.shippingLine" @change="handleFilterChange">
                <option value="">{{ t('loading.all_shipping_lines') }}</option>
                <option v-for="line in shippingLines" :key="line.id" :value="line.name">
                  {{ line.name }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-anchor"></i>
                {{ t('loading.loading_port') }}:
              </label>
              <select v-model="filters.loadingPort" @change="handleFilterChange">
                <option value="">{{ t('loading.all_loading_ports') }}</option>
                <option v-for="port in loadingPorts" :key="port.id" :value="port.loading_port">
                  {{ port.loading_port }}
                </option>
              </select>
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-map-marker-alt"></i>
                {{ t('loading.discharge_port') }}:
              </label>
              <select v-model="filters.dischargePort" @change="handleFilterChange">
                <option value="">{{ t('loading.all_discharge_ports') }}</option>
                <option v-for="port in dischargePorts" :key="port.id" :value="port.discharge_port">
                  {{ port.discharge_port }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- Date & Time Filters Group -->
        <div class="filter-group-category date-filters">
          <div class="category-header">
            <i class="fas fa-calendar-alt"></i>
            <span>Date & Time</span>
          </div>
          <div class="category-filters">
            <div class="filter-group">
              <label>
                <i class="fas fa-calendar"></i>
                {{ t('loading.date_from') }}:
              </label>
              <input type="date" v-model="filters.dateFrom" @change="handleFilterChange" />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-calendar"></i>
                {{ t('loading.date_to') }}:
              </label>
              <input type="date" v-model="filters.dateTo" @change="handleFilterChange" />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-dollar-sign"></i>
                {{ t('loading.sold_date_from') }}:
              </label>
              <input type="date" v-model="filters.soldDateFrom" @change="handleFilterChange" />
            </div>
            <div class="filter-group">
              <label>
                <i class="fas fa-dollar-sign"></i>
                {{ t('loading.sold_date_to') }}:
              </label>
              <input type="date" v-model="filters.soldDateTo" @change="handleFilterChange" />
            </div>
          </div>
        </div>
      </div>

      <!-- Apply Filter Button -->
      <div class="filter-actions">
        <button type="button" @click="applyFilters" class="apply-filter-btn" :disabled="loading">
          <i class="fas fa-search"></i>
          {{ t('loading.apply_filters') }}
        </button>
        <button type="button" @click="clearFilters" class="clear-filters-btn" :disabled="loading">
          <i class="fas fa-broom"></i>
          {{ t('loading.clear_filters') }}
        </button>
      </div>
    </div>

    <div class="table-wrapper">
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>{{ t('loading.loading') }}</span>
      </div>

      <div v-else-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>

      <div v-else-if="loadingRecords.length === 0" class="empty-state">
        <i class="fas fa-inbox fa-2x"></i>
        <p>{{ t('loading.no_loading_records_found') }}</p>
        <button @click="openAddDialog" class="add-first-btn">
          <i class="fas fa-plus"></i>
          {{ t('loading.add_first_record') }}
        </button>
      </div>

      <div v-else class="table-content">
        <table class="loading-table">
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
              <th @click="sortByColumn('date_loading')" class="sortable-header">
                {{ t('loading.operation_date') }}
                <i
                  v-if="sortBy === 'date_loading'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('shipping_line_name')" class="sortable-header">
                {{ t('loading.shipping_line') }}
                <i
                  v-if="sortBy === 'shipping_line_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('freight')" class="sortable-header">
                {{ t('loading.freight') }}
                <i
                  v-if="sortBy === 'freight'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('loading_port_name')" class="sortable-header">
                {{ t('loading.loading_port') }}
                <i
                  v-if="sortBy === 'loading_port_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('discharge_port_name')" class="sortable-header">
                {{ t('loading.discharge_port') }}
                <i
                  v-if="sortBy === 'discharge_port_name'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('EDD')" class="sortable-header">
                {{ t('loading.edd') }}
                <i
                  v-if="sortBy === 'EDD'"
                  :class="sortOrder === 'asc' ? 'fas fa-sort-up' : 'fas fa-sort-down'"
                ></i>
                <i v-else class="fas fa-sort sort-inactive"></i>
              </th>
              <th @click="sortByColumn('date_loaded')" class="sortable-header">
                {{ t('loading.loaded_date') }}
                <i
                  v-if="sortBy === 'date_loaded'"
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
              v-for="record in loadingRecords"
              :key="record.id"
              @click="selectLoadingRecord(record.id)"
              :class="{ 'selected-row': selectedLoadingId === record.id }"
              :data-record-id="record.id"
              class="table-row"
            >
              <td class="id-cell">#{{ record.id }}</td>
              <td class="date-cell">
                {{ record.date_loading ? formatDate(record.date_loading) : '-' }}
              </td>
              <td class="shipping-line-cell">
                {{ record.shipping_line_name || '-' }}
              </td>
              <td class="freight-cell">
                {{ record.freight ? `$${record.freight}` : '-' }}
              </td>
              <td class="port-cell">
                {{ record.loading_port_name || '-' }}
              </td>
              <td class="port-cell">
                {{ record.discharge_port_name || '-' }}
              </td>
              <td class="date-cell">
                {{ record.EDD ? formatDate(record.EDD) : '-' }}
              </td>
              <td class="date-cell">
                {{ record.date_loaded ? formatDate(record.date_loaded) : '-' }}
              </td>
              <td class="notes-cell" :title="record.note">
                {{ record.note ? truncateText(record.note, 30) : '-' }}
              </td>

              <td class="actions-cell">
                <div class="action-buttons">
                  <button
                    @click="editRecord(record)"
                    class="action-btn edit-btn"
                    :title="t('loading.edit_record')"
                  >
                    <i class="fas fa-edit"></i>
                  </button>
                  <button
                    @click="openTaskForLoading(record)"
                    class="action-btn task-btn"
                    :disabled="isProcessingTask"
                    :title="t('loading.add_task')"
                  >
                    <i v-if="isProcessingTask" class="fas fa-spinner fa-spin"></i>
                    <i v-else class="fas fa-tasks"></i>
                  </button>
                  <button
                    @click="deleteRecord(record)"
                    class="action-btn delete-btn"
                    :disabled="isDeletingRecord"
                    :title="t('loading.delete_record')"
                  >
                    <i v-if="isDeletingRecord" class="fas fa-spinner fa-spin"></i>
                    <i v-else class="fas fa-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="table-footer">
      <span class="record-count">
        {{ t('loading.showing_records', { count: loadingRecords.length, total: totalRecords }) }}
      </span>
    </div>

    <ContainersTable
      :selectedLoadingId="selectedLoadingId"
      :selectedContainerId="selectedLoadedContainerId"
      :soldDateFrom="filters.soldDateFrom"
      :soldDateTo="filters.soldDateTo"
      :carNameFilter="filters.carName"
      :vinFilter="filters.vin"
      :clientNameFilter="filters.clientName"
      :clientIdFilter="filters.clientId"
      :carIdFilter="filters.carId"
      :containerIdFilter="filters.containerRef"
      @container-click="handleContainerClick"
      @refresh-unassigned-cars="handleRefreshUnassignedCars"
      @container-created="handleContainerCreated"
      ref="containersTableRef"
    />

    <LoadingAssignedCars
      ref="assignedCarsRef"
      :selectedLoadedContainerId="selectedLoadedContainerId"
      :soldDateFrom="filters.soldDateFrom"
      :soldDateTo="filters.soldDateTo"
      :carNameFilter="filters.carName"
      :vinFilter="filters.vin"
      :clientNameFilter="filters.clientName"
      :clientIdFilter="filters.clientId"
      :carIdFilter="filters.carId"
      :containerIdFilter="filters.containerRef"
      @car-unassigned="handleCarUnassigned"
    />
    <UnassignedCars
      ref="unassignedCarsRef"
      :selectedLoadedContainerId="selectedLoadedContainerId"
      :selectedContainerOnBoard="selectedContainerOnBoard"
      :assignedCarsCount="assignedCarsCount"
      :carNameFilter="filters.carName"
      :vinFilter="filters.vin"
      :clientNameFilter="filters.clientName"
      :clientIdFilter="filters.clientId"
      :carIdFilter="filters.carId"
      :containerIdFilter="filters.containerRef"
      @car-assigned="handleCarAssigned"
    />

    <!-- Add/Edit Dialog -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-truck-loading"></i>
            {{ isEditing ? 'Edit Loading Record' : 'Add New Loading Record' }}
          </h3>
          <button class="close-btn" @click="closeDialog" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <form @submit.prevent="saveRecord" class="dialog-content">
          <div class="form-row">
            <div class="form-group">
              <label for="date_loading">
                <i class="fas fa-calendar"></i>
                Operation Date
              </label>
              <input
                type="date"
                id="date_loading"
                v-model="formData.date_loading"
                required
                :disabled="isSubmitting"
                min="1900-01-01"
                max="2100-12-31"
                @blur="validateDateField('date_loading')"
              />
            </div>

            <div class="form-group">
              <label for="id_shipping_line">
                <i class="fas fa-ship"></i>
                Shipping Line
              </label>
              <div class="input-with-button">
                <select
                  id="id_shipping_line"
                  v-model="formData.id_shipping_line"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Shipping Line</option>
                  <option v-for="line in shippingLines" :key="line.id" :value="line.id">
                    {{ line.name }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddShippingLineDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Shipping Line"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="freight">
                <i class="fas fa-dollar-sign"></i>
                Freight
              </label>
              <input
                type="number"
                id="freight"
                v-model="formData.freight"
                step="0.01"
                min="0"
                :disabled="isSubmitting"
                placeholder="0.00"
              />
            </div>

            <div class="form-group">
              <label for="id_loading_port">
                <i class="fas fa-anchor"></i>
                Loading Port
              </label>
              <div class="input-with-button">
                <select
                  id="id_loading_port"
                  v-model="formData.id_loading_port"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Loading Port</option>
                  <option v-for="port in loadingPorts" :key="port.id" :value="port.id">
                    {{ port.loading_port }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddLoadingPortDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Loading Port"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="id_discharge_port">
                <i class="fas fa-anchor"></i>
                Discharge Port
              </label>
              <div class="input-with-button">
                <select
                  id="id_discharge_port"
                  v-model="formData.id_discharge_port"
                  required
                  :disabled="isSubmitting"
                >
                  <option value="">Select Discharge Port</option>
                  <option v-for="port in dischargePorts" :key="port.id" :value="port.id">
                    {{ port.discharge_port }}
                  </option>
                </select>
                <button
                  type="button"
                  @click="openAddDischargePortDialog"
                  class="add-item-btn"
                  :disabled="isSubmitting"
                  title="Add New Discharge Port"
                >
                  <i class="fas fa-plus"></i>
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="EDD">
                <i class="fas fa-calendar-check"></i>
                EDD (Estimated Delivery Date)
              </label>
              <input type="date" id="EDD" v-model="formData.EDD" :disabled="isSubmitting" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label for="date_loaded">
                <i class="fas fa-check-circle"></i>
                Loaded Date
              </label>
              <input
                type="date"
                id="date_loaded"
                v-model="formData.date_loaded"
                :disabled="isSubmitting"
              />
            </div>
          </div>

          <div class="form-group full-width">
            <label for="note">
              <i class="fas fa-sticky-note"></i>
              Notes
            </label>
            <textarea
              id="note"
              v-model="formData.note"
              rows="3"
              :disabled="isSubmitting"
              placeholder="Enter any additional notes..."
            ></textarea>
          </div>

          <div class="dialog-actions">
            <button type="button" @click="closeDialog" class="cancel-btn" :disabled="isSubmitting">
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="submit"
              class="save-btn"
              :disabled="isSubmitting"
              :class="{ processing: isSubmitting }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isSubmitting ? 'Saving...' : isEditing ? 'Update' : 'Save' }}</span>
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Quick Add Shipping Line Dialog -->
    <div
      v-if="showShippingLineDialog"
      class="dialog-overlay"
      @click.self="showShippingLineDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-ship"></i>
            Add New Shipping Line
          </h3>
          <button
            class="close-btn"
            @click="showShippingLineDialog = false"
            :disabled="isAddingItem"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="shipping_line_name">
              <i class="fas fa-ship"></i>
              Shipping Line Name
            </label>
            <input
              type="text"
              id="shipping_line_name"
              v-model="quickAddForm.shipping_line_name"
              placeholder="Enter shipping line name"
              :disabled="isAddingItem"
              @keyup.enter="saveShippingLine"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showShippingLineDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveShippingLine"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Add Loading Port Dialog -->
    <div
      v-if="showLoadingPortDialog"
      class="dialog-overlay"
      @click.self="showLoadingPortDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-anchor"></i>
            Add New Loading Port
          </h3>
          <button class="close-btn" @click="showLoadingPortDialog = false" :disabled="isAddingItem">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="loading_port_name">
              <i class="fas fa-anchor"></i>
              Loading Port Name
            </label>
            <input
              type="text"
              id="loading_port_name"
              v-model="quickAddForm.loading_port_name"
              placeholder="Enter loading port name"
              :disabled="isAddingItem"
              @keyup.enter="saveLoadingPort"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showLoadingPortDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveLoadingPort"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Add Discharge Port Dialog -->
    <div
      v-if="showDischargePortDialog"
      class="dialog-overlay"
      @click.self="showDischargePortDialog = false"
    >
      <div class="quick-add-dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-anchor"></i>
            Add New Discharge Port
          </h3>
          <button
            class="close-btn"
            @click="showDischargePortDialog = false"
            :disabled="isAddingItem"
          >
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div class="dialog-content">
          <div class="form-group">
            <label for="discharge_port_name">
              <i class="fas fa-anchor"></i>
              Discharge Port Name
            </label>
            <input
              type="text"
              id="discharge_port_name"
              v-model="quickAddForm.discharge_port_name"
              placeholder="Enter discharge port name"
              :disabled="isAddingItem"
              @keyup.enter="saveDischargePort"
            />
          </div>

          <div class="dialog-actions">
            <button
              type="button"
              @click="showDischargePortDialog = false"
              class="cancel-btn"
              :disabled="isAddingItem"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button
              type="button"
              @click="saveDischargePort"
              class="save-btn"
              :disabled="isAddingItem"
              :class="{ processing: isAddingItem }"
            >
              <i class="fas fa-save"></i>
              <span>{{ isAddingItem ? 'Adding...' : 'Add' }}</span>
              <i v-if="isAddingItem" class="fas fa-spinner fa-spin loading-indicator"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Task Form Modal -->
  <TaskForm
    v-if="selectedLoadingForTask"
    :entity-type="'loading'"
    :entity-data="selectedLoadingForTask"
    :is-visible="showTaskForm"
    @task-created="handleTaskCreated"
    @cancel="showTaskForm = false"
  />
</template>

<script setup>
import { ref, onMounted, watch, nextTick, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useApi } from '@/composables/useApi'
import ContainersTable from './ContainersTable.vue'
import LoadingAssignedCars from './LoadingAssignedCars.vue'
import UnassignedCars from './UnassignedCars.vue'
import TaskForm from '../car-stock/TaskForm.vue'

const { t } = useI18n()
const { callApi, getFileUrl } = useApi()

const loadingRecords = ref([])
const totalRecords = ref(0)
const loading = ref(false)
const error = ref(null)

// CRUD state
const showDialog = ref(false)
const isEditing = ref(false)
const isSubmitting = ref(false)
const editingRecord = ref(null)

// Add loading states for double-click prevention
const isDeletingRecord = ref(false)
const isProcessingTask = ref(false)
const isPrintingRecord = ref(false)

// Quick add dialog states
const showShippingLineDialog = ref(false)
const showLoadingPortDialog = ref(false)
const showDischargePortDialog = ref(false)
const isAddingItem = ref(false)
const selectedLoadingId = ref(null)
const selectedLoadedContainerId = ref(null)
const selectedContainerOnBoard = ref(false)
const assignedCarsRef = ref(null)
const unassignedCarsRef = ref(null)
const containersTableRef = ref(null)

// Form data
const formData = ref({
  date_loading: '',
  id_shipping_line: '',
  freight: '',
  id_loading_port: '',
  id_discharge_port: '',
  EDD: '',
  date_loaded: '',
  note: '',
})

// Quick add form data
const quickAddForm = ref({
  shipping_line_name: '',
  loading_port_name: '',
  discharge_port_name: '',
})

// Reference data
const shippingLines = ref([])
const loadingPorts = ref([])
const dischargePorts = ref([])

// Filters and sorting state
const filters = ref({
  search: '',
  shippingLine: '',
  loadingPort: '',
  dischargePort: '',
  dateFrom: '',
  dateTo: '',
  soldDateFrom: '',
  soldDateTo: '',
  carName: '',
  vin: '',
  clientName: '',
  clientId: '',
  carId: '',
  container: '',
  containerRef: '',
})

const sortBy = ref('id')
const sortOrder = ref('desc')
const allLoadingRecords = ref([]) // Store all records for filtering
const hasActiveFilters = ref(false)

// Filter visibility state
const showFilters = ref(false)

// Add task form state
const showTaskForm = ref(false)
const selectedLoadingForTask = ref(null)

// Computed property for assigned cars count
const assignedCarsCount = computed(() => {
  if (!assignedCarsRef.value) return 0
  return assignedCarsRef.value.assignedCars?.length || 0
})

const fetchLoadingRecords = async () => {
  loading.value = true
  error.value = null

  try {
    let query = `
      SELECT 
        l.id,
        l.date_loading,
        l.id_shipping_line,
        l.freight,
        l.id_loading_port,
        l.id_discharge_port,
        l.EDD,
        l.date_loaded,
        l.note,
        sl.name as shipping_line_name,
        lp.loading_port as loading_port_name,
        dp.discharge_port as discharge_port_name,
        GROUP_CONCAT(DISTINCT cn.car_name) as car_names,
        GROUP_CONCAT(DISTINCT cs.vin) as vins,
        GROUP_CONCAT(DISTINCT cs.id) as car_ids,
        GROUP_CONCAT(DISTINCT cl.name) as client_names,
        GROUP_CONCAT(DISTINCT cl.id_no) as client_ids,
        GROUP_CONCAT(DISTINCT lc.ref_container) as container_refs,
        GROUP_CONCAT(DISTINCT lc.id_container) as container_ids
      FROM loading l
      LEFT JOIN shipping_lines sl ON l.id_shipping_line = sl.id
      LEFT JOIN loading_ports lp ON l.id_loading_port = lp.id
      LEFT JOIN discharge_ports dp ON l.id_discharge_port = dp.id
      LEFT JOIN loaded_containers lc ON l.id = lc.id_loading
      LEFT JOIN cars_stock cs ON lc.id = cs.id_loaded_container
      LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
      LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
      LEFT JOIN clients cl ON cs.id_client = cl.id
    `

    let params = []

    // Add sold date filter if specified
    if (filters.value.soldDateFrom || filters.value.soldDateTo) {
      query += `
        INNER JOIN loaded_containers lc ON l.id = lc.id_loading
        INNER JOIN cars_stock cs ON lc.id = cs.id_loaded_container
        INNER JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE sb.date_sell IS NOT NULL
      `

      if (filters.value.soldDateFrom) {
        query += ` AND DATE(sb.date_sell) >= ?`
        params.push(filters.value.soldDateFrom)
      }

      if (filters.value.soldDateTo) {
        query += ` AND DATE(sb.date_sell) <= ?`
        params.push(filters.value.soldDateTo)
      }

      query += ` GROUP BY l.id`
    } else {
      // Add GROUP BY for car and client data when not using sold date filters
      query += ` GROUP BY l.id`
    }

    query += ` ORDER BY l.id DESC, l.date_loading DESC`

    console.log('Loading Records Query:', query)
    console.log('Loading Records Params:', params)
    console.log('Sold Date Filters:', {
      from: filters.value.soldDateFrom,
      to: filters.value.soldDateTo,
    })
    console.log('All Filters Object:', filters.value)

    const result = await callApi({
      query,
      params,
    })

    console.log('Loading Records Result:', result)

    if (result.success) {
      allLoadingRecords.value = result.data || []
      console.log('All Loading Records:', allLoadingRecords.value.length)

      // Debug: Check the structure of the first record
      if (allLoadingRecords.value.length > 0) {
        console.log('Sample Loading Record Structure:', {
          id: allLoadingRecords.value[0].id,
          container_refs: allLoadingRecords.value[0].container_refs,
          container_ids: allLoadingRecords.value[0].container_ids,
          car_ids: allLoadingRecords.value[0].car_ids,
        })
      }

      // Only apply client-side filtering if no sold date filters are active
      if (!filters.value.soldDateFrom && !filters.value.soldDateTo) {
        applyFiltersAndSorting()
      } else {
        // For sold date filters, just apply sorting without additional filtering
        applySortingOnly()
      }
    } else {
      error.value = result.error || 'Failed to fetch loading records'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred while fetching loading records'
  } finally {
    loading.value = false
  }
}

const applyFiltersAndSorting = () => {
  let filteredRecords = [...allLoadingRecords.value]

  // Apply search filter
  if (filters.value.search) {
    const searchTerm = filters.value.search.toLowerCase()
    filteredRecords = filteredRecords.filter(
      (record) =>
        record.id.toString().includes(searchTerm) ||
        (record.shipping_line_name &&
          record.shipping_line_name.toLowerCase().includes(searchTerm)) ||
        (record.loading_port_name && record.loading_port_name.toLowerCase().includes(searchTerm)) ||
        (record.discharge_port_name &&
          record.discharge_port_name.toLowerCase().includes(searchTerm)) ||
        (record.container_refs && record.container_refs.toLowerCase().includes(searchTerm)) ||
        (record.note && record.note.toLowerCase().includes(searchTerm)),
    )
  }

  // Apply car name filter
  if (filters.value.carName) {
    const carNameTerm = filters.value.carName.toLowerCase()
    filteredRecords = filteredRecords.filter(
      (record) => record.car_names && record.car_names.toLowerCase().includes(carNameTerm),
    )
  }

  // Apply VIN filter
  if (filters.value.vin) {
    const vinTerm = filters.value.vin.toLowerCase()
    filteredRecords = filteredRecords.filter(
      (record) => record.vins && record.vins.toLowerCase().includes(vinTerm),
    )
  }

  // Apply client name filter
  if (filters.value.clientName) {
    const clientNameTerm = filters.value.clientName.toLowerCase()
    filteredRecords = filteredRecords.filter(
      (record) => record.client_names && record.client_names.toLowerCase().includes(clientNameTerm),
    )
  }

  // Apply client ID filter
  if (filters.value.clientId) {
    const clientIdTerm = filters.value.clientId.toLowerCase()
    filteredRecords = filteredRecords.filter(
      (record) => record.client_ids && record.client_ids.toLowerCase().includes(clientIdTerm),
    )
  }

  // Apply car ID filter
  if (filters.value.carId) {
    const carIdTerm = filters.value.carId.trim()
    filteredRecords = filteredRecords.filter(
      (record) => record.car_ids && record.car_ids.split(',').some((id) => id.trim() === carIdTerm),
    )
  }

  // Apply container filter
  if (filters.value.container) {
    const containerTerm = filters.value.container.trim()
    filteredRecords = filteredRecords.filter(
      (record) =>
        record.container_ids &&
        record.container_ids.split(',').some((id) => id.trim() === containerTerm),
    )
  }

  // Apply container reference filter
  if (filters.value.containerRef) {
    const containerRefTerm = filters.value.containerRef.toLowerCase()
    console.log('Container Ref Filter Debug:', {
      filterTerm: containerRefTerm,
      sampleRecord: filteredRecords[0],
      containerRefsField: filteredRecords[0]?.container_refs,
    })
    filteredRecords = filteredRecords.filter(
      (record) =>
        record.container_refs &&
        record.container_refs
          .split(',')
          .some((ref) => ref.trim().toLowerCase().includes(containerRefTerm)),
    )
  }

  // Apply shipping line filter
  if (filters.value.shippingLine) {
    filteredRecords = filteredRecords.filter(
      (record) => record.shipping_line_name === filters.value.shippingLine,
    )
  }

  // Apply loading port filter
  if (filters.value.loadingPort) {
    filteredRecords = filteredRecords.filter(
      (record) => record.loading_port_name === filters.value.loadingPort,
    )
  }

  // Apply discharge port filter
  if (filters.value.dischargePort) {
    filteredRecords = filteredRecords.filter(
      (record) => record.discharge_port_name === filters.value.dischargePort,
    )
  }

  // Apply date range filters
  if (filters.value.dateFrom) {
    filteredRecords = filteredRecords.filter(
      (record) => record.date_loading && record.date_loading >= filters.value.dateFrom,
    )
  }

  if (filters.value.dateTo) {
    filteredRecords = filteredRecords.filter(
      (record) => record.date_loading && record.date_loading <= filters.value.dateTo,
    )
  }

  // Apply sorting
  filteredRecords.sort((a, b) => {
    let aValue = a[sortBy.value]
    let bValue = b[sortBy.value]

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Convert to strings for comparison
    aValue = aValue.toString().toLowerCase()
    bValue = bValue.toString().toLowerCase()

    if (sortOrder.value === 'asc') {
      return aValue.localeCompare(bValue)
    } else {
      return bValue.localeCompare(aValue)
    }
  })

  // Update hasActiveFilters
  hasActiveFilters.value = !!(
    filters.value.search ||
    filters.value.shippingLine ||
    filters.value.loadingPort ||
    filters.value.dischargePort ||
    filters.value.dateFrom ||
    filters.value.dateTo ||
    filters.value.soldDateFrom ||
    filters.value.soldDateTo ||
    filters.value.carName ||
    filters.value.vin ||
    filters.value.clientName ||
    filters.value.clientId ||
    filters.value.carId ||
    filters.value.container ||
    filters.value.containerRef
  )

  // Show all filtered records (removed the 5-record limit)
  loadingRecords.value = filteredRecords
  totalRecords.value = allLoadingRecords.value.length
}

const handleFilterChange = () => {
  // Only dropdown and date filters trigger automatic updates
  // Text filters (search, carName, vin, clientName, clientId, containerRef) require manual application
  if (
    filters.value.soldDateFrom ||
    filters.value.soldDateTo ||
    filters.value.shippingLine ||
    filters.value.loadingPort ||
    filters.value.dischargePort ||
    filters.value.dateFrom ||
    filters.value.dateTo
  ) {
    fetchLoadingRecords()
  }
  // Text filters are now handled manually via the Apply Filter button
}

const applyFilters = () => {
  // Apply all filters including text filters
  if (
    filters.value.soldDateFrom ||
    filters.value.soldDateTo ||
    filters.value.carName ||
    filters.value.vin ||
    filters.value.clientName ||
    filters.value.clientId ||
    filters.value.carId ||
    filters.value.container ||
    filters.value.containerRef
  ) {
    // If any business filters are active, fetch from database
    fetchLoadingRecords()
  } else {
    // Otherwise apply client-side filtering
    applyFiltersAndSorting()
  }
}

const applySorting = () => {
  applyFiltersAndSorting()
}

const applySortingOnly = () => {
  // Apply only sorting without filtering
  let sortedRecords = [...allLoadingRecords.value]

  // Apply sorting
  sortedRecords.sort((a, b) => {
    let aValue = a[sortBy.value]
    let bValue = b[sortBy.value]

    // Handle date comparison
    if (sortBy.value === 'date_loading') {
      aValue = aValue ? new Date(aValue).getTime() : 0
      bValue = bValue ? new Date(bValue).getTime() : 0
    }

    // Handle numeric fields
    if (['id', 'freight'].includes(sortBy.value)) {
      aValue = Number(aValue) || 0
      bValue = Number(bValue) || 0
    }

    // Handle null values
    if (aValue === null || aValue === undefined) aValue = ''
    if (bValue === null || bValue === undefined) bValue = ''

    // Compare values based on direction
    if (sortOrder.value === 'asc') {
      return aValue > bValue ? 1 : -1
    } else {
      return aValue < bValue ? 1 : -1
    }
  })

  loadingRecords.value = sortedRecords
}

const sortByColumn = (column) => {
  if (sortBy.value === column) {
    // If clicking the same column, toggle the order
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    // If clicking a different column, set it as the new sort column and default to ascending
    sortBy.value = column
    sortOrder.value = 'asc'
  }

  // Apply appropriate sorting based on whether sold date filters are active
  if (filters.value.soldDateFrom || filters.value.soldDateTo) {
    applySortingOnly()
  } else {
    applyFiltersAndSorting()
  }
}

const clearFilters = () => {
  filters.value = {
    search: '',
    shippingLine: '',
    loadingPort: '',
    dischargePort: '',
    dateFrom: '',
    dateTo: '',
    soldDateFrom: '',
    soldDateTo: '',
    carName: '',
    vin: '',
    clientName: '',
    clientId: '',
    carId: '',
    container: '',
    containerRef: '',
  }
  sortBy.value = 'id'
  sortOrder.value = 'desc'
  applyFiltersAndSorting()

  // Refresh data to reset child tables
  fetchLoadingRecords()
}

const toggleFilters = () => {
  showFilters.value = !showFilters.value
}

const fetchTotalRecords = async () => {
  try {
    const result = await callApi({
      query: 'SELECT COUNT(*) as total FROM loading',
      params: [],
    })

    if (result.success) {
      totalRecords.value = result.data[0].total
    }
  } catch (err) {
    console.error('Failed to fetch total records:', err)
  }
}

const fetchReferenceData = async () => {
  try {
    // Fetch shipping lines
    const shippingLinesResult = await callApi({
      query: 'SELECT id, name FROM shipping_lines ORDER BY name',
      params: [],
    })
    if (shippingLinesResult.success) {
      shippingLines.value = shippingLinesResult.data
    }

    // Fetch loading ports
    const loadingPortsResult = await callApi({
      query: 'SELECT id, loading_port FROM loading_ports ORDER BY loading_port',
      params: [],
    })
    if (loadingPortsResult.success) {
      loadingPorts.value = loadingPortsResult.data
    }

    // Fetch discharge ports
    const dischargePortsResult = await callApi({
      query: 'SELECT id, discharge_port FROM discharge_ports ORDER BY discharge_port',
      params: [],
    })
    if (dischargePortsResult.success) {
      dischargePorts.value = dischargePortsResult.data
    }
  } catch (err) {
    console.error('Failed to fetch reference data:', err)
  }
}

const refreshData = async () => {
  await Promise.all([fetchLoadingRecords(), fetchTotalRecords()])
}

const openAddDialog = () => {
  isEditing.value = false
  editingRecord.value = null
  resetForm()
  showDialog.value = true
}

const editRecord = (record) => {
  console.log('Editing record:', record)
  isEditing.value = true
  editingRecord.value = record
  formData.value = {
    date_loading: record.date_loading || '',
    id_shipping_line: record.id_shipping_line || '',
    freight: record.freight || '',
    id_loading_port: record.id_loading_port || '',
    id_discharge_port: record.id_discharge_port || '',
    EDD: record.EDD || '',
    date_loaded: record.date_loaded || '',
    note: record.note || '',
  }
  console.log('Form data set to:', formData.value)
  showDialog.value = true
}

const deleteRecord = async (record) => {
  if (!confirm(`Are you sure you want to delete loading record #${record.id}?`)) {
    return
  }

  if (isDeletingRecord.value) return // Prevent double-click

  try {
    isDeletingRecord.value = true
    // First check if the loading record has any containers
    const containersCheck = await callApi({
      query: 'SELECT COUNT(*) as container_count FROM loaded_containers WHERE id_loading = ?',
      params: [record.id],
    })

    if (!containersCheck.success) {
      alert('Failed to check containers: ' + containersCheck.error)
      return
    }

    const containerCount = containersCheck.data[0]?.container_count || 0

    if (containerCount > 0) {
      alert(
        `Cannot delete loading record #${record.id}. This record has ${containerCount} container(s) assigned to it. Please delete all containers first.`,
      )
      return
    }

    // If no containers, proceed with deletion
    const result = await callApi({
      query: 'DELETE FROM loading WHERE id = ?',
      params: [record.id],
    })

    if (result.success) {
      await refreshData()
    } else {
      alert('Failed to delete record: ' + result.error)
    }
  } catch (err) {
    alert('Error deleting record: ' + err.message)
  } finally {
    isDeletingRecord.value = false
  }
}

const validateDateInput = (event) => {
  const input = event.target
  const value = input.value

  // Check if the value matches the expected format
  if (value && !value.match(/^\d{4}-\d{2}-\d{2}$/)) {
    alert('Invalid date format. Please use YYYY-MM-DD format.')
    input.value = ''
    return
  }

  // Check year range
  if (value) {
    const year = parseInt(value.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('Year must be between 1900 and 2100.')
      input.value = ''
      return
    }
  }
}

const saveRecord = async () => {
  if (isSubmitting.value) return

  console.log('=== DEBUG: saveRecord called ===')
  console.log('Form data:', JSON.stringify(formData.value, null, 2))
  console.log('Date loading value:', formData.value.date_loading)
  console.log('Date loading type:', typeof formData.value.date_loading)
  console.log('Date loading length:', formData.value.date_loading?.length)

  // Pre-validation for date formats
  if (formData.value.date_loading) {
    console.log('Validating date_loading:', formData.value.date_loading)
    console.log('Date regex test:', formData.value.date_loading.match(/^\d{4}-\d{2}-\d{2}$/))

    if (!formData.value.date_loading.match(/^\d{4}-\d{2}-\d{2}$/)) {
      console.log('Date format validation failed')
      alert('Invalid loading date format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.date_loading.substring(0, 4))
    console.log('Year extracted:', year)
    if (year < 1900 || year > 2100) {
      console.log('Year validation failed:', year)
      alert('Loading date year must be between 1900 and 2100.')
      return
    }
  }

  if (formData.value.EDD) {
    if (!formData.value.EDD.match(/^\d{4}-\d{2}-\d{2}$/)) {
      alert('Invalid EDD format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.EDD.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('EDD year must be between 1900 and 2100.')
      return
    }
  }

  if (formData.value.date_loaded) {
    if (!formData.value.date_loaded.match(/^\d{4}-\d{2}-\d{2}$/)) {
      alert('Invalid loaded date format. Please use YYYY-MM-DD format.')
      return
    }
    const year = parseInt(formData.value.date_loaded.substring(0, 4))
    if (year < 1900 || year > 2100) {
      alert('Loaded date year must be between 1900 and 2100.')
      return
    }
  }

  // Validate and format dates
  const validateAndFormatDate = (dateString) => {
    if (!dateString || !dateString.trim()) return null
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      throw new Error('Invalid date format')
    }
    return date.toISOString().split('T')[0]
  }

  try {
    // Validate required fields
    if (!formData.value.date_loading) {
      alert('Loading date is required')
      return
    }

    // Format dates - convert empty strings to null
    const formattedDateLoading = validateAndFormatDate(formData.value.date_loading)
    const formattedEDD = validateAndFormatDate(formData.value.EDD)
    const formattedDateLoaded = validateAndFormatDate(formData.value.date_loaded)

    console.log('Formatted dates:', {
      date_loading: formattedDateLoading,
      EDD: formattedEDD,
      date_loaded: formattedDateLoaded,
    })

    isSubmitting.value = true
    let result

    if (isEditing.value) {
      // Update existing record
      result = await callApi({
        query: `
          UPDATE loading 
          SET date_loading = ?, id_shipping_line = ?, freight = ?, 
              id_loading_port = ?, id_discharge_port = ?, EDD = ?, 
              date_loaded = ?, note = ?
          WHERE id = ?
        `,
        params: [
          formattedDateLoading,
          formData.value.id_shipping_line || null,
          formData.value.freight || null,
          formData.value.id_loading_port || null,
          formData.value.id_discharge_port || null,
          formattedEDD,
          formattedDateLoaded,
          formData.value.note || null,
          editingRecord.value.id,
        ],
      })
    } else {
      // Insert new record
      const insertParams = [
        formattedDateLoading,
        formData.value.id_shipping_line || null,
        formData.value.freight || null,
        formData.value.id_loading_port || null,
        formData.value.id_discharge_port || null,
        formattedEDD,
        formattedDateLoaded,
        formData.value.note || null,
      ]

      console.log('Insert parameters:', insertParams)

      result = await callApi({
        query: `
          INSERT INTO loading 
          (date_loading, id_shipping_line, freight, id_loading_port, 
           id_discharge_port, EDD, date_loaded, note)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `,
        params: insertParams,
      })
    }

    if (result.success) {
      console.log('Save successful, result:', result)
      console.log('isEditing:', isEditing.value)
      console.log('insertId:', result.lastInsertId)

      closeDialog()

      // If this was a new record, get the saved record data
      let savedRecord = null
      if (!isEditing.value && result.lastInsertId) {
        console.log('Fetching new record data for ID:', result.lastInsertId)
        // For new records, we need to fetch the complete record data
        const fetchResult = await callApi({
          query: `
            SELECT l.*, 
                   sl.name as shipping_line_name,
                   lp.loading_port as loading_port_name,
                   dp.discharge_port as discharge_port_name
            FROM loading l
            LEFT JOIN shipping_lines sl ON l.id_shipping_line = sl.id
            LEFT JOIN loading_ports lp ON l.id_loading_port = lp.id
            LEFT JOIN discharge_ports dp ON l.id_discharge_port = dp.id
            WHERE l.id = ?
          `,
          params: [result.lastInsertId],
        })

        console.log('Fetch result:', fetchResult)

        if (fetchResult.success && fetchResult.data.length > 0) {
          savedRecord = fetchResult.data[0]
          console.log('Saved record data:', savedRecord)
        }
      } else if (isEditing.value) {
        console.log('Using edited record data')
        // For edited records, use the current form data with the record ID
        savedRecord = {
          id: editingRecord.value.id,
          ...formData.value,
        }
        console.log('Edited record data:', savedRecord)
      }

      console.log('About to refresh data')
      await refreshData()
      console.log('Data refreshed, loadingRecords length:', loadingRecords.value.length)

      // If we have a saved record, select it after the data is refreshed
      if (savedRecord) {
        console.log('Setting selectedLoadingId to:', savedRecord.id)
        selectedLoadingId.value = savedRecord.id
        // Scroll to the newly added/edited record
        nextTick(() => {
          console.log('Looking for element with data-record-id:', savedRecord.id)
          const recordElement = document.querySelector(`[data-record-id="${savedRecord.id}"]`)
          console.log('Found element:', recordElement)
          if (recordElement) {
            recordElement.scrollIntoView({ behavior: 'smooth', block: 'center' })
            // Add highlight effect
            recordElement.classList.add('highlight-new-record')
            setTimeout(() => {
              recordElement.classList.remove('highlight-new-record')
            }, 3000)
          }
        })
      } else {
        console.log('No saved record to select')
      }
    } else {
      alert('Failed to save record: ' + result.error)
    }
  } catch (err) {
    if (err.message === 'Invalid date format') {
      alert('Please check the date format. Dates should be in YYYY-MM-DD format.')
    } else {
      alert('Error saving record: ' + err.message)
    }
  } finally {
    isSubmitting.value = false
  }
}

const closeDialog = () => {
  showDialog.value = false
  resetForm()
}

const resetForm = () => {
  const today = new Date().toISOString().split('T')[0]
  formData.value = {
    date_loading: today,
    id_shipping_line: '',
    freight: '',
    id_loading_port: '',
    id_discharge_port: '',
    EDD: '',
    date_loaded: '',
    note: '',
  }
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

const viewAllRecords = () => {
  // TODO: Implement view all records functionality
  console.log('View all records clicked')
}

const openAddShippingLineDialog = () => {
  showShippingLineDialog.value = true
  quickAddForm.value.shipping_line_name = ''
}

const openAddLoadingPortDialog = () => {
  showLoadingPortDialog.value = true
  quickAddForm.value.loading_port_name = ''
}

const openAddDischargePortDialog = () => {
  showDischargePortDialog.value = true
  quickAddForm.value.discharge_port_name = ''
}

const saveShippingLine = async () => {
  if (!quickAddForm.value.shipping_line_name.trim()) {
    alert('Please enter a shipping line name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO shipping_lines (name) VALUES (?)',
      params: [quickAddForm.value.shipping_line_name.trim()],
    })

    if (result.success) {
      // Refresh shipping lines data
      await fetchReferenceData()
      showShippingLineDialog.value = false
      quickAddForm.value.shipping_line_name = ''
    } else {
      alert('Failed to add shipping line: ' + result.error)
    }
  } catch (err) {
    alert('Error adding shipping line: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

const saveLoadingPort = async () => {
  if (!quickAddForm.value.loading_port_name.trim()) {
    alert('Please enter a loading port name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO loading_ports (loading_port) VALUES (?)',
      params: [quickAddForm.value.loading_port_name.trim()],
    })

    if (result.success) {
      // Refresh loading ports data
      await fetchReferenceData()
      showLoadingPortDialog.value = false
      quickAddForm.value.loading_port_name = ''
    } else {
      alert('Failed to add loading port: ' + result.error)
    }
  } catch (err) {
    alert('Error adding loading port: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

const saveDischargePort = async () => {
  if (!quickAddForm.value.discharge_port_name.trim()) {
    alert('Please enter a discharge port name')
    return
  }

  isAddingItem.value = true
  try {
    const result = await callApi({
      query: 'INSERT INTO discharge_ports (discharge_port) VALUES (?)',
      params: [quickAddForm.value.discharge_port_name.trim()],
    })

    if (result.success) {
      // Refresh discharge ports data
      await fetchReferenceData()
      showDischargePortDialog.value = false
      quickAddForm.value.discharge_port_name = ''
    } else {
      alert('Failed to add discharge port: ' + result.error)
    }
  } catch (err) {
    alert('Error adding discharge port: ' + err.message)
  } finally {
    isAddingItem.value = false
  }
}

// Watchers
watch(
  formData,
  (newVal) => {
    // Validate date formats as user types
    if (newVal.date_loading && newVal.date_loading.length === 10) {
      if (!newVal.date_loading.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid date format. Please use YYYY-MM-DD format.')
        formData.value.date_loading = ''
        return
      }

      const year = parseInt(newVal.date_loading.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('Year must be between 1900 and 2100.')
        formData.value.date_loading = ''
        return
      }
    }

    if (newVal.EDD && newVal.EDD.length === 10) {
      if (!newVal.EDD.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid EDD format. Please use YYYY-MM-DD format.')
        formData.value.EDD = ''
        return
      }

      const year = parseInt(newVal.EDD.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('EDD year must be between 1900 and 2100.')
        formData.value.EDD = ''
        return
      }
    }

    if (newVal.date_loaded && newVal.date_loaded.length === 10) {
      if (!newVal.date_loaded.match(/^\d{4}-\d{2}-\d{2}$/)) {
        alert('Invalid loaded date format. Please use YYYY-MM-DD format.')
        formData.value.date_loaded = ''
        return
      }

      const year = parseInt(newVal.date_loaded.substring(0, 4))
      if (year < 1900 || year > 2100) {
        alert('Loaded date year must be between 1900 and 2100.')
        formData.value.date_loaded = ''
        return
      }
    }
  },
  { deep: true },
)

// Watch for changes in filters object
watch(
  () => filters.value,
  (newFilters) => {
    console.log('Filters changed:', newFilters)
    console.log('Sold date filters changed:', {
      from: newFilters.soldDateFrom,
      to: newFilters.soldDateTo,
    })

    // Refresh child components when filters change to apply the same filtering
    if (selectedLoadingId.value) {
      nextTick(() => {
        if (containersTableRef.value) {
          containersTableRef.value.fetchContainers()
        }
        if (assignedCarsRef.value && selectedLoadedContainerId.value) {
          assignedCarsRef.value.refreshData()
        }
        if (unassignedCarsRef.value) {
          unassignedCarsRef.value.refreshData()
        }
      })
    }
  },
  { deep: true },
)

const selectLoadingRecord = (id) => {
  selectedLoadingId.value = id
}

const handleContainerClick = async (containerId) => {
  selectedLoadedContainerId.value = containerId

  // Fetch the container's on board status
  try {
    const result = await callApi({
      query: 'SELECT date_on_board FROM loaded_containers WHERE id = ?',
      params: [containerId],
    })

    if (result.success && result.data && result.data.length > 0) {
      selectedContainerOnBoard.value = !!result.data[0].date_on_board
    } else {
      selectedContainerOnBoard.value = false
    }
  } catch (err) {
    console.error('Error fetching container on board status:', err)
    selectedContainerOnBoard.value = false
  }

  // Scroll to unassigned cars table after a short delay to ensure it's rendered
  setTimeout(() => {
    if (unassignedCarsRef.value) {
      // Get the actual DOM element, handling Vue 3 ref structure
      let element = unassignedCarsRef.value.$el

      // If it's a text node, try to get the parent element
      if (element && element.nodeType === Node.TEXT_NODE) {
        element = element.parentElement
      }

      // Alternative: try to get the element by querying the DOM
      if (!element || !element.getBoundingClientRect) {
        element =
          document.querySelector('#unassigned-cars-table') ||
          document.querySelector('.unassigned-cars-container')
      }

      if (element && element.getBoundingClientRect) {
        const yOffset = -50 // Increased offset to account for any fixed headers
        const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset
        window.scrollTo({ top: y, behavior: 'smooth' })
      }
    }
  }, 100) // Reduced delay for faster response

  // Alternative scroll method if the first one doesn't work
  setTimeout(() => {
    let element = unassignedCarsRef.value?.$el

    // If it's a text node, try to get the parent element
    if (element && element.nodeType === Node.TEXT_NODE) {
      element = element.parentElement
    }

    // Alternative: try to get the element by querying the DOM
    if (!element || !element.scrollIntoView) {
      element =
        document.querySelector('#unassigned-cars-table') ||
        document.querySelector('.unassigned-cars-container')
    }

    if (element && element.scrollIntoView) {
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest',
      })
    }
  }, 200)
}

const handleContainerCreated = async (newContainer) => {
  // Select the newly created container
  selectedLoadedContainerId.value = newContainer.id
  selectedContainerOnBoard.value = false

  // Scroll to unassigned cars table after a short delay to ensure it's rendered
  setTimeout(() => {
    if (unassignedCarsRef.value) {
      // Get the actual DOM element, handling Vue 3 ref structure
      let element = unassignedCarsRef.value.$el

      // If it's a text node, try to get the parent element
      if (element && element.nodeType === Node.TEXT_NODE) {
        element = element.parentElement
      }

      // Alternative: try to get the element by querying the DOM
      if (!element || !element.getBoundingClientRect) {
        element =
          document.querySelector('#unassigned-cars-table') ||
          document.querySelector('.unassigned-cars-container')
      }

      if (element && element.getBoundingClientRect) {
        const yOffset = -50 // Increased offset to account for any fixed headers
        const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset
        window.scrollTo({ top: y, behavior: 'smooth' })
      }
    }
  }, 100) // Reduced delay for faster response

  // Alternative scroll method if the first one doesn't work
  setTimeout(() => {
    let element = unassignedCarsRef.value?.$el

    // If it's a text node, try to get the parent element
    if (element && element.nodeType === Node.TEXT_NODE) {
      element = element.parentElement
    }

    // Alternative: try to get the element by querying the DOM
    if (!element || !element.scrollIntoView) {
      element =
        document.querySelector('#unassigned-cars-table') ||
        document.querySelector('.unassigned-cars-container')
    }

    if (element && element.scrollIntoView) {
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start',
        inline: 'nearest',
      })
    }
  }, 200)
}

const handleCarUnassigned = () => {
  // Refresh unassigned cars table when a car is unassigned
  if (unassignedCarsRef.value) {
    unassignedCarsRef.value.refreshData()
  }
  // Refresh containers table to update on board button states
  if (containersTableRef.value) {
    containersTableRef.value.fetchContainers()
  }
  // Refresh selected container on board status
  if (selectedLoadedContainerId.value) {
    refreshSelectedContainerStatus()
  }
}

const handleCarAssigned = async () => {
  // Refresh assigned cars table when a car is assigned
  if (assignedCarsRef.value) {
    await assignedCarsRef.value.refreshData()
    // Notify UnassignedCars.vue that assigned cars have been refreshed
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new Event('assigned-cars-refreshed'))
    }
  }
  // Refresh containers table to update on board button states
  if (containersTableRef.value) {
    containersTableRef.value.fetchContainers()
  }
  // Refresh selected container on board status
  if (selectedLoadedContainerId.value) {
    refreshSelectedContainerStatus()
  }
}

const handleRefreshUnassignedCars = () => {
  // Refresh unassigned cars table when a container is deleted
  console.log('Container deleted - refreshing unassigned cars')
  if (unassignedCarsRef.value) {
    unassignedCarsRef.value.refreshData()
  }
  // Also refresh assigned cars table to update the display
  if (assignedCarsRef.value) {
    assignedCarsRef.value.refreshData()
  }
}

const refreshSelectedContainerStatus = async () => {
  try {
    const result = await callApi({
      query: 'SELECT date_on_board FROM loaded_containers WHERE id = ?',
      params: [selectedLoadedContainerId.value],
    })

    if (result.success && result.data && result.data.length > 0) {
      selectedContainerOnBoard.value = !!result.data[0].date_on_board
    } else {
      selectedContainerOnBoard.value = false
    }
  } catch (err) {
    console.error('Error refreshing container on board status:', err)
    selectedContainerOnBoard.value = false
  }
}

const printLoadingRecord = async () => {
  if (!selectedLoadingId.value) {
    alert('Please select a loading record first')
    return
  }

  if (isPrintingRecord.value) return // Prevent double-click

  try {
    isPrintingRecord.value = true
    // Get the loading record details
    const loadingRecord = loadingRecords.value.find(
      (record) => record.id === selectedLoadingId.value,
    )
    if (!loadingRecord) {
      alert('Loading record not found')
      return
    }

    // Get all containers and their assigned cars for this loading
    const result = await callApi({
      query: `
        SELECT 
          lc.id as loaded_container_id,
          lc.ref_container,
          lc.so,
          lc.date_departed,
          lc.date_loaded,
          lc.date_on_board,
          lc.is_released,
          lc.note as container_note,
          c.name as container_name,
          cs.id as car_id,
          cs.vin,
          cs.price_cell,
          cn.car_name,
          clr.color,
          cl.name as client_name,
          cl.mobiles as client_mobiles,
          cl.id_no as client_id_no,
          cl.nin as client_nin,
          cl.id_copy_path
        FROM loaded_containers lc
        LEFT JOIN containers c ON lc.id_container = c.id
        LEFT JOIN cars_stock cs ON lc.id = cs.id_loaded_container
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN clients cl ON cs.id_client = cl.id
        WHERE lc.id_loading = ?
        ORDER BY lc.id, cs.id
      `,
      params: [selectedLoadingId.value],
    })

    if (!result.success) {
      alert('Failed to load container data')
      return
    }

    // Group cars by containers
    const containersData = {}
    result.data.forEach((row) => {
      if (!containersData[row.loaded_container_id]) {
        containersData[row.loaded_container_id] = {
          id: row.loaded_container_id,
          name: row.container_name,
          ref_container: row.ref_container,
          so: row.so,
          date_departed: row.date_departed,
          date_loaded: row.date_loaded,
          date_on_board: row.date_on_board,
          is_released: row.is_released,
          note: row.container_note,
          cars: [],
        }
      }
      if (row.car_id) {
        containersData[row.loaded_container_id].cars.push({
          id: row.car_id,
          vin: row.vin,
          car_name: row.car_name,
          color: row.color,
          price_cell: row.price_cell,
          client_name: row.client_name,
          client_mobiles: row.client_mobiles,
          client_id_no: row.client_id_no,
          client_nin: row.client_nin,
          id_copy_path: row.id_copy_path,
        })
      }
    })

    // Create the print content
    const printContent = generatePrintContent(loadingRecord, containersData)

    // Open new tab with print content (no automatic print dialog)
    const printWindow = window.open('', '_blank')
    printWindow.document.write(printContent)
    printWindow.document.close()
  } catch (err) {
    console.error('Error generating print content:', err)
    alert('Error generating print content')
  } finally {
    isPrintingRecord.value = false
  }
}

const generatePrintContent = (loadingRecord, containersData) => {
  const containers = Object.values(containersData)
  const totalCars = containers.reduce((sum, container) => sum + container.cars.length, 0)

  return `
    <!DOCTYPE html>
    <html>
    <head>
      <title>Loading Record #${loadingRecord.id} - Print</title>
      <style>
        body {
          font-family: Arial, sans-serif;
          margin: 20px;
          line-height: 1.4;
        }
        .header {
          text-align: center;
          margin-bottom: 30px;
          border-bottom: 2px solid #333;
          padding-bottom: 20px;
        }
        .header h1 {
          margin: 0;
          color: #333;
          font-size: 24px;
        }
        .loading-info {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 20px;
          margin-bottom: 30px;
          background: #f8f9fa;
          padding: 20px;
          border-radius: 8px;
        }
        .info-group {
          margin-bottom: 15px;
        }
        .info-label {
          font-weight: bold;
          color: #555;
          margin-bottom: 5px;
        }
        .info-value {
          color: #333;
        }
        .container-section {
          margin-bottom: 40px;
          page-break-inside: avoid;
        }
        .container-header {
          background: #e3f2fd;
          padding: 15px;
          border-radius: 8px;
          margin-bottom: 15px;
          border-left: 4px solid #2196f3;
        }
        .container-title {
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 10px 0;
          color: #1976d2;
        }
        .cars-table {
          width: 100%;
          border-collapse: collapse;
          margin-top: 10px;
        }
        .cars-table th {
          background: #f5f5f5;
          padding: 10px;
          text-align: left;
          border: 1px solid #ddd;
          font-weight: bold;
        }
        .cars-table td {
          padding: 8px 10px;
          border: 1px solid #ddd;
          vertical-align: top;
        }
        .cars-table tr:nth-child(even) {
          background: #f9f9f9;
        }
        .client-id-image {
          width: 100px;
          height: 70px;
          object-fit: cover;
          border-radius: 6px;
          border: 2px solid #ddd;
          cursor: pointer;
        }
        .client-info {
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .client-details {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }
        .client-name {
          font-weight: 500;
        }
        .client-id-no {
          font-size: 0.85rem;
          color: #666;
        }
        .client-mobile {
          font-size: 0.8rem;
          color: #1e40af;
          font-weight: 500;
          margin-top: 2px;
          display: flex;
          align-items: center;
          gap: 4px;
          background: #eff6ff;
          padding: 4px 8px;
          border-radius: 4px;
          border-left: 3px solid #3b82f6;
        }
        .client-mobile i {
          font-size: 0.7rem;
          color: #3b82f6;
        }
        .client-mobile strong {
          color: #1e40af;
          font-weight: 600;
        }
        .client-nin {
          font-size: 0.75rem;
          color: #1e40af;
          font-weight: 600;
          background: #dbeafe;
          border: 1px solid #93c5fd;
          border-radius: 4px;
          padding: 2px 4px;
          font-family: 'Courier New', monospace;
          text-align: center;
          margin-top: 2px;
          display: inline-block;
          width: fit-content;
        }
        .summary {
          margin-top: 30px;
          padding: 20px;
          background: #e8f5e8;
          border-radius: 8px;
          border-left: 4px solid #4caf50;
        }
        .summary h3 {
          margin: 0 0 15px 0;
          color: #2e7d32;
        }
        .summary-stats {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          gap: 15px;
        }
        .stat-item {
          text-align: center;
        }
        .stat-number {
          font-size: 24px;
          font-weight: bold;
          color: #2e7d32;
        }
        .stat-label {
          font-size: 12px;
          color: #666;
          text-transform: uppercase;
        }
        @media print {
          body { margin: 0; }
          .container-section { page-break-inside: avoid; }
        }
      </style>
    </head>
    <body>
      <div class="header">
        <h1>Loading Record #${loadingRecord.id}</h1>
        <p>Generated on ${new Date().toLocaleDateString()} at ${new Date().toLocaleTimeString()}</p>
      </div>

      <div class="loading-info">
        <div class="info-group">
          <div class="info-label">Operation Date:</div>
          <div class="info-value">${loadingRecord.date_loading ? new Date(loadingRecord.date_loading).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Shipping Line:</div>
          <div class="info-value">${loadingRecord.shipping_line_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Freight:</div>
          <div class="info-value">${loadingRecord.freight ? `$${loadingRecord.freight}` : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Loading Port:</div>
          <div class="info-value">${loadingRecord.loading_port_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Discharge Port:</div>
          <div class="info-value">${loadingRecord.discharge_port_name || 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">EDD:</div>
          <div class="info-value">${loadingRecord.EDD ? new Date(loadingRecord.EDD).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Loaded Date:</div>
          <div class="info-value">${loadingRecord.date_loaded ? new Date(loadingRecord.date_loaded).toLocaleDateString() : 'Not set'}</div>
        </div>
        <div class="info-group">
          <div class="info-label">Notes:</div>
          <div class="info-value">${loadingRecord.note || 'No notes'}</div>
        </div>
      </div>

      ${containers
        .map(
          (container) => `
        <div class="container-section">
          <div class="container-header">
            <div class="container-title">
              Container: ${container.name || 'Unnamed'} 
              ${container.ref_container ? `(${container.ref_container})` : ''}
              ${container.so ? ` - SO: ${container.so}` : ''}
              ${container.is_released ? ' - RELEASED' : ''}
              - Loading #${loadingRecord.id} - Container #${container.id}
            </div>
          </div>
          
          ${
            container.cars.length > 0
              ? `
            <table class="cars-table">
              <thead>
                <tr>
                  <th>Car ID</th>
                  <th>Car Name</th>
                  <th>Color</th>
                  <th>VIN</th>
                  <th>Client</th>
                </tr>
              </thead>
              <tbody>
                ${container.cars
                  .map(
                    (car) => `
                  <tr>
                    <td>#${car.id}</td>
                    <td>${car.car_name || 'N/A'}</td>
                    <td>${car.color || 'N/A'}</td>
                    <td>${car.vin || 'N/A'}</td>
                    <td>
                      <div class="client-info">
                        ${
                          car.id_copy_path
                            ? `
                          <img 
                            src="${getFileUrl(car.id_copy_path)}" 
                            alt="Client ID" 
                            class="client-id-image"
                            onerror="this.style.display='none'"
                          />
                        `
                            : ''
                        }
                        <div class="client-details">
                          <div class="client-name">${car.client_name || 'N/A'}</div>
                          ${car.client_mobiles && car.client_mobiles !== 'please provide mobile' ? `<div class="client-mobile"><i class="fas fa-phone"></i> <strong>Mobile:</strong> ${car.client_mobiles}</div>` : ''}
                          <div class="client-id-no">${car.client_id_no || 'No ID'}</div>
                          ${car.client_nin ? `<div class="client-nin">${car.client_nin}</div>` : ''}
                        </div>
                      </div>
                    </td>
                  </tr>
                `,
                  )
                  .join('')}
              </tbody>
            </table>
          `
              : '<p style="text-align: center; color: #666; font-style: italic;">No cars assigned to this container</p>'
          }
        </div>
      `,
        )
        .join('')}

      <div class="summary">
        <h3>Summary</h3>
        <div class="summary-stats">
          <div class="stat-item">
            <div class="stat-number">${containers.length}</div>
            <div class="stat-label">Containers</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${totalCars}</div>
            <div class="stat-label">Total Cars</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${containers.filter((c) => c.date_on_board).length}</div>
            <div class="stat-label">On Board</div>
          </div>
          <div class="stat-item">
            <div class="stat-number">${containers.filter((c) => !c.date_on_board).length}</div>
            <div class="stat-label">Pending</div>
          </div>
        </div>
      </div>
    </body>
    </html>
  `
}

// Add task handling methods
const openTaskForLoading = (record) => {
  if (isProcessingTask.value) return // Prevent double-click
  console.log('openTaskForLoading called with record:', record)
  console.log('Current showTaskForm value:', showTaskForm.value)
  selectedLoadingForTask.value = record
  showTaskForm.value = true
  console.log('After setting showTaskForm to true:', showTaskForm.value)
  console.log('selectedLoadingForTask value:', selectedLoadingForTask.value)
}

const handleTaskCreated = () => {
  isProcessingTask.value = false
  showTaskForm.value = false
  // Don't set selectedLoadingForTask to null to avoid prop validation errors
  // Optionally refresh data if needed
}

onMounted(() => {
  Promise.all([refreshData(), fetchReferenceData()])
})
</script>

<style scoped>
.loading-table-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
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

.header-actions {
  display: flex;
  gap: 8px;
}

.toggle-filters-btn {
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
  background-color: #f8fafc;
  color: #374151;
  border: 1px solid #d1d5db;
}

.toggle-filters-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
}

.toggle-filters-btn.active {
  background-color: #dbeafe;
  color: #1d4ed8;
  border: 1px solid #93c5fd;
}

.add-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.add-btn:hover:not(:disabled) {
  background: #2980b9;
}

.add-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.print-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.print-btn:hover:not(:disabled) {
  background: #2980b9;
}

.print-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #2980b9;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.refresh-btn.processing {
  background: #95a5a6;
}

.loading-indicator {
  font-size: 0.8rem;
}

.table-wrapper {
  position: relative;
  max-height: 200px;
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
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.empty-state i {
  color: #d1d5db;
}

.table-content {
  overflow-x: auto;
}

.loading-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 1rem;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.loading-table th,
.loading-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.loading-table th {
  background-color: #f8f9fa;
  font-weight: 600;
  color: #495057;
}

.sortable-header {
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s ease;
}

.sortable-header:hover {
  background-color: #e9ecef;
}

.sortable-header i {
  margin-left: 5px;
  font-size: 0.8em;
}

.sort-inactive {
  color: #adb5bd;
}

.loading-table tr:hover {
  background-color: #f8f9fa;
}

.loading-table td {
  vertical-align: middle;
}

.loading-table .actions {
  display: flex;
  gap: 8px;
  justify-content: flex-start;
}

.loading-table .btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: all 0.2s ease;
}

.loading-table .btn-edit {
  background-color: #007bff;
  color: white;
}

.loading-table .btn-edit:hover {
  background-color: #0056b3;
}

.loading-table .btn-delete {
  background-color: #dc3545;
  color: white;
}

.loading-table .btn-delete:hover {
  background-color: #c82333;
}

.loading-table .btn-print {
  background-color: #28a745;
  color: white;
}

.loading-table .btn-print:hover {
  background-color: #218838;
}

.loading-table .btn-details {
  background-color: #17a2b8;
  color: white;
}

.loading-table .btn-details:hover {
  background-color: #138496;
}

.loading-table .status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
}

.loading-table .status-loaded {
  background-color: #d4edda;
  color: #155724;
}

.loading-table .status-pending {
  background-color: #fff3cd;
  color: #856404;
}

.loading-table .status-in-progress {
  background-color: #cce5ff;
  color: #004085;
}

.loading-table .empty-message {
  text-align: center;
  padding: 2rem;
  color: #6c757d;
  font-style: italic;
}

.loading-table .loading-message {
  text-align: center;
  padding: 2rem;
  color: #6c757d;
}

.loading-table .error-message {
  text-align: center;
  padding: 2rem;
  color: #dc3545;
  background-color: #f8d7da;
  border-radius: 4px;
  margin: 1rem 0;
}

/* Responsive design */
@media (max-width: 768px) {
  .loading-table {
    font-size: 0.875rem;
  }

  .loading-table th,
  .loading-table td {
    padding: 8px 6px;
  }

  .loading-table .actions {
    flex-direction: column;
    gap: 4px;
  }

  .loading-table .btn {
    padding: 4px 8px;
    font-size: 0.75rem;
  }
}

.selected-row {
  background-color: #e3f2fd !important;
  border-left: 3px solid #3498db;
}

.table-row {
  cursor: pointer;
}

.id-cell {
  font-weight: 600;
  color: #3498db;
  width: 60px;
}

.date-cell {
  width: 100px;
  white-space: nowrap;
}

.shipping-line-cell {
  width: 120px;
}

.freight-cell {
  width: 80px;
  text-align: right;
  font-weight: 600;
}

.port-cell {
  width: 120px;
}

.notes-cell {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions-cell {
  width: 100px;
  text-align: right;
}

.action-buttons {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
}

.action-btn {
  padding: 4px 8px;
  background: none;
  border: none;
  cursor: pointer;
  transition: color 0.2s ease;
}

.action-btn:hover {
  color: #3498db;
}

.task-btn {
  color: #6b7280;
}

.task-btn:hover {
  color: #f59e0b;
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

.record-count {
  color: #6b7280;
}

.view-all-link a {
  color: #3498db;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color 0.2s ease;
}

.view-all-link a:hover {
  color: #2980b9;
  text-decoration: underline;
}

/* Scrollbar styling */
.table-wrapper::-webkit-scrollbar {
  width: 8px;
}

.table-wrapper::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.table-wrapper::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.table-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* Responsive design */
@media (max-width: 768px) {
  .table-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }

  .table-footer {
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }

  .loading-table {
    font-size: 0.8rem;
  }

  .loading-table th,
  .loading-table td {
    padding: 8px 4px;
  }
}

/* Dialog styling */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e5e7eb;
  background-color: #f8fafc;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
}

.dialog-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1.1rem;
}

.dialog-header h3 i {
  color: #3498db;
}

.close-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 1.2rem;
  color: #6b7280;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover:not(:disabled) {
  background-color: #f3f4f6;
  color: #374151;
}

.close-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dialog-content {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-group label i {
  color: #6b7280;
  font-size: 0.8rem;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  background-color: white;
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
  cursor: not-allowed;
}

.form-group textarea {
  resize: vertical;
  min-height: 80px;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.cancel-btn,
.save-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.cancel-btn {
  background-color: #f3f4f6;
  color: #374151;
}

.cancel-btn:hover:not(:disabled) {
  background-color: #e5e7eb;
  color: #1f2937;
}

.cancel-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.save-btn {
  background-color: #3498db;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.save-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.save-btn.processing {
  background-color: #95a5a6;
}

.loading-indicator {
  font-size: 0.8rem;
}

.add-first-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  cursor: pointer;
  color: #3498db;
  font-size: 0.9rem;
  text-decoration: underline;
  transition: color 0.2s ease;
}

.add-first-btn:hover {
  color: #2980b9;
}

/* Responsive design for dialog */
@media (max-width: 768px) {
  .dialog {
    width: 95%;
    margin: 20px;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .dialog-actions {
    flex-direction: column;
  }

  .cancel-btn,
  .save-btn {
    width: 100%;
    justify-content: center;
  }
}

.input-with-button {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-with-button select {
  flex: 1;
}

.add-item-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 38px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 0.8rem;
}

.add-item-btn:hover:not(:disabled) {
  background-color: #2980b9;
}

.add-item-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.quick-add-dialog {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 400px;
  max-height: 90vh;
  overflow-y: auto;
}

/* Filters Section Styles */
.filters-section {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin: 16px 0;
  padding: 20px;
  border: 1px solid #e5e7eb;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.filters-header h4 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1f2937;
  font-size: 1rem;
  font-weight: 600;
}

.filters-header h4 i {
  color: #6b7280;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.clear-filters-btn:hover {
  background: linear-gradient(135deg, #d97706, #b45309);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.clear-filters-btn:disabled {
  background: linear-gradient(135deg, #9ca3af, #6b7280);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group label {
  font-size: 0.85rem;
  font-weight: 500;
  color: #374151;
}

.filter-group input,
.filter-group select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 0.9rem;
  transition: border-color 0.2s ease;
  background-color: white;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

/* Responsive design for filters */
@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }

  .filters-header {
    flex-direction: column;
    gap: 12px;
    align-items: flex-start;
  }
}

/* Beautiful Filter Category Groups */
.filter-group-category {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 2px solid transparent;
  transition: all 0.3s ease;
  overflow: hidden;
}

.filter-group-category:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.category-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  font-weight: 600;
  font-size: 1rem;
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.category-header i {
  font-size: 1.1rem;
}

.category-filters {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Search & General Filters - Blue */
.search-filters {
  border-color: #3b82f6;
}

.search-filters .category-header {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
}

/* Vehicle Information Filters - Green */
.vehicle-filters {
  border-color: #10b981;
}

.vehicle-filters .category-header {
  background: linear-gradient(135deg, #10b981, #059669);
}

/* Client & Business Filters - Purple */
.client-filters {
  border-color: #8b5cf6;
}

.client-filters .category-header {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
}

/* Container & Storage Filters - Orange */
.container-filters {
  border-color: #f59e0b;
}

.container-filters .category-header {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

/* Shipping & Logistics Filters - Teal */
.shipping-filters {
  border-color: #14b8a6;
}

.shipping-filters .category-header {
  background: linear-gradient(135deg, #14b8a6, #0d9488);
}

/* Date & Time Filters - Red */
.date-filters {
  border-color: #ef4444;
}

.date-filters .category-header {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

/* Enhanced Filter Grid Layout */
.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

/* Enhanced Filter Group Styling */
.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 6px;
}

.filter-group label i {
  color: #6b7280;
  font-size: 0.8rem;
}

.filter-group input,
.filter-group select {
  padding: 10px 14px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  background-color: white;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  transform: translateY(-1px);
}

/* Responsive adjustments for filter categories */
@media (max-width: 1200px) {
  .filters-grid {
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }
}

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }

  .filter-group-category {
    margin-bottom: 16px;
  }

  .category-header {
    padding: 14px 16px;
    font-size: 0.95rem;
  }

  .category-filters {
    padding: 16px;
  }
}

/* Additional filters section styling */
.filters-grid:last-of-type {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.filters-grid:last-of-type .filter-group {
  margin-bottom: 8px;
}

/* Highlight effect for newly added records */
.highlight-new-record {
  animation: highlightNewRecord 3s ease-in-out;
}

@keyframes highlightNewRecord {
  0% {
    background-color: #d4edda;
    border-left: 4px solid #28a745;
  }
  50% {
    background-color: #d1ecf1;
    border-left: 4px solid #17a2b8;
  }
  100% {
    background-color: transparent;
    border-left: none;
  }
}

/* Task Modal Styles */
.task-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1001;
  backdrop-filter: blur(4px);
}

.task-modal {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.filter-actions {
  display: flex;
  gap: 16px;
  margin-top: 24px;
  justify-content: center;
  padding: 20px;
  background: linear-gradient(135deg, #f8fafc, #f1f5f9);
  border-radius: 12px;
  border: 1px solid #e2e8f0;
}

.apply-filter-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.apply-filter-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #059669, #047857);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(16, 185, 129, 0.3);
}

.apply-filter-btn:disabled {
  background: linear-gradient(135deg, #9ca3af, #6b7280);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.clear-filters-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #d97706, #b45309);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(245, 158, 11, 0.3);
}

.clear-filters-btn:disabled {
  background: linear-gradient(135deg, #9ca3af, #6b7280);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}
</style>
