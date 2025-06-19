<template>
  <section class="dashboard-section savings-section">
    <h3>Savings Goals</h3>
    <div v-if="error" class="error-message">{{ error }}</div>
    <button @click="$emit('openModal', 'savings')" class="btn add-btn">Add Savings Goal</button>
    
    <ul v-if="items.length > 0" class="item-list">
      <li v-for="item in items" :key="item.id" class="list-item">
        <div>
          <strong>{{ item.goal_name }}</strong><br>
          Target: {{ formatCurrency(item.goal_amount) }}, Current: {{ formatCurrency(item.current_amount) }}<br>
          <small v-if="item.target_date">Target Date: {{ formatDate(item.target_date) }}</small><br>
          <small v-if="item.notes">Notes: {{ item.notes }}</small>
        </div>
        <div class="actions">
          <button @click="$emit('openModal', 'savings', item)" class="btn edit-btn">Edit</button>
          <button @click="deleteSavingsItem(item.id)" class="btn delete-btn">Delete</button>
        </div>
      </li>
    </ul>
    <p v-if="loading" class="loading-message">Loading savings goals...</p>
    <p v-if="!loading && items.length === 0 && !error && !initialLoad">No savings goals yet.</p>
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

const fetchSavings = async () => {
  loading.value = true;
  error.value = null;
  initialLoad.value = false;
  try {
    const response = await api.get('/savings'); 
    items.value = response.data || []; // Corrected: API returns the array directly
  } catch (err) {
    console.error("Error fetching savings goals:", err);
    error.value = "Failed to load savings goals. " + (err.response?.data?.error || err.message);
  } finally {
    loading.value = false;
  }
};

const deleteSavingsItem = async (id) => {
  if (!confirm("Are you sure you want to delete this savings goal?")) return;
  try {
    await api.delete(`/savings/${id}`);
    fetchSavings(); // Refresh list
    emit('item-changed'); // Notify parent
  } catch (err) {
    console.error("Error deleting savings goal:", err);
    alert("Failed to delete savings goal: " + (err.response?.data?.error || err.message));
  }
};

onMounted(fetchSavings);
</script>

<style scoped src="../assets/section-styles.css"></style>