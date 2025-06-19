import axios from 'axios';

const apiClient = axios.create({
  baseURL: 'http://localhost:8080/api/v1', // Your Go API base URL
  headers: {
    'Content-Type': 'application/json',
    // You might add Authorization header here if you re-implement JWT auth
    // 'Authorization': `Bearer ${localStorage.getItem('token')}`
  }
});

// Optional: Interceptors for request or response handling
// apiClient.interceptors.response.use(response => response, error => {
//   if (error.response && error.response.status === 401) {
//     // Handle unauthorized access, e.g., redirect to login
//     console.error("Unauthorized, redirecting to login...");
//     // window.location.href = '/login'; // If you have a login page
//   }
//   return Promise.reject(error);
// });

export default apiClient;