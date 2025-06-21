<template>
  <section class="dashboard-section expense-section">
    <h3>Expenses (Current Month)</h3>
    <div v-if="error" class="error-message">{{ error }}</div>
    <button @click="$emit('openModal', 'expenses')" class="btn add-btn">Add Expense</button>
    
    <ul v-if="items.length > 0" class="item-list">
      <li v-for="item in items" :key="item.ID" class="list-item"> <!-- Changed item.id to item.ID -->
        <span>{{ formatDate(item.date) }} - {{ item.category }}: {{ formatCurrency(item.amount) }}</span>
        <small v-if="item.note">({{ item.note }})</small>
        <div class="actions">
          <button @click="$emit('openModal', 'expenses', item)" class="btn edit-btn">Edit</button>
          <button @click="deleteExpenseItem(item.ID)" class="btn delete-btn">Delete</button> <!-- Changed item.id to item.ID -->
        </div>
      </li>
    </ul>
    <p v-if="loading" class="loading-message">Loading expenses...</p>
    <p v-if="!loading && items.length === 0 && !error && !initialLoad">No expense records yet.</p>
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

const fetchExpenses = async () => {
  loading.value = true;
  error.value = null;
  initialLoad.value = false; // This remains to control the "No records yet" message logic

  const formatDateToYYYYMMDD = (date) => {
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    return `${year}-${month}-${day}`;
  };

  const today = new Date();
  const firstDay = new Date(today.getFullYear(), today.getMonth(), 1);
  const lastDay = new Date(today.getFullYear(), today.getMonth() + 1, 0); // 0 day of next month gives last day of current

  const startDateStr = formatDateToYYYYMMDD(firstDay);
  const endDateStr = formatDateToYYYYMMDD(lastDay);

  try {
    // Note: API endpoint is '/expenses'
    const response = await api.get(`/expenses?startDate=${startDateStr}&endDate=${endDateStr}`);
    items.value = response.data || [];
  } catch (err) {
    console.error("Error fetching expenses:", err);
    error.value = "Failed to load expense data. " + (err.response?.data?.error || err.message);
  } finally {
    loading.value = false;
  }
};

const deleteExpenseItem = async (id) => {
  if (!confirm("Are you sure you want to delete this expense item?")) return;
  try {
    // Note: API endpoint is '/expenses/:id'
    await api.delete(`/expenses/${id}`);
    fetchExpenses(); // Refresh list
    emit('item-changed'); // Notify parent
  } catch (err) {
    console.error("Error deleting expense item:", err);
    alert("Failed to delete expense item: " + (err.response?.data?.error || err.message));
  }
};

onMounted(fetchExpenses);
</script>

<style scoped src="../assets/section-styles.css"></style>