// Debug script to test API connectivity
console.log('=== API Debug Information ===')

// Current location info
console.log('Current URL:', window.location.href)
console.log('Hostname:', window.location.hostname)
console.log('Protocol:', window.location.protocol)
console.log('Pathname:', window.location.pathname)

// Construct API URL
const hostname = window.location.hostname
const protocol = window.location.protocol
const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
const API_BASE_URL = isLocalhost ? 'http://localhost:8000/api' : `${protocol}//${hostname}/api`
const API_URL = `${API_BASE_URL}/api.php`

console.log('=== API URLs ===')
console.log('isLocalhost:', isLocalhost)
console.log('API_BASE_URL:', API_BASE_URL)
console.log('API_URL:', API_URL)

// Test API connectivity
async function testAPIConnectivity() {
  console.log('=== Testing API Connectivity ===')

  try {
    console.log('Making test request to:', API_URL)

    const response = await fetch(API_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query: 'SELECT 1 as test',
      }),
    })

    console.log('Response Status:', response.status)
    console.log('Response Status Text:', response.statusText)
    console.log('Response Headers:', Object.fromEntries(response.headers.entries()))

    const text = await response.text()
    console.log('Response Text (first 500 chars):', text.substring(0, 500))

    // Try to parse as JSON
    try {
      const json = JSON.parse(text)
      console.log('Parsed JSON:', json)
    } catch (parseError) {
      console.error('Failed to parse as JSON:', parseError)
      console.log('Full response text:', text)
    }
  } catch (error) {
    console.error('API Test Failed:', error)
    console.error('Error details:', {
      name: error.name,
      message: error.message,
      stack: error.stack,
    })
  }
}

// Run the test
testAPIConnectivity()

// Also test with a simple GET request to see if the file exists
async function testFileExists() {
  console.log('=== Testing if API file exists ===')

  try {
    const response = await fetch(API_URL, {
      method: 'GET',
    })

    console.log('GET Response Status:', response.status)
    console.log('GET Response Text (first 200 chars):', (await response.text()).substring(0, 200))
  } catch (error) {
    console.error('GET Test Failed:', error)
  }
}

testFileExists()
