<template>
  <section class="dashboard-section debts-section">
    <h3>Debts</h3>
    <div v-if="error" class="error-message">{{ error }}</div>
    <button @click="$emit('openModal', 'debts')" class="btn add-btn">Add Debt</button>
    
    <ul v-if="items.length > 0" class="item-list">
      <li v-for="item in items" :key="item.ID" class="list-item"> <!-- Changed item.id to item.ID -->
        <div>
          <strong>{{ item.debtor_name }}</strong> - {{ formatCurrency(item.amount) }}<br>
          Due: {{ formatDate(item.due_date) }} - Status: <span :class="`status-${item.status?.toLowerCase()}`">{{ item.status }}</span><br>
          <small v-if="item.description">Description: {{ item.description }}</small>
        </div>
        <div class="actions">
          <button @click="$emit('openModal', 'debts', item)" class="btn edit-btn">Edit</button>
          <button @click="deleteDebtItem(item.ID)" class="btn delete-btn">Delete</button> <!-- Changed item.id to item.ID -->
        </div>
      </li>
    </ul>
    <p v-if="loading" class="loading-message">Loading debts...</p>
    <p v-if="!loading && items.length === 0 && !error && !initialLoad">No debt records yet.</p>
  </section>
</template>

<script setup>
import { ref, onMounted, defineEmits } from 'vue';
import api from '../services/api.js';

const emit = defineEmits(['openModal', 'item-changed']);

const items = ref([]);
const loading = ref(false);
const error = ref(null);
const initialLoad = ref(true);

const formatCurrency = (value) => value ? `$${Number(value).toFixed(2)}` : '$0.00';
const formatDate = (dateString) => {
  if (!dateString) return '';
  // Ensure date is parsed correctly, especially if it's just YYYY-MM-DD
  const date = new Date(dateString + 'T00:00:00'); // Treat as local time
  return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
};

const fetchDebts = async () => {
  loading.value = true;
  error.value = null;
  initialLoad.value = false;
  try {
    const response = await api.get('/debts'); 
    items.value = response.data || []; // Corrected: API returns the array directly
  } catch (err) {
    console.error("Error fetching debts:", err);
    error.value = "Failed to load debts. " + (err.response?.data?.error || err.message);
  } finally {
    loading.value = false;
  }
};

const deleteDebtItem = async (id) => {
  if (!confirm("Are you sure you want to delete this debt record?")) return;
  try {
    await api.delete(`/debts/${id}`);
    fetchDebts(); // Refresh list
    emit('item-changed'); // Notify parent
  } catch (err) {
    console.error("Error deleting debt record:", err);
    alert("Failed to delete debt record: " + (err.response?.data?.error || err.message));
  }
};

onMounted(fetchDebts);
</script>

<style scoped src="../assets/section-styles.css"></style>
<style scoped>
/* Additional styles for status, if needed */
.status-paid { color: green; font-weight: bold; }
.status-pending { color: orange; }
.status-overdue { color: red; font-weight: bold; }
</style>