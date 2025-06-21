<template>
  <section class="dashboard-section income-section">
    <h3>Income</h3>
    <div v-if="error" class="error-message">{{ error }}</div>
    <button @click="$emit('openModal', 'income')" class="btn add-btn">Add Income</button>
    
    <ul v-if="items.length > 0" class="item-list">
      <li v-for="item in items" :key="item.ID" class="list-item"> <!-- Changed item.id to item.ID -->
        <span>{{ formatDate(item.date) }} - {{ item.category }}: {{ formatCurrency(item.amount) }}</span>
        <small v-if="item.note">({{ item.note }})</small>
        <div class="actions">
          <button @click="$emit('openModal', 'income', item)" class="btn edit-btn">Edit</button>
          <button @click="deleteIncomeItem(item.ID)" class="btn delete-btn">Delete</button> <!-- Changed item.id to item.ID -->
        </div>
      </li>
    </ul>
    <p v-if="loading" class="loading-message">Loading income...</p>
    <p v-if="!loading && items.length === 0 && !error && !initialLoad">No income records yet.</p>
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

const fetchIncome = async () => {
  loading.value = true;
  error.value = null;
  initialLoad.value = false;
  try {
    const response = await api.get('/income');
    items.value = response.data || []; // Corrected: API returns the array directly
  } catch (err) {
    console.error("Error fetching income:", err);
    error.value = "Failed to load income data. " + (err.response?.data?.error || err.message);
  } finally {
    loading.value = false;
  }
};

const deleteIncomeItem = async (id) => {
  if (!confirm("Are you sure you want to delete this income item?")) return;
  try {
    await api.delete(`/income/${id}`);
    fetchIncome(); // Refresh list
    emit('item-changed'); // Notify parent
  } catch (err) {
    console.error("Error deleting income item:", err);
    alert("Failed to delete income item: " + (err.response?.data?.error || err.message));
  }
};

onMounted(fetchIncome);
</script>

<style scoped src="../assets/section-styles.css"></style>
<!-- You'll create section-styles.css or put styles directly here -->