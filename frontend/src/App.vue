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
import Dashboard from './components/Dashboard.vue';

const currentYear = new Date().getFullYear();
</script>

<style scoped>
/* Styles from main.html can be scoped here or moved to main.css */
.app-header { background-color: #2c3e50; color: white; padding: 1em 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
.app-header h1 { margin: 0; text-align: center; font-size: 1.8em; }
.container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
.app-main { padding: 20px 0; }
.app-footer { text-align: center; padding: 20px 0; background-color: #34495e; color: #ecf0f1; margin-top: 30px; font-size: 0.9em;}

/* Add other global layout styles if necessary */
</style>