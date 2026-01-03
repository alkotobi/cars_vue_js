<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useEnhancedI18n } from '@/composables/useI18n'
import { useApi } from '../../composables/useApi'
import SendSelectionForm from './SendSelectionForm.vue'
import CarStockPrintOptions from './CarStockPrintOptions.vue'
import SelectionChatButton from './SelectionChatButton.vue'

const { t } = useEnhancedI18n()
const { callApi, getFileUrl, getAssets, loadLetterhead } = useApi()
const letterHeadUrl = ref(null)
const user = ref(null)

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'selection-loaded'])

const selections = ref([])
const loading = ref(false)
const error = ref(null)
const selectedSelection = ref(null)
const showSendForm = ref(false)
const selectedCarIdsForSend = ref([])
const selectedSelectionIdForSend = ref(null)
const deletingSelectionId = ref(null)
const comments = ref({}) // Object mapping selection_id to array of comments
const loadingComments = ref({}) // Object mapping selection_id to loading state
const newComment = ref({}) // Object mapping selection_id to new comment text
const showComments = ref({}) // Object mapping selection_id to show/hide comments
const isAddingComment = ref({}) // Object mapping selection_id to adding state
const changingStatusId = ref(null) // Track which selection is having its status changed
const showingCars = ref({}) // Object mapping selection_id to show/hide cars table
const selectionCars = ref({}) // Object mapping selection_id to array of cars
const loadingCars = ref({}) // Object mapping selection_id to loading state
const allSelections = ref([]) // Store all selections for filtering
const filters = ref({
  search: '',
  status: '',
  priority: '',
  dateFrom: '',
  dateTo: '',
})
const showFilters = ref(false)

const statusColors = {
  pending: '#f59e0b',
  in_progress: '#3b82f6',
  completed: '#10b981',
  cancelled: '#ef4444',
}

const priorityColors = {
  low: '#6b7280',
  medium: '#3b82f6',
  high: '#f59e0b',
  urgent: '#ef4444',
}

watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      loadSelections()
    }
  },
)

const loadSelections = async () => {
  loading.value = true
  error.value = null

  try {
    if (!user.value) {
      const userStr = localStorage.getItem('user')
      user.value = userStr ? JSON.parse(userStr) : null
    }

    if (!user.value || !user.value.id) {
      throw new Error('User not found')
    }

    const userId = parseInt(user.value.id)

    const result = await callApi({
      query: `
        SELECT DISTINCT
          cs.id,
          cs.name,
          cs.description,
          cs.selection_data,
          cs.status,
          cs.priority,
          cs.created_at,
          cs.assigned_to_team,
          cs.job_done_on,
          cs.due_date,
          cs.deadline,
          cs.status_changed_at,
          cs.previous_status,
          t.name as team_name,
          u.username as creator_name,
          u2.username as sender_name,
          cs.owned_by
        FROM car_selections cs
        LEFT JOIN teams t ON cs.assigned_to_team = t.id
        LEFT JOIN users u ON cs.user_create_selection = u.id
        LEFT JOIN users u2 ON cs.sent_by_user_id = u2.id
        LEFT JOIN team_members tm ON cs.assigned_to_team = tm.team_id AND tm.user_id = ?
        WHERE (cs.owned_by IS NOT NULL AND JSON_CONTAINS(cs.owned_by, ?))
           OR (cs.assigned_to_team IS NOT NULL AND tm.user_id IS NOT NULL)
        ORDER BY cs.created_at DESC
      `,
      params: [userId, JSON.stringify(userId)],
    })

    if (result.success) {
      const mappedData = result.data.map((sel) => ({
        ...sel,
        selection_data: sel.selection_data ? JSON.parse(sel.selection_data) : [],
        owned_by: sel.owned_by ? JSON.parse(sel.owned_by) : [],
      }))
      allSelections.value = mappedData
      selections.value = mappedData
    } else {
      error.value = result.error || t('carStock.failed_to_load_selections')
    }
  } catch (err) {
    error.value = err.message || t('carStock.error_loading_selections')
  } finally {
    loading.value = false
  }
}

const getCarCount = (selection) => {
  if (!selection.selection_data) return 0
  try {
    const data =
      typeof selection.selection_data === 'string'
        ? JSON.parse(selection.selection_data)
        : selection.selection_data
    return Array.isArray(data) ? data.length : 0
  } catch {
    return 0
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString()
}

const handleLoadSelection = (selection) => {
  emit('selection-loaded', {
    id: selection.id,
    carIds: selection.selection_data || [],
  })
  emit('close')
}

const handleClose = () => {
  selectedSelection.value = null
  showSendForm.value = false
  emit('close')
}

const handleSendSelection = (selection) => {
  selectedCarIdsForSend.value = selection.selection_data || []
  selectedSelectionIdForSend.value = selection.id // Pass the selection ID to update existing selection
  showSendForm.value = true
}

const handleSelectionSent = () => {
  showSendForm.value = false
  selectedCarIdsForSend.value = []
  selectedSelectionIdForSend.value = null
  loadSelections()
  alert(t('carStock.selection_sent_successfully') || 'Selection sent successfully')
}

const handleDeleteSelection = async (selection) => {
  if (
    !confirm(
      t('carStock.confirm_remove_selection') ||
        `Are you sure you want to remove "${selection.name}" from your list?`,
    )
  ) {
    return
  }

  deletingSelectionId.value = selection.id

  try {
    const userStr = localStorage.getItem('user')
    const user = userStr ? JSON.parse(userStr) : null

    if (!user || !user.id) {
      throw new Error('User not found')
    }

    const userId = parseInt(user.id)

    // Get current owned_by array
    const currentOwnedBy = Array.isArray(selection.owned_by) ? selection.owned_by : []

    // Remove current user from owned_by
    const updatedOwnedBy = currentOwnedBy.filter((id) => parseInt(id) !== userId)

    // If no users left, delete the selection entirely
    if (updatedOwnedBy.length === 0) {
      const deleteResult = await callApi({
        query: 'DELETE FROM car_selections WHERE id = ?',
        params: [selection.id],
      })

      if (deleteResult.success) {
        await loadSelections()
        alert(t('carStock.selection_deleted_successfully') || 'Selection deleted successfully')
      } else {
        alert(
          deleteResult.error ||
            t('carStock.failed_to_delete_selection') ||
            'Failed to delete selection',
        )
      }
    } else {
      // Update owned_by to remove current user
      const result = await callApi({
        query: 'UPDATE car_selections SET owned_by = ? WHERE id = ?',
        params: [JSON.stringify(updatedOwnedBy), selection.id],
      })

      if (result.success) {
        await loadSelections()
        alert(t('carStock.selection_removed_successfully') || 'Selection removed from your list')
      } else {
        alert(
          result.error || t('carStock.failed_to_remove_selection') || 'Failed to remove selection',
        )
      }
    }
  } catch (err) {
    alert(err.message || t('carStock.error_removing_selection') || 'Error removing selection')
  } finally {
    deletingSelectionId.value = null
  }
}

const loadComments = async (selectionId) => {
  if (loadingComments.value[selectionId]) return

  loadingComments.value[selectionId] = true
  try {
    const result = await callApi({
      query: `
        SELECT sc.id, sc.comment, sc.created_at, sc.updated_at, u.username, u.id as user_id
        FROM selection_comments sc
        LEFT JOIN users u ON sc.user_id = u.id
        WHERE sc.selection_id = ?
        ORDER BY sc.created_at ASC
      `,
      params: [selectionId],
    })

    if (result.success) {
      comments.value[selectionId] = result.data || []
    }
  } catch (err) {
    console.error('Error loading comments:', err)
  } finally {
    loadingComments.value[selectionId] = false
  }
}

const toggleComments = (selectionId) => {
  if (!showComments.value[selectionId]) {
    showComments.value[selectionId] = true
    if (!comments.value[selectionId]) {
      loadComments(selectionId)
    }
  } else {
    showComments.value[selectionId] = false
  }
}

const addComment = async (selectionId) => {
  const commentText = (newComment.value[selectionId] || '').trim()
  if (!commentText) return

  isAddingComment.value[selectionId] = true
  try {
    const userStr = localStorage.getItem('user')
    const user = userStr ? JSON.parse(userStr) : null

    if (!user || !user.id) {
      throw new Error('User not found')
    }

    const result = await callApi({
      query: 'INSERT INTO selection_comments (selection_id, user_id, comment) VALUES (?, ?, ?)',
      params: [selectionId, user.id, commentText],
    })

    if (result.success) {
      newComment.value[selectionId] = ''
      await loadComments(selectionId)
    } else {
      alert(result.error || t('carStock.failed_to_add_comment') || 'Failed to add comment')
    }
  } catch (err) {
    alert(err.message || t('carStock.error_adding_comment') || 'Error adding comment')
  } finally {
    isAddingComment.value[selectionId] = false
  }
}

const isDeadlinePassed = (deadline) => {
  if (!deadline) return false
  return new Date(deadline) < new Date()
}

const handleStatusChange = async (selection, newStatus) => {
  if (selection.status === newStatus) return

  changingStatusId.value = selection.id

  try {
    const result = await callApi({
      query: `
        UPDATE car_selections 
        SET 
          status = ?,
          previous_status = ?,
          status_changed_at = NOW()
        WHERE id = ?
      `,
      params: [newStatus, selection.status, selection.id],
    })

    if (result.success) {
      await loadSelections()
      alert(t('carStock.status_updated_successfully') || 'Status updated successfully')
    } else {
      alert(result.error || t('carStock.failed_to_update_status') || 'Failed to update status')
    }
  } catch (err) {
    alert(err.message || t('carStock.error_updating_status') || 'Error updating status')
  } finally {
    changingStatusId.value = null
  }
}

const getAvailableStatuses = (currentStatus) => {
  const allStatuses = ['pending', 'in_progress', 'completed', 'cancelled']
  return allStatuses.filter((status) => status !== currentStatus)
}

const loadSelectionCars = async (selection) => {
  const selectionId = selection.id
  let carIds = []

  try {
    if (selection.selection_data) {
      carIds =
        typeof selection.selection_data === 'string'
          ? JSON.parse(selection.selection_data)
          : selection.selection_data
    }
  } catch (e) {
    console.error('Error parsing selection_data:', e)
    carIds = []
  }

  if (!carIds || carIds.length === 0) {
    selectionCars.value[selectionId] = []
    return
  }

  if (loadingCars.value[selectionId]) return

  loadingCars.value[selectionId] = true

  try {
    const placeholders = carIds.map(() => '?').join(',')
    const result = await callApi({
      query: `
        SELECT 
          cs.id,
          cn.car_name,
          clr.color,
          clr.hexa,
          cs.vin,
          cs.price_cell,
          cs.cfr_da,
          cs.freight,
          cs.rate,
          cs.notes,
          c.name as client_name,
          c.id_no as client_id_no,
          c.id_copy_path as client_id_picture,
          lp.loading_port,
          dp.discharge_port,
          cs.container_ref,
          w.warhouse_name as warehouse_name,
          bb.bill_ref as buy_bill_ref,
          bb.date_buy,
          sb.bill_ref as sell_bill_ref,
          sb.date_sell,
          cs.export_lisence_ref,
          CASE
            WHEN cs.id_sell IS NOT NULL THEN 'Sold'
            ELSE 'Available'
          END as status
        FROM cars_stock cs
        LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
        LEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id
        LEFT JOIN sell_bill sb ON cs.id_sell = sb.id
        LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
        LEFT JOIN colors clr ON cs.id_color = clr.id
        LEFT JOIN clients c ON cs.id_client = c.id
        LEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id
        LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
        LEFT JOIN warehouses w ON cs.id_warehouse = w.id
        WHERE cs.id IN (${placeholders})
        ORDER BY cs.id ASC
      `,
      params: carIds,
    })

    if (result.success) {
      selectionCars.value[selectionId] = result.data || []
    } else {
      console.error('Failed to load selection cars:', result.error)
      selectionCars.value[selectionId] = []
    }
  } catch (err) {
    console.error('Error loading selection cars:', err)
    selectionCars.value[selectionId] = []
  } finally {
    loadingCars.value[selectionId] = false
  }
}

const toggleShowCars = async (selection) => {
  const selectionId = selection.id

  if (showingCars.value[selectionId]) {
    // Hide cars
    showingCars.value[selectionId] = false
  } else {
    // Show cars for this selection
    showingCars.value[selectionId] = true

    // Load cars if not already loaded
    if (!selectionCars.value[selectionId]) {
      await loadSelectionCars(selection)
    }
  }
}

const getTextColor = (backgroundColor) => {
  if (!backgroundColor) return '#374151'

  // Convert hex to RGB
  const hex = backgroundColor.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)

  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255

  // Return white text for dark backgrounds, black text for light backgrounds
  return luminance > 0.5 ? '#000000' : '#ffffff'
}

const applyFilters = () => {
  let filtered = [...allSelections.value]

  // Search filter (name or description)
  if (filters.value.search.trim()) {
    const searchTerm = filters.value.search.trim().toLowerCase()
    filtered = filtered.filter((sel) => {
      const name = (sel.name || '').toLowerCase()
      const description = (sel.description || '').toLowerCase()
      return name.includes(searchTerm) || description.includes(searchTerm)
    })
  }

  // Status filter
  if (filters.value.status) {
    filtered = filtered.filter((sel) => sel.status === filters.value.status)
  }

  // Priority filter
  if (filters.value.priority) {
    filtered = filtered.filter((sel) => sel.priority === filters.value.priority)
  }

  // Date range filter
  if (filters.value.dateFrom) {
    filtered = filtered.filter((sel) => {
      const createdDate = new Date(sel.created_at)
      return createdDate >= new Date(filters.value.dateFrom)
    })
  }

  if (filters.value.dateTo) {
    filtered = filtered.filter((sel) => {
      const createdDate = new Date(sel.created_at)
      const toDate = new Date(filters.value.dateTo)
      toDate.setHours(23, 59, 59, 999) // Include the entire end date
      return createdDate <= toDate
    })
  }

  selections.value = filtered
}

const resetFilters = () => {
  filters.value = {
    search: '',
    status: '',
    priority: '',
    dateFrom: '',
    dateTo: '',
  }
  selections.value = [...allSelections.value]
}

const toggleFilters = () => {
  showFilters.value = !showFilters.value
}

// Print options modal state
const showPrintOptions = ref(false)
const selectedSelectionForPrint = ref(null)

const handlePrintSelection = async (selection) => {
  // Load cars for this selection if not already loaded
  if (!selectionCars.value[selection.id]) {
    await loadSelectionCars(selection)
  }

  const cars = selectionCars.value[selection.id] || []

  if (cars.length === 0) {
    alert(t('carStock.no_cars_in_selection') || 'No cars in this selection')
    return
  }

  selectedSelectionForPrint.value = selection
  showPrintOptions.value = true
}

const handlePrintWithOptions = async (printData) => {
  const { columns, cars, subject, coreContent, groupBy } = printData
  const selection = selectedSelectionForPrint.value

  if (!selection) return

  try {
    // Load assets if not already loaded
    if (!letterHeadUrl.value) {
      try {
        letterHeadUrl.value = await loadLetterhead()
      } catch (err) {
        console.error('Failed to load assets:', err)
      }
    }

    // Get user info
    let currentUser = user.value
    if (!currentUser) {
      const userStr = localStorage.getItem('user')
      currentUser = userStr ? JSON.parse(userStr) : null
    }

    // Generate REF number
    const generateRef = () => {
      const userName = currentUser?.name || currentUser?.username || 'USR'
      const userPrefix = userName.substring(0, 3).toUpperCase()
      const buttonPrefix = 'PS' // Print Selection
      const now = new Date()
      const day = now.getDate().toString().padStart(2, '0')
      const month = (now.getMonth() + 1).toString().padStart(2, '0')
      const year = now.getFullYear().toString().slice(-2)
      const dateStr = `${day}${month}${year}`
      const storageKey = `ref_sequence_print_selection_${dateStr}`
      let sequence = parseInt(localStorage.getItem(storageKey) || '0') + 1
      localStorage.setItem(storageKey, sequence.toString())
      const sequenceStr = sequence.toString().padStart(3, '0')
      return `${userPrefix}${buttonPrefix}${dateStr}-${sequenceStr}`
    }

    const title =
      subject ||
      `${t('carStock.selection') || 'Selection'}: ${selection.name} (ID: ${selection.id})`

    const contentBeforeTable = `
      <div style="margin: 20px 0; padding: 15px; background-color: #f8f9fa; border-left: 4px solid #3b82f6; border-radius: 4px;">
        <p><strong>${t('carStock.selection_name') || 'Selection Name'}:</strong> ${selection.name}</p>
        ${selection.description ? `<p><strong>${t('carStock.description') || 'Description'}:</strong></p><div style="white-space: pre-wrap; margin-left: 20px; margin-top: 5px; margin-bottom: 10px;">${selection.description}</div>` : ''}
        <p><strong>${t('carStock.status') || 'Status'}:</strong> ${t(`carStock.status_${selection.status}`) || selection.status}</p>
        <p><strong>${t('carStock.priority') || 'Priority'}:</strong> ${t(`carStock.priority_${selection.priority}`) || selection.priority}</p>
        ${selection.deadline ? `<p><strong>${t('carStock.deadline') || 'Deadline'}:</strong> ${formatDate(selection.deadline)}</p>` : ''}
        <p><strong>${t('carStock.created_at') || 'Created'}:</strong> ${formatDate(selection.created_at)}</p>
        ${coreContent ? `<div style="margin-top: 15px; padding: 10px; background-color: #eff6ff; border-left: 3px solid #3b82f6; border-radius: 4px; white-space: pre-wrap;">${coreContent}</div>` : ''}
      </div>
    `

    const contentAfterTable = `
      <p><strong>${t('carStock.total_cars') || 'Total Cars'}:</strong> ${cars.length}</p>
      <p><strong>${t('carStock.report_type') || 'Report Type'}:</strong> ${t('carStock.selection_report') || 'Selection Report'}</p>
    `

    // Create print window
    const printWindow = window.open('', '_blank', 'width=800,height=600')

    const reportHTML = `
      <!DOCTYPE html>
      <html>
      <head>
        <title>${title}</title>
        <style>
          body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
          .print-report { max-width: 210mm; margin: 0 auto; }
          .report-header { text-align: center; margin-bottom: 20px; }
          .letter-head { max-width: 100%; height: auto; max-height: 120px; }
          .report-date { text-align: right; margin-bottom: 20px; font-size: 14px; color: #666; float: right; clear: both; width: 100%; }
          .report-ref { text-align: right; margin-bottom: 20px; font-size: 14px; color: #666; float: right; clear: both; width: 100%; }
          .report-title { text-align: center; margin-bottom: 30px; font-size: 24px; font-weight: bold; color: #333; text-transform: uppercase; clear: both; }
          .content-before-table { margin-bottom: 20px; line-height: 1.6; font-size: 14px; }
          .table-container { margin-bottom: 20px; overflow-x: auto; }
          .report-table { width: 100%; border-collapse: collapse; font-size: 12px; }
          .table-header { background-color: #f8f9fa; border: 1px solid #dee2e6; padding: 8px 12px; text-align: left; font-weight: bold; color: #495057; white-space: nowrap; }
          .table-row:nth-child(even) { background-color: #f8f9fa; }
          .table-cell { border: 1px solid #dee2e6; padding: 8px 12px; text-align: left; vertical-align: top; }
          .group-header-row { background-color: #e0f2fe; }
          .group-header-cell { background-color: #e0f2fe; border: 2px solid #0ea5e9; padding: 10px 12px; font-weight: bold; color: #0c4a6e; }
          .content-after-table { margin-top: 20px; line-height: 1.6; font-size: 14px; }
          @media print { body { padding: 0; margin: 0; } .report-table { font-size: 10px; } .table-header, .table-cell { padding: 6px 8px; } .letter-head { max-height: 80px; } .report-title { font-size: 20px; margin-bottom: 20px; } }
        </style>
      </head>
      <body>
        <div class="print-report">
          <div class="report-header">
            <img 
              src="${letterHeadUrl.value || ''}" 
              alt="Letter Head" 
              class="letter-head"
              style="max-width: 100%; height: auto; max-height: 120px;"
              onerror="this.style.display='none'"
            />
          </div>
          <div class="report-date">
            <span><strong>${t('carStock.date') || 'Date'}:</strong> ${new Date().toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</span>
          </div>
          <div class="report-ref">
            <span><strong>REF:</strong> ${generateRef()}</span>
          </div>
          <h1 class="report-title">${title}</h1>
          ${contentBeforeTable}
          <div class="table-container">
            <table class="report-table">
              <thead>
                <tr>
                  ${columns.map((col) => `<th class="table-header">${col.label}</th>`).join('')}
                </tr>
              </thead>
              <tbody>
                ${(() => {
                  const groupByValue = groupBy || ''
                  if (!groupByValue || groupByValue === '') {
                    // No grouping - render all cars normally
                    return cars
                      .map((car) => {
                        const cells = columns.map((col) => {
                          const value = car[col.key]
                          let displayValue = '-'

                          // Handle calculated values first
                          if (col.key === 'cfr_usd') {
                            if (value) {
                              displayValue = '$' + parseFloat(value).toLocaleString()
                            } else if (car.price_cell && car.freight) {
                              const fob = parseFloat(car.price_cell) || 0
                              const freight = parseFloat(car.freight) || 0
                              displayValue = '$' + (fob + freight).toLocaleString()
                            }
                          } else if (col.key === 'cfr_dza') {
                            if (value) {
                              displayValue = parseFloat(value).toLocaleString()
                            } else if (car.cfr_da) {
                              displayValue = parseFloat(car.cfr_da).toLocaleString()
                            } else if (car.price_cell && car.freight && car.rate) {
                              const fob = parseFloat(car.price_cell) || 0
                              const freight = parseFloat(car.freight) || 0
                              const rate = parseFloat(car.rate) || 0
                              const cfrUsd = fob + freight
                              const cfrDza = cfrUsd * rate
                              displayValue = cfrDza.toLocaleString()
                            }
                          } else if (value !== null && value !== undefined) {
                            if (col.key === 'client_id_picture' && value) {
                              displayValue = `<img src="${getFileUrl(value)}" alt="Client ID" style="max-width: 100px; max-height: 60px; object-fit: contain;" />`
                            } else if (col.key.includes('date') && value) {
                              displayValue = new Date(value).toLocaleDateString()
                            } else if (
                              ['price_cell', 'freight', 'rate', 'cfr_da'].includes(col.key) &&
                              value
                            ) {
                              if (col.key === 'price_cell' || col.key === 'freight') {
                                displayValue = '$' + parseFloat(value).toLocaleString()
                              } else {
                                displayValue = parseFloat(value).toLocaleString()
                              }
                            } else {
                              displayValue = value.toString()
                            }
                          }
                          return `<td class="table-cell">${displayValue}</td>`
                        })
                        return `<tr class="table-row">${cells.join('')}</tr>`
                      })
                      .join('')
                  } else {
                    // Group by selected column
                    const groups = {}
                    cars.forEach((car) => {
                      const groupValue = car[groupByValue] || '(No ' + groupByValue + ')'
                      if (!groups[groupValue]) {
                        groups[groupValue] = []
                      }
                      groups[groupValue].push(car)
                    })

                    // Sort groups by group value
                    const sortedGroups = Object.keys(groups).sort()

                    return sortedGroups
                      .map((groupValue) => {
                        const groupCars = groups[groupValue]
                        const groupLabel =
                          groupByValue === 'buy_bill_ref'
                            ? 'Buy Bill Ref'
                            : groupByValue === 'container_ref'
                              ? 'Container Ref'
                              : groupByValue === 'client_name'
                                ? 'Client'
                                : groupByValue === 'loading_port'
                                  ? 'Loading Port'
                                  : groupByValue === 'discharge_port'
                                    ? 'Discharge Port'
                                    : groupByValue === 'warehouse_name'
                                      ? 'Warehouse'
                                      : groupByValue === 'status'
                                        ? 'Status'
                                        : groupByValue === 'color'
                                          ? 'Color'
                                          : groupByValue

                        return `
                          <tr class="group-header-row">
                            <td colspan="${columns.length}" class="group-header-cell">
                              <strong>${groupLabel}: ${groupValue}</strong> (${groupCars.length} car${groupCars.length === 1 ? '' : 's'})
                            </td>
                          </tr>
                          ${groupCars
                            .map((car) => {
                              const cells = columns.map((col) => {
                                const value = car[col.key]
                                let displayValue = '-'

                                // Handle calculated values first
                                if (col.key === 'cfr_usd') {
                                  if (value) {
                                    displayValue = '$' + parseFloat(value).toLocaleString()
                                  } else if (car.price_cell && car.freight) {
                                    const fob = parseFloat(car.price_cell) || 0
                                    const freight = parseFloat(car.freight) || 0
                                    displayValue = '$' + (fob + freight).toLocaleString()
                                  }
                                } else if (col.key === 'cfr_dza') {
                                  if (value) {
                                    displayValue = parseFloat(value).toLocaleString()
                                  } else if (car.cfr_da) {
                                    displayValue = parseFloat(car.cfr_da).toLocaleString()
                                  } else if (car.price_cell && car.freight && car.rate) {
                                    const fob = parseFloat(car.price_cell) || 0
                                    const freight = parseFloat(car.freight) || 0
                                    const rate = parseFloat(car.rate) || 0
                                    const cfrUsd = fob + freight
                                    const cfrDza = cfrUsd * rate
                                    displayValue = cfrDza.toLocaleString()
                                  }
                                } else if (value !== null && value !== undefined) {
                                  if (col.key === 'client_id_picture' && value) {
                                    displayValue = `<img src="${getFileUrl(value)}" alt="Client ID" style="max-width: 100px; max-height: 60px; object-fit: contain;" />`
                                  } else if (col.key.includes('date') && value) {
                                    displayValue = new Date(value).toLocaleDateString()
                                  } else if (
                                    ['price_cell', 'freight', 'rate', 'cfr_da'].includes(col.key) &&
                                    value
                                  ) {
                                    if (col.key === 'price_cell' || col.key === 'freight') {
                                      displayValue = '$' + parseFloat(value).toLocaleString()
                                    } else {
                                      displayValue = parseFloat(value).toLocaleString()
                                    }
                                  } else {
                                    displayValue = value.toString()
                                  }
                                }
                                return `<td class="table-cell">${displayValue}</td>`
                              })
                              return `<tr class="table-row">${cells.join('')}</tr>`
                            })
                            .join('')}
                        `
                      })
                      .join('')
                  }
                })()}
              </tbody>
            </table>
          </div>
          ${contentAfterTable}
        </div>
      </body>
      </html>
    `

    printWindow.document.write(reportHTML)
    printWindow.document.close()

    // Wait for content to load then print
    printWindow.onload = () => {
      printWindow.print()
      printWindow.close()
    }

    // Close the modal
    showPrintOptions.value = false
    selectedSelectionForPrint.value = null
  } catch (err) {
    console.error('Error printing selection:', err)
    alert(t('carStock.failed_to_print_selection') || 'Failed to print selection')
  }
}

const handleClosePrintOptions = () => {
  showPrintOptions.value = false
  selectedSelectionForPrint.value = null
}
</script>

<template>
  <div v-if="show" class="modal-overlay" @click.self="handleClose">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>{{ t('carStock.show_selections') }}</h3>
        <button @click="handleClose" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <!-- Filter Toolbar -->
      <div class="filter-toolbar">
        <div class="toolbar-main">
          <button @click="toggleFilters" class="filter-toggle-btn" :class="{ active: showFilters }">
            <i class="fas fa-filter"></i>
            {{ t('carStock.filters') || 'Filters' }}
            <i class="fas fa-chevron-down" :class="{ rotated: showFilters }"></i>
          </button>
          <div
            class="filter-summary"
            v-if="
              filters.search ||
              filters.status ||
              filters.priority ||
              filters.dateFrom ||
              filters.dateTo
            "
          >
            <span class="filter-count">{{ selections.length }} / {{ allSelections.length }}</span>
            <button
              @click="resetFilters"
              class="clear-filters-btn"
              :title="t('carStock.clear_filters') || 'Clear Filters'"
            >
              <i class="fas fa-times"></i>
            </button>
          </div>
        </div>

        <div v-if="showFilters" class="filters-panel">
          <div class="filters-grid">
            <!-- Search Filter -->
            <div class="filter-group">
              <label>
                <i class="fas fa-search"></i>
                {{ t('carStock.search') || 'Search' }}
              </label>
              <input
                type="text"
                v-model="filters.search"
                :placeholder="
                  t('carStock.search_selections_placeholder') || 'Search by name or description...'
                "
                @input="applyFilters"
                class="filter-input"
              />
            </div>

            <!-- Status Filter -->
            <div class="filter-group">
              <label>
                <i class="fas fa-info-circle"></i>
                {{ t('carStock.status') || 'Status' }}
              </label>
              <select v-model="filters.status" @change="applyFilters" class="filter-input">
                <option value="">{{ t('carStock.all') || 'All' }}</option>
                <option value="pending">{{ t('carStock.status_pending') || 'Pending' }}</option>
                <option value="in_progress">
                  {{ t('carStock.status_in_progress') || 'In Progress' }}
                </option>
                <option value="completed">
                  {{ t('carStock.status_completed') || 'Completed' }}
                </option>
                <option value="cancelled">
                  {{ t('carStock.status_cancelled') || 'Cancelled' }}
                </option>
              </select>
            </div>

            <!-- Priority Filter -->
            <div class="filter-group">
              <label>
                <i class="fas fa-exclamation-circle"></i>
                {{ t('carStock.priority') || 'Priority' }}
              </label>
              <select v-model="filters.priority" @change="applyFilters" class="filter-input">
                <option value="">{{ t('carStock.all') || 'All' }}</option>
                <option value="low">{{ t('carStock.priority_low') || 'Low' }}</option>
                <option value="medium">{{ t('carStock.priority_medium') || 'Medium' }}</option>
                <option value="high">{{ t('carStock.priority_high') || 'High' }}</option>
                <option value="urgent">{{ t('carStock.priority_urgent') || 'Urgent' }}</option>
              </select>
            </div>

            <!-- Date From Filter -->
            <div class="filter-group">
              <label>
                <i class="fas fa-calendar-alt"></i>
                {{ t('carStock.date_from') || 'Date From' }}
              </label>
              <input
                type="date"
                v-model="filters.dateFrom"
                @change="applyFilters"
                class="filter-input"
              />
            </div>

            <!-- Date To Filter -->
            <div class="filter-group">
              <label>
                <i class="fas fa-calendar-alt"></i>
                {{ t('carStock.date_to') || 'Date To' }}
              </label>
              <input
                type="date"
                v-model="filters.dateTo"
                @change="applyFilters"
                class="filter-input"
              />
            </div>
          </div>

          <div class="filters-actions">
            <button @click="resetFilters" class="reset-filters-btn">
              <i class="fas fa-undo"></i>
              {{ t('carStock.reset_filters') || 'Reset Filters' }}
            </button>
          </div>
        </div>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading-state">
          <i class="fas fa-spinner fa-spin"></i>
          <span>{{ t('carStock.loading') }}</span>
        </div>

        <div v-else-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div v-else-if="selections.length === 0" class="empty-state">
          <i class="fas fa-inbox"></i>
          <p>{{ t('carStock.no_selections_found') }}</p>
        </div>

        <div v-else class="selections-list">
          <div
            v-for="selection in selections"
            :key="selection.id"
            class="selection-item"
            :class="{ selected: selectedSelection?.id === selection.id }"
          >
            <div class="selection-header">
              <div class="selection-title">
                <h4>{{ selection.name }} (ID: {{ selection.id }})</h4>
                <div class="selection-badges">
                  <div class="status-badge-container">
                    <span
                      class="status-badge"
                      :style="{
                        backgroundColor: statusColors[selection.status] + '20',
                        color: statusColors[selection.status],
                      }"
                    >
                      {{ t(`carStock.status_${selection.status}`) || selection.status }}
                    </span>
                    <select
                      v-if="changingStatusId !== selection.id"
                      @change="handleStatusChange(selection, $event.target.value)"
                      class="status-select"
                      :value="selection.status"
                      :title="t('carStock.change_status') || 'Change Status'"
                    >
                      <option :value="selection.status" disabled>
                        {{ t(`carStock.status_${selection.status}`) || selection.status }}
                      </option>
                      <option
                        v-for="status in getAvailableStatuses(selection.status)"
                        :key="status"
                        :value="status"
                      >
                        {{ t(`carStock.status_${status}`) || status }}
                      </option>
                    </select>
                    <i v-else class="fas fa-spinner fa-spin status-changing"></i>
                  </div>
                  <span
                    class="priority-badge"
                    :style="{
                      backgroundColor: priorityColors[selection.priority] + '20',
                      color: priorityColors[selection.priority],
                    }"
                  >
                    {{ t(`carStock.priority_${selection.priority}`) || selection.priority }}
                  </span>
                </div>
              </div>
              <div class="selection-actions">
                <SelectionChatButton :selection="selection" />
                <button
                  @click="toggleShowCars(selection)"
                  class="action-btn view-cars-btn"
                  :title="
                    showingCars[selection.id]
                      ? t('carStock.hide_cars') || 'Hide Cars'
                      : t('carStock.view_cars') || 'View Cars'
                  "
                >
                  <i class="fas" :class="showingCars[selection.id] ? 'fa-eye-slash' : 'fa-eye'"></i>
                  {{
                    showingCars[selection.id]
                      ? t('carStock.hide_cars') || 'Hide Cars'
                      : t('carStock.view_cars') || 'View Cars'
                  }}
                </button>
                <button
                  @click="handlePrintSelection(selection)"
                  class="action-btn print-btn"
                  :title="t('carStock.print_selection') || 'Print Selection'"
                >
                  <i class="fas fa-print"></i>
                  {{ t('carStock.print') || 'Print' }}
                </button>
                <button
                  @click="handleLoadSelection(selection)"
                  class="action-btn load-btn"
                  :title="t('carStock.load_selection')"
                >
                  <i class="fas fa-download"></i>
                  {{ t('carStock.load') }}
                </button>
                <button
                  @click="handleSendSelection(selection)"
                  class="action-btn send-btn"
                  :title="t('carStock.send_selection')"
                >
                  <i class="fas fa-paper-plane"></i>
                  {{ t('carStock.send') }}
                </button>
                <button
                  @click="handleDeleteSelection(selection)"
                  class="action-btn delete-btn"
                  :title="t('carStock.delete_selection')"
                  :disabled="deletingSelectionId === selection.id"
                >
                  <i v-if="deletingSelectionId === selection.id" class="fas fa-spinner fa-spin"></i>
                  <i v-else class="fas fa-trash"></i>
                  {{ t('carStock.delete') }}
                </button>
              </div>
            </div>

            <div v-if="selection.description" class="selection-description">
              <pre>{{ selection.description }}</pre>
            </div>

            <div class="selection-details">
              <div class="detail-item">
                <i class="fas fa-car"></i>
                <span
                  >{{ getCarCount(selection) }}
                  {{
                    getCarCount(selection) === 1
                      ? t('carStockToolbar.car')
                      : t('carStockToolbar.cars')
                  }}</span
                >
              </div>
              <div class="detail-item">
                <i class="fas fa-user"></i>
                <span>{{ t('carStock.created_by') }}: {{ selection.creator_name || '-' }}</span>
              </div>
              <div v-if="selection.team_name" class="detail-item">
                <i class="fas fa-users"></i>
                <span>{{ t('carStock.team') }}: {{ selection.team_name }}</span>
              </div>
              <div v-if="selection.sender_name" class="detail-item">
                <i class="fas fa-paper-plane"></i>
                <span>{{ t('carStock.sent_by') }}: {{ selection.sender_name }}</span>
              </div>
              <div class="detail-item">
                <i class="fas fa-calendar"></i>
                <span>{{ t('carStock.created_at') }}: {{ formatDate(selection.created_at) }}</span>
              </div>
              <div v-if="selection.due_date" class="detail-item">
                <i class="fas fa-clock"></i>
                <span>{{ t('carStock.due_date') }}: {{ formatDate(selection.due_date) }}</span>
              </div>
              <div
                v-if="selection.deadline"
                class="detail-item"
                :class="{ 'deadline-passed': isDeadlinePassed(selection.deadline) }"
              >
                <i class="fas fa-calendar-times"></i>
                <span>{{ t('carStock.deadline') }}: {{ formatDate(selection.deadline) }}</span>
                <span v-if="isDeadlinePassed(selection.deadline)" class="deadline-warning"
                  >({{ t('carStock.overdue') || 'Overdue' }})</span
                >
              </div>
              <div v-if="selection.status_changed_at" class="detail-item">
                <i class="fas fa-exchange-alt"></i>
                <span
                  >{{ t('carStock.status_changed_at') || 'Status Changed' }}:
                  {{ formatDate(selection.status_changed_at) }}</span
                >
                <span v-if="selection.previous_status" class="status-change-info">
                  ({{ t('carStock.from') || 'from' }}:
                  {{
                    t(`carStock.status_${selection.previous_status}`) || selection.previous_status
                  }})
                </span>
              </div>
              <div v-if="selection.job_done_on" class="detail-item">
                <i class="fas fa-check-circle"></i>
                <span
                  >{{ t('carStock.completed_at') }}: {{ formatDate(selection.job_done_on) }}</span
                >
              </div>
            </div>

            <!-- Cars Table Section -->
            <div v-if="showingCars[selection.id]" class="cars-table-section">
              <div v-if="loadingCars[selection.id]" class="loading-cars">
                <i class="fas fa-spinner fa-spin"></i>
                <span>{{ t('carStock.loading_cars') || 'Loading cars...' }}</span>
              </div>
              <div
                v-else-if="selectionCars[selection.id] && selectionCars[selection.id].length > 0"
                class="simple-cars-table-container"
              >
                <table class="simple-cars-table">
                  <thead>
                    <tr>
                      <th>{{ t('carStock.id') || 'ID' }}</th>
                      <th>{{ t('carStock.car_details') || 'Car' }}</th>
                      <th>{{ t('carStock.client') || 'Client' }}</th>
                      <th>{{ t('carStock.fob') || 'FOB' }}</th>
                      <th>{{ t('carStock.status') || 'Status' }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="car in selectionCars[selection.id]" :key="car.id">
                      <td>{{ car.id }}</td>
                      <td>
                        <div class="car-info-cell">
                          <div class="car-name">{{ car.car_name || '-' }}</div>
                          <div v-if="car.color" class="car-badges">
                            <span
                              class="color-badge"
                              :style="{
                                backgroundColor: car.hexa || '#000000',
                                color: getTextColor(car.hexa || '#000000'),
                              }"
                            >
                              {{ car.color }}
                            </span>
                          </div>
                          <div v-if="car.vin" class="car-vin">{{ car.vin }}</div>
                        </div>
                      </td>
                      <td>{{ car.client_name || '-' }}</td>
                      <td>${{ car.price_cell || '0' }}</td>
                      <td>
                        <span
                          class="car-status-badge"
                          :class="{
                            'status-available': car.status === 'Available',
                            'status-sold': car.status === 'Sold',
                          }"
                        >
                          {{ car.status }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-else class="no-cars-message">
                <i class="fas fa-car"></i>
                <span>{{ t('carStock.no_cars_in_selection') || 'No cars in this selection' }}</span>
              </div>
            </div>

            <!-- Comments Section -->
            <div class="comments-section">
              <button
                @click="toggleComments(selection.id)"
                class="comments-toggle-btn"
                :title="t('carStock.comments') || 'Comments'"
              >
                <i class="fas fa-comments"></i>
                <span>{{ t('carStock.comments') || 'Comments' }}</span>
                <i class="fas fa-chevron-down" :class="{ rotated: showComments[selection.id] }"></i>
              </button>

              <div v-if="showComments[selection.id]" class="comments-container">
                <div v-if="loadingComments[selection.id]" class="loading-comments">
                  <i class="fas fa-spinner fa-spin"></i>
                  <span>{{ t('carStock.loading') }}</span>
                </div>
                <div
                  v-else-if="comments[selection.id] && comments[selection.id].length > 0"
                  class="comments-list"
                >
                  <div
                    v-for="comment in comments[selection.id]"
                    :key="comment.id"
                    class="comment-item"
                  >
                    <div class="comment-header">
                      <span class="comment-author">{{ comment.username || '-' }}</span>
                      <span class="comment-date">{{ formatDate(comment.created_at) }}</span>
                    </div>
                    <div class="comment-text">{{ comment.comment }}</div>
                  </div>
                </div>
                <div v-else class="no-comments">
                  {{ t('carStock.no_comments') || 'No comments yet' }}
                </div>

                <!-- Add Comment Form -->
                <div class="add-comment-form">
                  <textarea
                    v-model="newComment[selection.id]"
                    :placeholder="t('carStock.add_comment_placeholder') || 'Add a comment...'"
                    class="comment-textarea"
                    rows="3"
                    :disabled="isAddingComment[selection.id]"
                  ></textarea>
                  <button
                    @click="addComment(selection.id)"
                    class="add-comment-btn"
                    :disabled="!newComment[selection.id]?.trim() || isAddingComment[selection.id]"
                  >
                    <i v-if="isAddingComment[selection.id]" class="fas fa-spinner fa-spin"></i>
                    <i v-else class="fas fa-paper-plane"></i>
                    {{ t('carStock.add_comment') || 'Add Comment' }}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="handleClose" class="btn btn-close">
          {{ t('carStock.close') }}
        </button>
      </div>
    </div>

    <SendSelectionForm
      :show="showSendForm"
      :selectedCarIds="selectedCarIdsForSend"
      :selectionId="selectedSelectionIdForSend"
      @close="showSendForm = false"
      @sent="handleSelectionSent"
    />

    <!-- Print Options Modal -->
    <CarStockPrintOptions
      :show="showPrintOptions"
      :selected-cars="
        selectedSelectionForPrint ? selectionCars[selectedSelectionForPrint.id] || [] : []
      "
      action-type="print"
      @close="handleClosePrintOptions"
      @print="handlePrintWithOptions"
    />
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  position: sticky;
  top: 0;
  background: white;
  z-index: 10;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.modal-body {
  padding: 24px;
  flex: 1;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  gap: 12px;
}

.loading-state i,
.empty-state i {
  font-size: 48px;
  color: #9ca3af;
}

.error-message {
  background-color: #fef2f2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.selections-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.selection-item {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
  background: white;
}

.selection-item:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
}

.selection-item.selected {
  border-color: #3b82f6;
  background-color: #eff6ff;
}

.selection-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.selection-title h4 {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.selection-badges {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.status-badge-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-badge,
.priority-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
  text-transform: capitalize;
}

.status-select {
  padding: 4px 8px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 11px;
  background-color: white;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s ease;
}

.status-select:hover {
  border-color: #3b82f6;
  background-color: #f0f9ff;
}

.status-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.status-changing {
  color: #3b82f6;
  font-size: 14px;
}

.selection-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-btn {
  padding: 6px 12px;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.load-btn {
  background-color: #3b82f6;
}

.load-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.send-btn {
  background-color: #10b981;
}

.send-btn:hover:not(:disabled) {
  background-color: #059669;
}

.delete-btn {
  background-color: #ef4444;
}

.delete-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.selection-description {
  color: #6b7280;
  font-size: 14px;
  margin-bottom: 12px;
  line-height: 1.5;
}

.selection-description pre {
  margin: 0;
  padding: 0;
  font-family: inherit;
  font-size: inherit;
  white-space: pre-wrap;
  word-wrap: break-word;
  color: inherit;
}

.selection-details {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 13px;
  color: #6b7280;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.detail-item i {
  color: #9ca3af;
  width: 16px;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-close {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-close:hover {
  background-color: #e5e7eb;
}

.deadline-passed {
  color: #ef4444;
  font-weight: 600;
}

.deadline-warning {
  color: #ef4444;
  font-weight: 600;
  margin-left: 8px;
}

.status-change-info {
  color: #6b7280;
  font-size: 12px;
  margin-left: 8px;
}

.comments-section {
  margin-top: 16px;
  border-top: 1px solid #e5e7eb;
  padding-top: 16px;
}

.comments-toggle-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background-color: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
  transition: all 0.2s ease;
  width: 100%;
  justify-content: space-between;
}

.comments-toggle-btn:hover {
  background-color: #e5e7eb;
}

.comments-toggle-btn i.fa-chevron-down {
  transition: transform 0.2s ease;
  font-size: 12px;
}

.comments-toggle-btn i.fa-chevron-down.rotated {
  transform: rotate(180deg);
}

.comments-container {
  margin-top: 12px;
  padding: 12px;
  background-color: #f9fafb;
  border-radius: 6px;
}

.loading-comments {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px;
  color: #6b7280;
  font-size: 14px;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.comment-item {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
}

.comment-author {
  font-weight: 600;
  color: #1f2937;
}

.comment-date {
  color: #6b7280;
}

.comment-text {
  color: #374151;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.no-comments {
  padding: 20px;
  text-align: center;
  color: #9ca3af;
  font-size: 14px;
  font-style: italic;
}

.add-comment-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.comment-textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  resize: vertical;
  min-height: 60px;
}

.comment-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.comment-textarea:disabled {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.add-comment-btn {
  align-self: flex-end;
  padding: 8px 16px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.add-comment-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.add-comment-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.view-cars-btn {
  background-color: #6366f1;
}

.view-cars-btn:hover:not(:disabled) {
  background-color: #4f46e5;
}

.print-btn {
  background-color: #059669;
}

.print-btn:hover:not(:disabled) {
  background-color: #047857;
}

.cars-table-section {
  margin-top: 16px;
  border-top: 1px solid #e5e7eb;
  padding-top: 16px;
}

.loading-cars,
.no-cars-message {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px;
  color: #6b7280;
  font-size: 14px;
}

.loading-cars i {
  color: #3b82f6;
}

.no-cars-message i {
  color: #9ca3af;
}

.simple-cars-table-container {
  overflow-x: auto;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background-color: #f9fafb;
}

.simple-cars-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.simple-cars-table thead {
  background-color: #f3f4f6;
}

.simple-cars-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
}

.simple-cars-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #e5e7eb;
  background-color: white;
}

.simple-cars-table tbody tr:hover {
  background-color: #f9fafb;
}

.simple-cars-table tbody tr:last-child td {
  border-bottom: none;
}

.car-info-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.car-name {
  font-weight: 500;
  color: #1f2937;
}

.car-badges {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.color-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.car-vin {
  font-size: 11px;
  color: #6b7280;
  font-family: monospace;
}

.car-status-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 500;
}

.car-status-badge.status-available {
  background-color: #d1fae5;
  color: #10b981;
  border: 1px solid #a7f3d0;
}

.car-status-badge.status-sold {
  background-color: #fee2e2;
  color: #ef4444;
  border: 1px solid #fca5a5;
}

/* Filter Toolbar Styles */
.filter-toolbar {
  border-bottom: 1px solid #e5e7eb;
  background-color: #f9fafb;
}

.toolbar-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
}

.filter-toggle-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: white;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  transition: all 0.2s ease;
}

.filter-toggle-btn:hover {
  background-color: #f3f4f6;
  border-color: #d1d5db;
}

.filter-toggle-btn.active {
  background-color: #3b82f6;
  color: white;
  border-color: #3b82f6;
}

.filter-toggle-btn i.fa-chevron-down {
  transition: transform 0.2s ease;
  font-size: 12px;
}

.filter-toggle-btn i.fa-chevron-down.rotated {
  transform: rotate(180deg);
}

.filter-summary {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-count {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  padding: 0;
  background-color: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 4px;
  cursor: pointer;
  color: #ef4444;
  transition: all 0.2s ease;
}

.clear-filters-btn:hover {
  background-color: #fecaca;
  border-color: #fca5a5;
}

.filters-panel {
  padding: 16px 20px;
  border-top: 1px solid #e5e7eb;
  background-color: white;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
}

.filter-group label i {
  color: #6b7280;
  font-size: 12px;
}

.filter-input {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  transition: all 0.2s ease;
}

.filter-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filters-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
  border-top: 1px solid #e5e7eb;
}

.reset-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background-color: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #374151;
  transition: all 0.2s ease;
}

.reset-filters-btn:hover {
  background-color: #e5e7eb;
  border-color: #d1d5db;
}

@media (max-width: 768px) {
  .filters-grid {
    grid-template-columns: 1fr;
  }
}
</style>
