<template>
  <div id="dashboard-app">
    <h2>Dashboard</h2>
    <p v-if="userName">Welcome, {{ userName }}!</p>

    <!-- Financial Summary -->
    <section class="dashboard-section summary-section">
      <h3>Financial Summary (Current Month)</h3>
      <div v-if="summary.loading" class="loading-message">Loading summary...</div>
      <div v-if="summary.error" class="error-message">{{ summary.error }}</div>
      <div v-if="!summary.loading && !summary.error && summary.data">
        <p>Total Income: <strong>{{ formatCurrency(summary.data.total_income) }}</strong></p>
        <p>Total Expenses: <strong>{{ formatCurrency(summary.data.total_expenses) }}</strong></p>
        <p>Net Balance: <strong>{{ formatCurrency(summary.data.net_balance) }}</strong></p>
      </div>
      <div v-if="!summary.loading && !summary.error && !summary.data && !summary.initialLoad">
        No summary data available for the current month.
      </div>
    </section>

    <!-- Sections (Income, Expenses, Savings, Debts) -->
    <IncomeSection @open-modal="openModal" :key="sectionKeys.income" @item-changed="handleItemChange" />
    <ExpensesSection @open-modal="openModal" :key="sectionKeys.expenses" @item-changed="handleItemChange" />
    <SavingsSection @open-modal="openModal" :key="sectionKeys.savings" @item-changed="handleItemChange" />
    <DebtsSection @open-modal="openModal" :key="sectionKeys.debts" @item-changed="handleItemChange" />

    <!-- Add/Edit Modal (Generic) -->
    <GenericModal
      v-if="showModal"
      :mode="modalMode"
      :item-data="currentItem"
      :is-editing="isEditingModal"
      @close="closeModal"
      :processing="isSaving"
      @save="handleFormSubmit"
    />

  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue';
import api from '../services/api.js';

// Import section components (we'll create these as placeholders)
import IncomeSection from './IncomeSection.vue';
import ExpensesSection from './ExpensesSection.vue';
import SavingsSection from './SavingsSection.vue';
import DebtsSection from './DebtsSection.vue';
import GenericModal from './GenericModal.vue';

const userName = ref("User"); // Placeholder, replace with actual user data logic

const summary = reactive({
  loading: false,
  error: null,
  data: null,
  initialLoad: true, // To prevent "No data" message on initial load
});

const showModal = ref(false);
const modalMode = ref(''); // 'income', 'expenses', 'savings', 'debts'
const currentItem = ref({});
const isEditingModal = ref(false);
const isSaving = ref(false); // To manage loading state during save

const sectionKeys = reactive({
  income: 0,
  expenses: 0,
  savings: 0,
  debts: 0,
});

// --- Utility Functions (from Go templates, now in JS) ---
const formatCurrency = (value) => {
  if (typeof value !== 'number') return value;
  return `$${value.toFixed(2)}`;
};

// const formatDate = (dateString) => {
//   if (!dateString) return '';
//   const date = new Date(dateString);
//   return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' });
// };

// --- API Calls ---
const fetchSummary = async () => {
  summary.loading = true;
  summary.error = null;
  summary.initialLoad = false;
  try {
    // Assuming your Go API endpoint for monthly summary is /api/v1/summary/monthly
    // It expects a "date" query parameter in "YYYY-MM" format.
    const today = new Date();
    const year = today.getFullYear();
    const month = (today.getMonth() + 1).toString().padStart(2, '0'); // JavaScript months are 0-indexed
    const dateParam = `${year}-${month}`;
    const response = await api.get(`/summary/monthly?date=${dateParam}`);
    summary.data = response.data; // Corrected: API returns the summary object directly
  } catch (err) {
    console.error("Error fetching summary:", err);
    summary.error = "Failed to load financial summary. " + (err.response?.data?.error || err.message);
    summary.data = null; // Clear data on error
  } finally {
    summary.loading = false;
  }
};

onMounted(() => {
  fetchSummary();
  // Fetch other initial data (income, expenses, etc.)
});

// --- Modal Logic ---
const openModal = (mode, itemToEdit = null) => {
  modalMode.value = mode;
  isEditingModal.value = !!itemToEdit;
  currentItem.value = itemToEdit ? { ...itemToEdit } : getDefaultItemStructure(mode);
  showModal.value = true;
  // Clear any previous error from modal in GenericModal if it holds one
};

const closeModal = () => {
  showModal.value = false;
  currentItem.value = {};
  modalMode.value = '';
  isEditingModal.value = false;
  // Clear any error in GenericModal if needed
};

const getDefaultItemStructure = (mode) => {
    // Provide basic structure for new items
    const common = { date: new Date().toISOString().split('T')[0], amount: 0, category: '', note: '' };
    if (mode === 'income' || mode === 'expenses') return { ...common };
    if (mode === 'savings') return { goal_name: '', goal_amount: 0, current_amount: 0, target_date: '', notes: '' };
    if (mode === 'debts') return { debtor_name: '', amount: 0, due_date: '', status: 'Pending', description: '' };
    return {};
};

const handleItemChange = () => {
  // This function is called when an item is deleted from a child section
  // Or could be used more broadly if sections managed their own additions/edits
  fetchSummary(); // Refresh summary
};

const handleFormSubmit = async (formData) => {
  isSaving.value = true;
  console.log(`Saving ${modalMode.value}:`, formData);
  const endpoint = `/${modalMode.value}`;
  let itemToSave = { ...formData };

  // Backend might not expect 'id' on create, and needs it in path for update
  let itemId = null;
  if (isEditingModal.value && itemToSave.id) {
    itemId = itemToSave.id;
    // delete itemToSave.id; // Some backends prefer ID not in payload for PUT
  }

  try {
    if (isEditingModal.value) {
      await api.put(`${endpoint}/${itemId}`, itemToSave);
    } else {
      await api.post(endpoint, itemToSave);
    }
    closeModal();
    fetchSummary(); // Refresh summary

    // Trigger re-render of the specific section by incrementing its key
    if (sectionKeys.hasOwnProperty(modalMode.value)) {
      sectionKeys[modalMode.value]++;
    }

  } catch (error) {
    console.error("Error saving item:", error);
    // You should pass this error to GenericModal to display it
    // For now, an alert:
    alert(`Error saving ${modalMode.value}: ${error.response?.data?.error || error.message}`);
  } finally {
    isSaving.value = false;
  }
};

</script>

<style scoped>
/* Styles from dashboard.html can be scoped here or moved to main.css */
.dashboard-section {
  background-color: #f9f9f9;
  border: 1px solid #e0e0e0;
  padding: 15px;
  margin-bottom: 20px;
  border-radius: 5px;
}

.dashboard-section h3 {
  margin-top: 0;
  color: #333;
}

/* Add other dashboard specific styles */
/* Item list, buttons, modal styles will go into their respective components or main.css */
.summary-section p {
    margin: 0.5em 0;
}
.summary-section strong {
    color: #2c3e50;
}

</style>