<template>
  <div id="vue-app">
    <header class="app-header">
      <div class="container">
        <h1>Finance Tracker</h1>
        <!-- Optional: Navigation can be added here if you expand to more pages -->
      </div>
    </header>

    <main class="app-main">
      <div class="container">
        <Dashboard />
      </div>
    </main>

    <footer class="app-footer">
      <div class="container">
        <p>&copy; {{ currentYear }} Finance Tracker. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script setup>
import './assets/main.css'; // Ensure main.css is imported
import Dashboard from './components/Dashboard.vue';

const currentYear = new Date().getFullYear();
</script>

<style scoped>
.app-header {
  background-color: #2c3e50; /* Dark blue-gray */
  color: #ffffff;
  padding: 1rem 0; /* Consistent padding unit */
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  /* Fixed header if desired:
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  z-index: 1000; */
}

.app-header h1 {
  margin: 0;
  text-align: center;
  font-size: 1.8em; /* Existing size, can be adjusted from main.css's h1 if needed */
  color: #ffffff; /* Ensure header text is white */
}

.container {
  max-width: 1200px;
  margin-left: auto; /* Use auto for centering block elements */
  margin-right: auto;
  padding: 0 1rem; /* Consistent padding unit */
}

.app-main {
  padding: 2rem 0; /* More vertical padding for main content area */
  /* If header is fixed, add top padding to main content to prevent overlap
  padding-top: 80px; /* Example: if header is ~60px high + some space */
}

.app-footer {
  text-align: center;
  padding: 1.5rem 0; /* Consistent padding unit */
  background-color: #343a40; /* Darker gray for footer */
  color: #f8f9fa; /* Light text for contrast */
  margin-top: 2rem; /* More space above footer */
  font-size: 0.9em;
}

/* Add other global layout styles if necessary */
</style>