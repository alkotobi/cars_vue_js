<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../composables/useApi'

const router = useRouter()
const { callApi, error: apiError, loading } = useApi()
const username = ref('')
const password = ref('')
const error = ref('')

const login = async () => {
  try {
    // Get user and verify credentials
    const result = await callApi({
      query: `SELECT u.id, u.username, u.role_id, u.password, r.role_name 
              FROM users u 
              JOIN roles r ON u.role_id = r.id 
              WHERE u.username = ?`,
      params: [username.value]
    })
    
    if (result.success && result.data.length > 0) {
      const user = result.data[0]

      // Verify password
      const verifyResult = await callApi({
        action: 'verify_password',
        password: password.value,
        hash: user.password,
      })

      if (verifyResult.success) {
        // Get user permissions
        const permissionsResult = await callApi({
          query: `SELECT p.permission_name, p.description 
                  FROM permissions p 
                  JOIN role_permissions rp ON p.id = rp.permission_id 
                  WHERE rp.role_id = ?`,
          params: [user.role_id]
        })
        
        // Store user data without the password hash
        const { password: _, ...userData } = user
        const userInfo = {
          ...userData,
          permissions: permissionsResult.success ? permissionsResult.data : []
        }
        
        localStorage.setItem('user', JSON.stringify(userInfo))
        router.push('/dashboard')
      } else {
        error.value = 'Invalid username or password'
      }
    } else {
      error.value = 'Invalid username or password'
    }
  } catch (err) {
    error.value = 'An error occurred during login'
    console.error(err)
  }
}
</script>

<template>
  <div class="login-container">
    <form @submit.prevent="login" class="login-form">
      <h2>Login</h2>
      <div class="form-group">
        <input type="text" v-model="username" placeholder="Username" required />
      </div>
      <div class="form-group">
        <input type="password" v-model="password" placeholder="Password" required />
      </div>
      <div v-if="error" class="error">{{ error }}</div>
      <button type="submit">Login</button>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
}

.login-form {
  width: 300px;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.form-group {
  margin-bottom: 15px;
}

input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  width: 100%;
  padding: 10px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #45a049;
}

.error {
  color: red;
  margin-bottom: 10px;
  text-align: center;
}
</style>
