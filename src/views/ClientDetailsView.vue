<script setup>
import { ref, onMounted, computed, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../composables/useApi'
import { useEnhancedI18n } from '../composables/useI18n'
import { ElMessage } from 'element-plus'
import ChatModal from '../components/ChatModal.vue'
import { useCarClientChat } from '../composables/useCarClientChat'

const route = useRoute()
const { callApi, error, getFileUrl } = useApi()
const { t, locale, availableLocales } = useEnhancedI18n()
const { createOrGetCarClientChatGroup } = useCarClientChat()
const client = ref(null)
const clientCars = ref([])
const isLoading = ref(false)
const shareUrl = ref(window.location.href)
const isDev = ref(process.env.NODE_ENV === 'development')
const user = ref(JSON.parse(localStorage.getItem('user')))
const isAdmin = computed(() => user.value?.role_id === 1)
const showChatModal = ref(false)

// Create computed property for available locales with proper format
const availableLanguages = computed(() => [
  { code: 'en', name: 'English' },
  { code: 'ar', name: 'العربية' },
  { code: 'fr', name: 'Français' },
  { code: 'zh', name: '中文' },
])

// Google Maps related
const googleMapsLoaded = ref(false)
const carLocations = ref({})

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(shareUrl.value)
    ElMessage({
      message: 'Link copied to clipboard!',
      type: 'success',
    })
  } catch (err) {
    ElMessage({
      message: 'Failed to copy link',
      type: 'error',
    })
  }
}

const loadGoogleMapsAPI = async () => {
  if (window.google && window.google.maps) {
    googleMapsLoaded.value = true
    return
  }

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    // Use environment variable for API key
    const apiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY || 'YOUR_API_KEY_HERE'
    script.src = `https://maps.googleapis.com/maps/api/js?key=${apiKey}&libraries=places`
    script.async = true
    script.defer = true

    script.onload = () => {
      googleMapsLoaded.value = true
      resolve()
    }

    script.onerror = () => {
      reject(new Error('Failed to load Google Maps API'))
    }

    document.head.appendChild(script)
  })
}

const initCarMap = (containerRef, mapElement) => {
  console.log('initCarMap called:', {
    containerRef,
    mapElement,
    googleMapsLoaded: googleMapsLoaded.value,
    carLocations: carLocations.value,
  })

  if (!googleMapsLoaded.value) {
    console.log('Google Maps not loaded yet')
    return
  }

  if (!carLocations.value[containerRef]) {
    console.log('No location data for container:', containerRef)
    return
  }

  if (!mapElement) {
    console.log('Map element not provided')
    return
  }

  const location = carLocations.value[containerRef]
  console.log('Creating map for location:', location)

  // Remove loading state
  const loadingElement = mapElement.querySelector('.map-loading')
  if (loadingElement) {
    loadingElement.remove()
  }

  const mapOptions = {
    center: { lat: 20, lng: 0 }, // Center on world view
    zoom: 2, // Zoom out to show continents
    mapTypeId: google.maps.MapTypeId.ROADMAP,
    disableDefaultUI: true,
    zoomControl: true,
    streetViewControl: false,
    mapTypeControl: false,
    fullscreenControl: false,
  }

  try {
    const map = new google.maps.Map(mapElement, mapOptions)
    console.log('Map created successfully')

    // Create boat icon
    const boatIcon = {
      url:
        'data:image/svg+xml;charset=UTF-8,' +
        encodeURIComponent(`
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 1.5L15 6H18L19.5 9V12L18 15H15L12 19.5L9 15H6L4.5 12V9L6 6H9L12 1.5Z" fill="#3B82F6" stroke="#1E40AF" stroke-width="0.5"/>
          <path d="M12 4.5L13.5 7.5H16.5L17.25 9V10.5L16.5 12H13.5L12 14.25L10.5 12H7.5L6.75 10.5V9L7.5 7.5H10.5L12 4.5Z" fill="#60A5FA" stroke="#3B82F6" stroke-width="0.25"/>
          <circle cx="12" cy="9" r="1.5" fill="#FFFFFF"/>
          <path d="M9 18L12 21L15 18" stroke="#1E40AF" stroke-width="0.5" fill="none"/>
        </svg>
      `),
      scaledSize: new google.maps.Size(24, 24),
      anchor: new google.maps.Point(12, 12),
    }

    // Add marker at the actual location
    try {
      // Try to use AdvancedMarkerElement if available
      if (google.maps.marker && google.maps.marker.AdvancedMarkerElement) {
        const marker = new google.maps.marker.AdvancedMarkerElement({
          position: { lat: location.lat, lng: location.lng },
          map: map,
          title: `Container: ${containerRef}`,
          content: new google.maps.marker.PinElement({
            background: '#3B82F6',
            borderColor: '#1E40AF',
            glyph: '🚢',
            glyphColor: '#FFFFFF',
            scale: 1.2,
          }).element,
        })
      } else {
        // Fallback to traditional Marker
        const marker = new google.maps.Marker({
          position: { lat: location.lat, lng: location.lng },
          map: map,
          icon: boatIcon,
          title: `Container: ${containerRef}`,
        })
      }
    } catch (markerError) {
      console.warn('AdvancedMarkerElement not available, using traditional Marker')
      const marker = new google.maps.Marker({
        position: { lat: location.lat, lng: location.lng },
        map: map,
        icon: boatIcon,
        title: `Container: ${containerRef}`,
      })
    }

    console.log('Marker added successfully')
  } catch (error) {
    console.error('Error creating map:', error)
  }
}

const initializeMaps = async () => {
  await nextTick()

  // Wait a bit more for DOM to be fully ready
  setTimeout(() => {
    clientCars.value.forEach((car) => {
      if (car.container_ref && carLocations.value[car.container_ref]) {
        const mapElement = document.querySelector(`[data-map-id="map-${car.id}"]`)
        if (mapElement) {
          console.log(`Initializing map for car ${car.id}, container ${car.container_ref}`)

          // Ensure the map container has proper dimensions
          if (mapElement.offsetHeight === 0) {
            console.log('Map container has no height, setting explicit height')
            mapElement.style.height = '200px'
          }

          initCarMap(car.container_ref, mapElement)
        } else {
          console.log(`Map element not found for car ${car.id}`)
        }
      }
    })
  }, 200) // Increased timeout to ensure DOM is ready
}

const fetchCarLocations = async () => {
  try {
    // Get all container references from client cars
    const containerRefs = clientCars.value
      .filter((car) => car.container_ref)
      .map((car) => car.container_ref)

    if (containerRefs.length === 0) return

    // Fetch tracking data for all container references
    const trackingResult = await callApi({
      action: 'execute_sql',
      query: `
        SELECT container_ref, tracking, time, id_user 
        FROM tracking 
        WHERE container_ref IN (${containerRefs.map(() => '?').join(',')})
      `,
      params: containerRefs,
      requiresAuth: false,
    })

    if (trackingResult.success) {
      const locations = {}
      trackingResult.results.forEach((track) => {
        if (track.tracking) {
          const [lat, lng] = track.tracking.split(',').map((coord) => parseFloat(coord))
          if (!isNaN(lat) && !isNaN(lng)) {
            locations[track.container_ref] = { lat, lng, time: track.time, user: track.id_user }
          }
        }
      })
      carLocations.value = locations
      console.log('Car locations loaded:', carLocations.value)
    }
  } catch (err) {
    console.error('Error fetching car locations:', err)
  }
}

// Watch for changes in carLocations and initialize maps
watch(
  carLocations,
  async (newLocations) => {
    if (Object.keys(newLocations).length > 0 && googleMapsLoaded.value) {
      await initializeMaps()
    }
  },
  { deep: true },
)

// Watch for Google Maps API loading
watch(googleMapsLoaded, async (loaded) => {
  if (loaded && Object.keys(carLocations.value).length > 0) {
    await initializeMaps()
  }
})

const fetchClientDetails = async () => {
  isLoading.value = true
  try {
    // First, get client ID from token
    const clientLookupResult = await callApi({
      query: `
        SELECT id FROM clients WHERE share_token = ?
      `,
      params: [route.params.token],
      requiresAuth: false,
    })

    if (!clientLookupResult.success || !clientLookupResult.data || clientLookupResult.data.length === 0) {
      error.value = 'Client not found or invalid link'
      isLoading.value = false
      return
    }

    const clientId = clientLookupResult.data[0].id

    const [clientResult, carsResult] = await Promise.all([
      callApi({
        query: `
          SELECT 
            c.id,
            c.name,
            c.email,
            c.mobiles,
            c.id_no,
            c.address,
            c.is_broker,
            COUNT(cs.id) as cars_count
          FROM clients c
          LEFT JOIN cars_stock cs ON c.id = cs.id_client
          WHERE c.share_token = ?
          GROUP BY c.id
        `,
        params: [route.params.token],
        requiresAuth: false,
      }),
      callApi({
        query: `
          SELECT 
            cs.*,
            w.warhouse_name as warehouse_name,
            bd.amount as buy_price,
            bd.year,
            bd.is_used_car,
            bd.is_big_car,
            cn.car_name as model,
            b.brand,
            c.color,
            dp.discharge_port,
            cs.container_ref,
            COALESCE(
              (SELECT SUM(sp.amount_da) 
               FROM sell_bill sb 
               JOIN sell_payments sp ON sp.id_sell_bill = sb.id 
               WHERE sb.id = cs.id_sell),
              0
            ) as total_paid,
            CASE 
              WHEN cs.id_client IS NOT NULL THEN 'Reserved'
              ELSE 'Available'
            END as status
          FROM cars_stock cs
          LEFT JOIN warehouses w ON cs.id_warehouse = w.id
          LEFT JOIN buy_details bd ON cs.id_buy_details = bd.id
          LEFT JOIN cars_names cn ON bd.id_car_name = cn.id
          LEFT JOIN brands b ON cn.id_brand = b.id
          LEFT JOIN colors c ON bd.id_color = c.id
          LEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id
          WHERE cs.id_client = ? AND cs.hidden = 0
          ORDER BY cs.in_wharhouse_date DESC, cs.id DESC
        `,
        params: [clientId],
        requiresAuth: false,
      }),
    ])

    console.log('Client Result:', clientResult)
    console.log('Cars Result:', carsResult)

    if (clientResult.success && clientResult.data.length > 0) {
      client.value = clientResult.data[0]
    } else {
      error.value = 'Client not found'
    }

    if (carsResult.success) {
      clientCars.value = carsResult.data
      console.log('Processed Cars:', clientCars.value)

      // Load Google Maps and fetch car locations
      await loadGoogleMapsAPI()
      await fetchCarLocations()

      // Create chat groups for each car
      await createChatGroupsForCars()
    }
  } catch (err) {
    console.error('Error fetching client details:', err)
    error.value = 'Failed to load client details'
  } finally {
    isLoading.value = false
  }
}

const formatCurrency = (value) => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'
  try {
    const date = new Date(timestamp)
    return date.toLocaleString()
  } catch (error) {
    return timestamp
  }
}

const formatContainerRef = (containerRef) => {
  if (!containerRef || containerRef === 'Not provided') return 'Not provided'

  // Show only first 4 characters, replace the rest with asterisks
  if (containerRef.length <= 4) {
    return containerRef
  }

  return containerRef.substring(0, 4) + '*'.repeat(containerRef.length - 4)
}

// Create chat groups for each car
const createChatGroupsForCars = async () => {
  if (!client.value || !clientCars.value || clientCars.value.length === 0) {
    return
  }

  try {
    for (const car of clientCars.value) {
      // Generate car name: "Brand Model" or "Model" or "Car ID"
      const carName = car.brand && car.model
        ? `${car.brand} ${car.model}`
        : car.model || car.brand || `Car ${car.id}`

      // Create or get chat group for this car
      // Note: We don't need to track users found here as it's just initialization
      await createOrGetCarClientChatGroup(
        client.value.id,
        client.value.name,
        car.id,
        carName
      )
    }
    console.log('Chat groups created/verified for all cars')
  } catch (err) {
    console.error('Error creating chat groups for cars:', err)
    // Don't show error to client, just log it
  }
}

// Refresh chat groups and check for new users to add
const refreshChatGroupsAndOpenChat = async () => {
  if (!client.value || !clientCars.value || clientCars.value.length === 0) {
    showChatModal.value = true
    return
  }

  try {
    let totalUsersFound = 0
    
    // Refresh all chat groups for each car (this will check for new users)
    for (const car of clientCars.value) {
      // Generate car name: "Brand Model" or "Model" or "Car ID"
      const carName = car.brand && car.model
        ? `${car.brand} ${car.model}`
        : car.model || car.brand || `Car ${car.id}`

      // Re-run createOrGetCarClientChatGroup to check for new users
      const result = await createOrGetCarClientChatGroup(
        client.value.id,
        client.value.name,
        car.id,
        carName
      )
      
      if (result && result.usersFound) {
        totalUsersFound += result.usersFound
      }
    }
    
    console.log('Chat groups refreshed, checking for new users completed')
    
    // Show informational message if no users found
    if (totalUsersFound === 0) {
      ElMessage({
        message: t('clientDetails.noUsersFoundForChat'),
        type: 'info',
        duration: 5000,
      })
    } else {
      ElMessage({
        message: t('clientDetails.usersAddedToChat', { count: totalUsersFound }),
        type: 'success',
        duration: 3000,
      })
    }
  } catch (err) {
    console.error('Error refreshing chat groups:', err)
    // Don't show error to client, just log it
  } finally {
    // Open chat modal after refreshing
    showChatModal.value = true
  }
}

onMounted(() => {
  fetchClientDetails()
})
</script>

<template>
  <div class="client-details">
    <div class="header">
      <div class="header-left">
        <h2>
          <i class="fas fa-user"></i>
          {{ t('clientDetails.title') }}
        </h2>
        <router-link to="/clients" class="back-btn" v-if="$route.query.from === 'clients'">
          <i class="fas fa-arrow-left"></i>
          {{ t('clientDetails.backToClients') }}
        </router-link>
      </div>
      <div class="header-right">
        <div class="language-switcher">
          <label for="language-select">{{ t('common.language') }}:</label>
          <select id="language-select" v-model="locale" class="language-select">
            <option v-for="lang in availableLanguages" :key="lang.code" :value="lang.code">
              {{ lang.name }}
            </option>
          </select>
        </div>
        <div class="share-section">
          <div class="share-url">
            <input type="text" :value="shareUrl" readonly />
            <button @click="copyToClipboard" class="copy-btn">
              <i class="fas fa-copy"></i>
              {{ t('clientDetails.copyLink') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Chat Button - Top of View -->
    <div v-if="client && !isLoading" class="info-section chat-button-section top-chat-button">
      <button @click="refreshChatGroupsAndOpenChat" class="open-chat-btn">
        <i class="fas fa-comments"></i>
        {{ t('clientDetails.openChat') }}
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <i class="fas fa-spinner fa-spin"></i>
      {{ t('clientDetails.loadingClientDetails') }}
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <div v-if="client && !isLoading" class="client-info">
      <div class="info-section">
        <h3>{{ t('clientDetails.clientInformation') }}</h3>
        <div class="client-name">
          <h2>{{ client.name }}</h2>
        </div>
      </div>

      <!-- Cars Section -->
      <div v-if="clientCars.length > 0" class="info-section cars-section">
        <h3>
          <i class="fas fa-car"></i>
          {{ t('clientDetails.cars') }}
        </h3>
        <div class="cars-grid">
          <div v-for="car in clientCars" :key="car.id" class="car-card">
            <div class="car-details">
              <h4>{{ car.brand || t('clientDetails.unknownBrand') }} {{ car.model || '' }}</h4>
              <div class="car-info-grid">
                <div class="car-info-item">
                  <label>VIN:</label>
                  <span>{{ car.vin || t('clientDetails.notAvailable') }}</span>
                </div>
                <div class="car-info-item">
                  <label>Year:</label>
                  <span>{{ car.year || t('clientDetails.notSpecified') }}</span>
                </div>
                <div class="car-info-item">
                  <label>Color:</label>
                  <span>{{ car.color || t('clientDetails.notSpecified') }}</span>
                </div>
                <div class="car-info-item">
                  <label>Documents Path:</label>
                  <span>{{ car.path_documents || t('clientDetails.notAvailable') }}</span>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.loadingDate') }}:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_loding
                        ? formatDate(car.date_loding)
                        : t('clientDetails.notAvailable')
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_loding ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.documentsSent') }}:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_send_documents
                        ? formatDate(car.date_send_documents)
                        : t('clientDetails.notAvailable')
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_send_documents ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.documentsReceived') }}:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_get_documents_from_supp
                        ? formatDate(car.date_get_documents_from_supp)
                        : t('clientDetails.notAvailable')
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_get_documents_from_supp
                          ? 'fa-check-circle done'
                          : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.blDate') }}:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_get_bl
                        ? formatDate(car.date_get_bl)
                        : t('clientDetails.notAvailable')
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_get_bl ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.freightPayment') }}:</label>
                  <div class="status-field">
                    <span>{{
                      car.date_pay_freight
                        ? formatDate(car.date_pay_freight)
                        : t('clientDetails.notAvailable')
                    }}</span>
                    <i
                      :class="[
                        'fas',
                        car.date_pay_freight ? 'fa-check-circle done' : 'fa-times-circle undone',
                      ]"
                    ></i>
                  </div>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.dischargePort') }}:</label>
                  <span>{{ car.discharge_port || t('clientDetails.notAvailable') }}</span>
                </div>
                <div class="car-info-item">
                  <label>Price:</label>
                  <span>{{ formatCurrency(car.price_cell || 0) }}</span>
                </div>
                <div v-if="false" class="car-info-item">
                  <label>Paid:</label>
                  <span>{{ formatCurrency(car.total_paid || 0) }}</span>
                </div>
                <div v-if="false" class="car-info-item">
                  <label>Balance:</label>
                  <span
                    :class="{ 'text-danger': (car.price_cell || 0) - (car.total_paid || 0) > 0 }"
                  >
                    {{ formatCurrency((car.price_cell || 0) - (car.total_paid || 0)) }}
                  </span>
                </div>
                <div v-if="car.is_used_car" class="car-info-item">
                  <label>Type:</label>
                  <span class="badge used">{{ t('clientDetails.usedCar') }}</span>
                </div>
                <div v-if="car.is_big_car" class="car-info-item">
                  <label>Size:</label>
                  <span class="badge big">{{ t('clientDetails.bigCar') }}</span>
                </div>
                <div class="car-info-item">
                  <label>{{ t('clientDetails.containerRef') }}:</label>
                  <span>{{ formatContainerRef(car.container_ref) }}</span>
                </div>
              </div>
              <div class="car-status" :class="car.status.toLowerCase()">
                <i class="fas fa-circle"></i>
                {{ car.status }}
              </div>

              <!-- Car Location Map -->
              <div v-if="car.container_ref && carLocations[car.container_ref]" class="car-location">
                <h5>
                  <i class="fas fa-map-marker-alt"></i>
                  {{ t('clientDetails.containerLocation') }}
                </h5>
                <div class="location-info">
                  <div class="location-details">
                    <div class="location-item">
                      <label>{{ t('clientDetails.coordinates') }}:</label>
                      <span class="coordinates"
                        >{{ carLocations[car.container_ref].lat.toFixed(6) }},
                        {{ carLocations[car.container_ref].lng.toFixed(6) }}</span
                      >
                    </div>
                    <div class="location-item">
                      <label>{{ t('clientDetails.lastUpdated') }}:</label>
                      <span>{{ formatTimestamp(carLocations[car.container_ref].time) }}</span>
                    </div>
                  </div>
                  <div class="map-container">
                    <div :data-map-id="`map-${car.id}`" class="car-map">
                      <div class="map-loading">
                        <i class="fas fa-spinner fa-spin"></i>
                        {{ t('clientDetails.loadingMap') }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- No Map Available -->
              <div v-else-if="!car.container_ref || car.container_ref === ''" class="no-map">
                <h5>
                  <i class="fas fa-map-marker-alt"></i>
                  {{ t('clientDetails.containerLocation') }}
                </h5>
                <div class="no-map-message">
                  <i class="fas fa-exclamation-circle"></i>
                  <span>{{ t('clientDetails.noLocationData') }}</span>
                </div>
              </div>

              <div
                v-if="car.path_documents || car.sell_pi_path || car.buy_pi_path"
                class="car-documents"
              >
                <h5>{{ t('clientDetails.documents') }}</h5>
                <div class="document-links">
                  <a
                    v-if="car.path_documents"
                    :href="getFileUrl(car.path_documents)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-alt"></i>
                    {{ t('clientDetails.carDocuments') }}
                  </a>
                  <a
                    v-if="car.sell_pi_path"
                    :href="getFileUrl(car.sell_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-invoice-dollar"></i>
                    {{ t('clientDetails.sellPI') }}
                  </a>
                  <a
                    v-if="car.buy_pi_path && isAdmin"
                    :href="getFileUrl(car.buy_pi_path)"
                    target="_blank"
                    class="document-link"
                  >
                    <i class="fas fa-file-invoice"></i>
                    {{ t('clientDetails.buyPI') }}
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-else-if="!isLoading" class="info-section no-cars">
        <div class="empty-state">
          <i class="fas fa-car fa-2x"></i>
          <p>{{ t('clientDetails.noCarsFound') }}</p>
        </div>
      </div>
    </div>

    <!-- Chat Modal -->
    <div v-if="showChatModal" class="chat-modal-overlay" @click="showChatModal = false">
      <div class="chat-modal" @click.stop>
        <div class="chat-modal-header">
          <h3>
            <i class="fas fa-comments"></i>
            {{ t('clientDetails.chat') }} {{ client.name }}
          </h3>
          <button @click="showChatModal = false" class="close-chat-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="chat-modal-content">
          <ChatModal
            v-if="client"
            :client-id="client.id"
            :client-name="client.name"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.client-details {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.language-switcher {
  display: flex;
  align-items: center;
  gap: 8px;
}

.language-switcher label {
  font-weight: 500;
  color: #666;
  font-size: 0.9em;
}

.language-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 0.9em;
  color: #333;
  background-color: #f5f5f5;
  cursor: pointer;
  transition: border-color 0.2s;
}

.language-select:hover {
  border-color: #bbb;
}

.share-section {
  flex-grow: 1;
  max-width: 500px;
}

.share-url {
  display: flex;
  gap: 8px;
  background: white;
  padding: 8px;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.share-url input {
  flex-grow: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #666;
  background: #f5f5f5;
  cursor: default;
}

.copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #e3f2fd;
  color: #1976d2;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.copy-btn:hover {
  background-color: #bbdefb;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #f0f0f0;
  border-radius: 4px;
  text-decoration: none;
  color: #333;
}

.back-btn:hover {
  background-color: #e0e0e0;
}

.loading-state {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px;
  color: #666;
}

.error-message {
  padding: 12px;
  background-color: #fee;
  color: #c00;
  border-radius: 4px;
  margin-bottom: 20px;
}

.client-info {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-section {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-section h3 {
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  gap: 8px;
}

.client-name {
  text-align: center;
  padding: 20px 0;
}

.client-name h2 {
  margin: 0;
  font-size: 2rem;
  font-weight: 600;
  color: #333;
  text-align: center;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-item label {
  font-weight: 500;
  color: #666;
}

.status-badges {
  display: flex;
  gap: 8px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
}

.badge.client {
  background-color: #e3f2fd;
  color: #1976d2;
}

.badge.broker {
  background-color: #fce4ec;
  color: #c2185b;
}

.badge.cars {
  background-color: #f5f5f5;
  color: #666;
}

.badge.cars.has-cars {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.badge.used {
  background-color: #fff3e0;
  color: #ef6c00;
}

.badge.big {
  background-color: #e8eaf6;
  color: #3f51b5;
}

/* Cars Section Styles */
.cars-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  margin-top: 16px;
}

.car-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
}

.car-card:hover {
  transform: translateY(-2px);
}

.car-images {
  height: 200px;
  overflow: hidden;
  background: #f5f5f5;
}

.car-images img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.car-details {
  padding: 16px;
}

.car-details h4 {
  margin: 0 0 12px 0;
  font-size: 1.2em;
  color: #333;
}

.car-info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.car-info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.car-info-item label {
  font-size: 0.85em;
  color: #666;
  font-weight: 500;
}

.car-info-item span {
  font-size: 0.95em;
  color: #333;
}

.car-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9em;
  margin-top: 8px;
}

.car-status.available {
  background-color: #e8f5e9;
  color: #2e7d32;
}

.car-status.sold {
  background-color: #fce4ec;
  color: #c2185b;
}

.car-status.pending {
  background-color: #fff3e0;
  color: #ef6c00;
}

.text-danger {
  color: #c2185b;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 40px;
  color: #666;
  text-align: center;
}

/* Mobile Responsive Styles */
@media (max-width: 768px) {
  .client-details {
    padding: 12px;
  }

  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-right {
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    width: 100%;
  }

  .language-switcher {
    width: 100%;
    justify-content: center;
  }

  .share-section {
    width: 100%;
    max-width: none;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .cars-grid {
    grid-template-columns: 1fr;
  }

  .car-info-grid {
    grid-template-columns: 1fr;
  }

  .car-images {
    height: 180px;
  }

  .car-details {
    padding: 12px;
  }

  .car-details h4 {
    font-size: 1.1em;
  }
}

/* Small Mobile Devices */
@media (max-width: 480px) {
  .client-details {
    padding: 8px;
  }

  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .share-url {
    flex-direction: column;
    gap: 8px;
  }

  .copy-btn {
    width: 100%;
    justify-content: center;
  }

  .car-images {
    height: 160px;
  }
}

.debug-section {
  margin: 20px 0;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #dee2e6;
}

.debug-info {
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
}

.debug-info h4 {
  margin: 10px 0;
  color: #666;
}

.debug-info pre {
  background: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #dee2e6;
  overflow-x: auto;
}

.car-documents {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.car-documents h5 {
  margin: 0 0 8px 0;
  font-size: 0.95em;
  color: #666;
}

.document-links {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.document-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background-color: #f5f5f5;
  color: #333;
  text-decoration: none;
  border-radius: 4px;
  font-size: 0.9em;
  transition: background-color 0.2s;
}

.document-link:hover {
  background-color: #e0e0e0;
}

.status-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-field i {
  font-size: 1.1em;
}

.status-field i.done {
  color: #4caf50;
}

.status-field i.undone {
  color: #f44336;
}

/* Car Location Map Styles */
.car-location {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.car-location h5 {
  margin: 0 0 12px 0;
  font-size: 0.95em;
  color: #666;
  display: flex;
  align-items: center;
  gap: 6px;
}

.car-location h5 i {
  color: #3b82f6;
}

.location-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.location-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.location-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.85em;
}

.location-item label {
  font-weight: 500;
  color: #666;
  min-width: 100px;
}

.location-item span {
  color: #333;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
}

.coordinates {
  color: #059669 !important;
  font-weight: 500;
}

.map-container {
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #e0e0e0;
}

.car-map {
  width: 100%;
  height: 200px;
  background: #f5f5f5;
  position: relative;
}

.map-loading {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 0.9em;
}

.map-loading i {
  font-size: 1.2em;
  color: #3b82f6;
}

/* Mobile responsive for maps */
@media (max-width: 768px) {
  .location-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .location-item label {
    min-width: auto;
  }

  .car-map {
    height: 180px;
  }
}

@media (max-width: 480px) {
  .car-map {
    height: 160px;
  }
}

.no-map {
  margin-top: 20px;
  padding: 20px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.no-map h5 {
  margin: 0 0 15px 0;
  color: #495057;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.no-map h5 i {
  color: #6c757d;
}

.no-map-message {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #6c757d;
  font-style: italic;
}

.no-map-message i {
  color: #ffc107;
  font-size: 16px;
}

/* Chat Button Section */
.chat-button-section {
  text-align: center;
  padding: 30px 20px;
}

.chat-button-section.top-chat-button {
  padding: 20px;
  margin-bottom: 20px;
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.3);
}

.open-chat-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 28px;
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(6, 182, 212, 0.3);
}

.top-chat-button .open-chat-btn {
  background-color: white;
  color: #06b6d4;
  font-size: 1.2rem;
  padding: 16px 32px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.top-chat-button .open-chat-btn:hover {
  background-color: #f0f9ff;
  color: #0891b2;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.open-chat-btn:hover {
  background-color: #0891b2;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.4);
}

.open-chat-btn:active {
  transform: translateY(0);
}

.open-chat-btn i {
  font-size: 1.2rem;
}

/* Chat Modal */
.chat-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  padding: 20px;
}

.chat-modal {
  background: white;
  border-radius: 12px;
  width: 100%;
  max-width: 1200px;
  height: 90%;
  max-height: 95vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.chat-modal-header {
  padding: 20px 24px;
  background-color: #06b6d4;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
}

.chat-modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.3rem;
}

.close-chat-btn {
  background: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1.3rem;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.close-chat-btn:hover {
  background: rgba(255, 255, 255, 0.3);
}

.chat-modal-content {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* Mobile Responsive for Chat Modal */
@media (max-width: 768px) {
  .chat-modal-overlay {
    padding: 0;
  }

  .chat-modal {
    max-width: 100%;
    max-height: 100vh;
    border-radius: 0;
    height: 100vh;
  }

  .chat-modal-header {
    padding: 16px 20px;
  }

  .chat-modal-header h3 {
    font-size: 1.1rem;
  }

  .open-chat-btn {
    padding: 12px 24px;
    font-size: 1rem;
    width: 100%;
    max-width: 300px;
  }
}

@media (max-width: 480px) {
  .chat-modal-header {
    padding: 14px 16px;
  }

  .chat-modal-header h3 {
    font-size: 1rem;
  }

  .open-chat-btn {
    padding: 10px 20px;
    font-size: 0.95rem;
  }
}
</style>
