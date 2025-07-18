<template>
  <div v-if="isVisible" class="modal-overlay" @click="handleOverlayClick">
    <div class="modal-content" @click.stop>
      <div class="popup-header">
        <h3>{{ $t('containersRef.selectLocation') }}</h3>
        <button @click="$emit('close')" class="close-btn">&times;</button>
      </div>

      <div class="map-container">
        <div id="google-map" ref="mapContainer" class="map"></div>
      </div>

      <div class="popup-footer">
        <div class="coordinates-display" v-if="selectedLocation">
          <p>
            <strong>{{ $t('containersRef.selectedCoordinates') }}:</strong>
          </p>
          <p>
            Lat: {{ selectedLocation.lat.toFixed(6) }}, Lng: {{ selectedLocation.lng.toFixed(6) }}
          </p>
        </div>
        <div class="actions">
          <button @click="selectCurrentLocation" class="btn btn-primary">
            {{ $t('containersRef.selectLocation') }}
          </button>
          <button @click="$emit('close')" class="btn btn-secondary">
            {{ $t('common.cancel') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { useApi } from '@/composables/useApi'

export default {
  name: 'GoogleMapPopup',
  props: {
    isVisible: {
      type: Boolean,
      default: false,
    },
    containerRef: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      map: null,
      marker: null,
      selectedLocation: null,
      googleMapsLoaded: false,
    }
  },
  setup() {
    const { callApi } = useApi()
    return { callApi }
  },
  async mounted() {
    await this.loadGoogleMapsAPI()
    this.initMap()
  },
  methods: {
    handleOverlayClick() {
      this.$emit('close')
    },
    async loadGoogleMapsAPI() {
      if (window.google && window.google.maps) {
        this.googleMapsLoaded = true
        return
      }

      return new Promise((resolve, reject) => {
        const script = document.createElement('script')
        // Use a more reliable API key or provide instructions
        script.src = `https://maps.googleapis.com/maps/api/js?key=AIzaSyB41DRUbKWJHPxaFjMAwdrzWzbVKartNGg&libraries=places`
        script.async = true
        script.defer = true

        script.onload = () => {
          this.googleMapsLoaded = true
          resolve()
        }

        script.onerror = () => {
          console.error('Failed to load Google Maps API. Please check your API key.')
          // Show a fallback message
          this.showFallbackMessage()
          reject(new Error('Failed to load Google Maps API'))
        }

        document.head.appendChild(script)
      })
    },

    showFallbackMessage() {
      const mapContainer = this.$refs.mapContainer
      if (mapContainer) {
        mapContainer.innerHTML = `
          <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; padding: 2rem; text-align: center; background: #f5f5f5;">
            <h3>Google Maps Not Available</h3>
            <p>To use the map functionality, you need to:</p>
            <ol style="text-align: left; max-width: 400px;">
              <li>Get a Google Maps API key from <a href="https://console.cloud.google.com/" target="_blank">Google Cloud Console</a></li>
              <li>Replace the API key in the GoogleMapPopup.vue component</li>
              <li>Enable the Maps JavaScript API in your Google Cloud project</li>
            </ol>
            <p><strong>For now, you can manually enter coordinates:</strong></p>
            <div style="margin-top: 1rem;">
              <input type="text" id="manual-lat" placeholder="Latitude" style="margin: 0.5rem; padding: 0.5rem; width: 150px;">
              <input type="text" id="manual-lng" placeholder="Longitude" style="margin: 0.5rem; padding: 0.5rem; width: 150px;">
              <button onclick="window.selectManualLocation()" style="margin: 0.5rem; padding: 0.5rem 1rem; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">
                Select Location
              </button>
            </div>
          </div>
        `

        // Add global function for manual location selection
        window.selectManualLocation = () => {
          const lat = document.getElementById('manual-lat').value
          const lng = document.getElementById('manual-lng').value

          if (lat && lng) {
            this.selectedLocation = {
              lat: parseFloat(lat),
              lng: parseFloat(lng),
            }
            this.selectCurrentLocation()
          } else {
            alert('Please enter both latitude and longitude')
          }
        }
      }
    },

    initMap() {
      if (!this.googleMapsLoaded) {
        // If Google Maps failed to load, show fallback
        if (!window.google || !window.google.maps) {
          this.showFallbackMessage()
          return
        }
        setTimeout(() => this.initMap(), 100)
        return
      }

      try {
        const mapOptions = {
          center: { lat: 20, lng: 0 }, // Center on world view
          zoom: 2, // Zoom out to show continents
          mapTypeId: google.maps.MapTypeId.ROADMAP,
        }

        this.map = new google.maps.Map(this.$refs.mapContainer, mapOptions)

        // Add click listener to map
        this.map.addListener('click', (event) => {
          this.setMarker(event.latLng)
        })

        // Remove automatic geolocation - just show world view
      } catch (error) {
        console.error('Error initializing Google Maps:', error)
        this.showFallbackMessage()
      }
    },

    getCurrentLocation() {
      // This method is no longer used for automatic centering
      // Keep it for potential future use if needed
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (position) => {
            const pos = {
              lat: position.coords.latitude,
              lng: position.coords.longitude,
            }
            // Don't automatically center or place marker
            // User can manually navigate to their location if needed
          },
          (error) => {
            console.log('Geolocation error:', error)
            // Default world view is already set in initMap
          },
        )
      }
    },

    setMarker(latLng) {
      // Check if Google Maps is available
      if (!window.google || !window.google.maps) {
        console.warn('Google Maps not available, using manual coordinates')
        this.selectedLocation = {
          lat: typeof latLng.lat === 'function' ? latLng.lat() : latLng.lat,
          lng: typeof latLng.lng === 'function' ? latLng.lng() : latLng.lng,
        }
        return
      }

      // Remove existing marker
      if (this.marker) {
        this.marker.setMap(null)
      }

      // Create custom boat icon
      const boatIcon = {
        url:
          'data:image/svg+xml;charset=UTF-8,' +
          encodeURIComponent(`
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M16 2L20 8H24L26 12V16L24 20H20L16 26L12 20H8L6 16V12L8 8H12L16 2Z" fill="#3B82F6" stroke="#1E40AF" stroke-width="1"/>
            <path d="M16 6L18 10H22L23 13V15L22 18H18L16 22L14 18H10L9 15V13L10 10H14L16 6Z" fill="#60A5FA" stroke="#3B82F6" stroke-width="0.5"/>
            <circle cx="16" cy="12" r="2" fill="#FFFFFF"/>
            <path d="M12 24L16 28L20 24" stroke="#1E40AF" stroke-width="1" fill="none"/>
          </svg>
        `),
        scaledSize: new google.maps.Size(32, 32),
        anchor: new google.maps.Point(16, 16),
        origin: new google.maps.Point(0, 0),
      }

      // Add new marker with boat icon
      this.marker = new google.maps.Marker({
        position: latLng,
        map: this.map,
        draggable: true,
        icon: boatIcon,
        title: `Container: ${this.containerRef}`,
      })

      this.selectedLocation = {
        lat: latLng.lat(),
        lng: latLng.lng(),
      }

      // Add drag listener
      this.marker.addListener('dragend', (event) => {
        this.selectedLocation = {
          lat: event.latLng.lat(),
          lng: event.latLng.lng(),
        }
      })
    },

    selectCurrentLocation() {
      if (this.selectedLocation) {
        // Save tracking location to database
        this.saveTrackingLocation()
      } else {
        alert(this.$t('containersRef.pleaseSelectLocation'))
      }
    },

    async saveTrackingLocation() {
      try {
        const trackingData = {
          container_ref: this.containerRef,
          tracking: `${this.selectedLocation.lat.toFixed(6)},${this.selectedLocation.lng.toFixed(6)}`,
          time: new Date().toISOString().slice(0, 19).replace('T', ' '), // MySQL timestamp format
          id_user: 1, // You can get this from user session if available
        }

        // Use INSERT ... ON DUPLICATE KEY UPDATE to handle duplicates
        const sql = `
          INSERT INTO tracking (container_ref, tracking, time, id_user) 
          VALUES (?, ?, ?, ?) 
          ON DUPLICATE KEY UPDATE 
          tracking = VALUES(tracking), 
          time = VALUES(time), 
          id_user = VALUES(id_user)
        `

        const result = await this.callApi({
          action: 'execute_sql',
          query: sql,
          params: [
            trackingData.container_ref,
            trackingData.tracking,
            trackingData.time,
            trackingData.id_user,
          ],
        })

        if (result.success) {
          alert(
            `${this.$t('containersRef.containerRef')}: ${this.containerRef}\n${this.$t('containersRef.coordinates')}: ${this.selectedLocation.lat.toFixed(6)}, ${this.selectedLocation.lng.toFixed(6)}\n\n${this.$t('containersRef.locationSaved')}`,
          )
          this.$emit('location-selected', {
            containerRef: this.containerRef,
            coordinates: this.selectedLocation,
          })
          this.$emit('close')
        } else {
          alert(`${this.$t('containersRef.saveError')}: ${result.message}`)
        }
      } catch (error) {
        console.error('Error saving tracking location:', error)
        alert(`${this.$t('containersRef.saveError')}: ${error.message}`)
      }
    },
  },
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10000;
  backdrop-filter: blur(2px);
}

.modal-content {
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
  width: 90%;
  max-width: 900px;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  z-index: 10001;
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(-20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.popup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  border-radius: 12px 12px 0 0;
}

.popup-header h3 {
  margin: 0;
  color: #333;
  font-size: 1.25rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #6c757d;
  padding: 0.5rem;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #e9ecef;
  color: #333;
}

.map-container {
  flex: 1;
  position: relative;
  min-height: 400px;
}

.map {
  width: 100%;
  height: 100%;
  min-height: 400px;
  border-radius: 0;
}

.popup-footer {
  background: white;
  padding: 1.5rem;
  border-top: 1px solid #e9ecef;
  border-radius: 0 0 12px 12px;
}

.coordinates-display {
  margin-bottom: 1rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.coordinates-display p {
  margin: 0.25rem 0;
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

.actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary {
  background: #007bff;
  color: white;
}

.btn-primary:hover {
  background: #0056b3;
  transform: translateY(-1px);
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-secondary:hover {
  background: #545b62;
  transform: translateY(-1px);
}

/* Responsive design */
@media (max-width: 768px) {
  .modal-content {
    width: 95%;
    max-height: 90vh;
    margin: 1rem;
  }

  .popup-header {
    padding: 1rem;
  }

  .popup-footer {
    padding: 1rem;
  }

  .actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .modal-content {
    width: 98%;
    max-height: 95vh;
  }

  .popup-header h3 {
    font-size: 1.1rem;
  }

  .map-container {
    min-height: 300px;
  }
}
</style>
