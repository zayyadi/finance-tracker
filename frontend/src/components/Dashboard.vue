<template>
  <div id="dashboard-app">
    <h2>Dashboard</h2>
    <p v-if="userName">Welcome, {{ userName }}!</p>

    <!-- Financial Summary -->
    <section class="dashboard-section summary-section">
      <h3 style="text-transform: capitalize;">Financial Summary ({{ summaryFilters.periodType }} - {{ summaryFilters.viewType }})</h3>

      <!-- Summary Filters -->
      <div class="summary-filters-container" style="margin-bottom: 1rem; display: flex; gap: 1rem; align-items: center; flex-wrap: wrap;">
        <div>
          <label for="summaryPeriodType" style="display: block; margin-bottom: .25rem; font-size: 0.9em;">Period Type:</label>
          <select id="summaryPeriodType" v-model="summaryFilters.periodType" @change="fetchSummary" class="form-control" style="width: auto; display: inline-block;">
            <option value="monthly">Monthly</option>
            <option value="weekly">Weekly</option>
            <option value="yearly">Yearly</option>
          </select>
        </div>

        <div v-if="summaryFilters.periodType === 'monthly'">
          <label for="summaryMonth" style="display: block; margin-bottom: .25rem; font-size: 0.9em;">Select Month:</label>
          <input type="month" id="summaryMonth" v-model="summaryFilters.selectedMonth" @change="fetchSummary" class="form-control" style="width: auto; display: inline-block;">
        </div>
        <div v-if="summaryFilters.periodType === 'weekly'">
          <label for="summaryWeekDate" style="display: block; margin-bottom: .25rem; font-size: 0.9em;">Select Date in Week:</label>
          <input type="date" id="summaryWeekDate" v-model="summaryFilters.selectedDate" @change="fetchSummary" class="form-control" style="width: auto; display: inline-block;">
        </div>
        <div v-if="summaryFilters.periodType === 'yearly'">
          <label for="summaryYear" style="display: block; margin-bottom: .25rem; font-size: 0.9em;">Select Year:</label>
          <input type="number" id="summaryYear" v-model="summaryFilters.selectedYear" @change="fetchSummary" placeholder="YYYY" class="form-control" style="width: auto; display: inline-block;">
        </div>

        <div>
          <label for="summaryViewType" style="display: block; margin-bottom: .25rem; font-size: 0.9em;">View Type:</label>
          <select id="summaryViewType" v-model="summaryFilters.viewType" @change="fetchSummary" class="form-control" style="width: auto; display: inline-block;">
            <option value="overall">Overall</option>
            <option value="income">Income Only</option>
            <option value="expenses">Expenses Only</option>
            <!-- <option value="savings">Savings</option> -->
            <!-- <option value="debts">Debts</option> -->
          </select>
        </div>
      </div>

      <div v-if="summary.loading" class="loading-message">Loading summary...</div>
      <div v-if="summary.error" class="error-message">{{ summary.error }}</div>
      <div v-if="!summary.loading && !summary.error && summary.data">
        <p v-if="summaryFilters.viewType === 'income' || summaryFilters.viewType === 'overall'">
          Total Income: <strong>{{ formatCurrency(summary.data.total_income) }}</strong>
        </p>
        <p v-if="summaryFilters.viewType === 'expenses' || summaryFilters.viewType === 'overall'">
          Total Expenses: <strong>{{ formatCurrency(summary.data.total_expenses) }}</strong>
        </p>
        <p v-if="summary.data.net_balance !== undefined"> <!-- Check if net_balance is part of the response -->
          Net Balance: <strong>{{ formatCurrency(summary.data.net_balance) }}</strong>
        </p>
      </div>
      <div v-if="!summary.loading && !summary.error && !summary.data && !summary.initialLoad">
        No summary data available for the selected period/view.
      </div>
    </section>

    <!-- Sections (Income, Expenses, Savings, Debts) -->
    <IncomeSection @open-modal="openModal" :key="sectionKeys.income" @item-changed="handleItemChange" />
    <ExpensesSection @open-modal="openModal" :key="sectionKeys.expenses" @item-changed="handleItemChange" />
    <SavingsSection @open-modal="openModal" :key="sectionKeys.savings" @item-changed="handleItemChange" />
    <DebtsSection @open-modal="openModal" :key="sectionKeys.debts" @item-changed="handleItemChange" />

    <!-- Analytics Section -->
    <section class="dashboard-section analytics-section">
      <h3>Analytics</h3>

      <!-- Expense Breakdown -->
      <div class="analytics-chart">
        <h4>Expense Breakdown (Current Month)</h4>
        <div v-if="expenseBreakdown.loading" class="loading-message">Loading expense breakdown...</div>
        <div v-if="expenseBreakdown.error" class="error-message">{{ expenseBreakdown.error }}</div>
        <div v-if="!expenseBreakdown.loading && !expenseBreakdown.error && pieChartData">
          <div style="height: 300px"> <!-- Set a height for the chart container -->
            <Pie :data="pieChartData" :options="pieChartOptions" />
          </div>
        </div>
        <div v-if="!expenseBreakdown.loading && !expenseBreakdown.error && !pieChartData && (!expenseBreakdown.data || expenseBreakdown.data.length === 0)">
          No expense data for category breakdown.
        </div>
      </div>

      <!-- Income vs. Expense Trend -->
      <div class="analytics-chart">
        <h4>Income vs. Expense Trend (Last 6 Months)</h4>
        <div v-if="incomeExpenseTrend.loading" class="loading-message">Loading trend data...</div>
        <div v-if="incomeExpenseTrend.error" class="error-message">{{ incomeExpenseTrend.error }}</div>
        <div v-if="!incomeExpenseTrend.loading && !incomeExpenseTrend.error && barChartData">
          <div style="height: 300px"> <!-- Set a height for the chart container -->
            <Bar :data="barChartData" :options="barChartOptions" />
          </div>
        </div>
        <div v-if="!incomeExpenseTrend.loading && !incomeExpenseTrend.error && !barChartData && (!incomeExpenseTrend.data || incomeExpenseTrend.data.length === 0)">
          No income/expense trend data available.
        </div>
      </div>
    </section>

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
import { ref, onMounted, reactive, computed } from 'vue'; // Ensure reactive and computed are imported
import api from '../services/api.js';
import { Pie, Bar } from 'vue-chartjs';
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement
} from 'chart.js';

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale, LinearScale, BarElement);

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

const today = new Date();
const summaryFilters = reactive({
  periodType: 'monthly', // 'monthly', 'weekly', 'yearly'
  selectedDate: today.toISOString().split('T')[0], // For daily/weekly, YYYY-MM-DD
  selectedMonth: today.toISOString().substring(0, 7), // For monthly, YYYY-MM
  selectedYear: today.getFullYear().toString(), // For yearly, YYYY
  viewType: 'overall', // 'overall', 'income', 'expenses'
});

const expenseBreakdown = reactive({
  loading: false,
  error: null,
  data: null, // To store array of { category: string, totalAmount: float64 }
});

const incomeExpenseTrend = reactive({
  loading: false,
  error: null,
  data: null, // To store array of { month: string, totalIncome: float64, totalExpenses: float64 }
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
  summary.initialLoad = false; // Keep this to manage "No data" message

  let endpoint = `/summary/${summaryFilters.periodType}`;
  let dateParam = '';

  if (summaryFilters.periodType === 'monthly') {
    dateParam = summaryFilters.selectedMonth; // YYYY-MM
  } else if (summaryFilters.periodType === 'weekly') {
    dateParam = summaryFilters.selectedDate; // YYYY-MM-DD
  } else if (summaryFilters.periodType === 'yearly') {
    dateParam = summaryFilters.selectedYear; // YYYY
  }

  if (!dateParam && (summaryFilters.periodType === 'monthly' || summaryFilters.periodType === 'weekly' || summaryFilters.periodType === 'yearly')) {
    summary.error = "Please select a valid date/period.";
    summary.loading = false;
    summary.data = null;
    return;
  }

  try {
    const response = await api.get(`${endpoint}?date=${dateParam}&view=${summaryFilters.viewType}`);
    summary.data = response.data;
  } catch (err) {
    console.error("Error fetching summary:", err);
    summary.error = "Failed to load financial summary. " + (err.response?.data?.error || err.message);
    summary.data = null; // Clear data on error
  } finally {
    summary.loading = false;
  }
};

const fetchExpenseBreakdown = async () => {
  expenseBreakdown.loading = true;
  expenseBreakdown.error = null;
  try {
    const response = await api.get('/analytics/expense-categories'); // Using api.js service
    expenseBreakdown.data = response.data;
  } catch (err) {
    console.error("Error fetching expense breakdown:", err);
    expenseBreakdown.error = "Failed to load expense breakdown. " + (err.response?.data?.error || err.message);
    expenseBreakdown.data = null;
  } finally {
    expenseBreakdown.loading = false;
  }
};

const fetchIncomeExpenseTrend = async () => {
  incomeExpenseTrend.loading = true;
  incomeExpenseTrend.error = null;
  try {
    // Assuming default 6 months from backend, or add query param: /analytics/income-expense-trend?months=6
    const response = await api.get('/analytics/income-expense-trend');
    incomeExpenseTrend.data = response.data;
  } catch (err) {
    console.error("Error fetching income/expense trend:", err);
    incomeExpenseTrend.error = "Failed to load income/expense trend. " + (err.response?.data?.error || err.message);
    incomeExpenseTrend.data = null;
  } finally {
    incomeExpenseTrend.loading = false;
  }
};


onMounted(() => {
  fetchSummary();
  fetchExpenseBreakdown();
  fetchIncomeExpenseTrend();
  // Fetch other initial data (income, expenses, etc.)
});

// --- Chart Computed Properties ---
const pieChartData = computed(() => {
  if (!expenseBreakdown.data || expenseBreakdown.data.length === 0) {
    return null;
  }
  // Ensure field names match the backend (Category, TotalAmount)
  return {
    labels: expenseBreakdown.data.map(item => item.category),
    datasets: [
      {
        backgroundColor: [
          '#4A90E2', // Primary Blue
          '#50E3C2', // Teal/Turquoise
          '#F5A623', // Orange
          '#BD10E0', // Purple
          '#7ED321', // Lime Green
          '#4A4A4A', // Dark Gray
          '#E0E0E0', // Light Gray
          '#F8E71C', // Yellow
          '#D0021B', // Red
          '#007bff'  // Another Blue
        ],
        borderColor: '#FFFFFF',
        borderWidth: 2,
        data: expenseBreakdown.data.map(item => item.total_amount),
      },
    ],
  };
});

const pieChartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom', // Or 'top', depending on space and preference
      labels: {
        font: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif' }, // Consistent font
        padding: 20, // Add padding to legend items
      }
    },
    title: {
      display: true, // Display chart title (already has section title, but can be more specific)
      text: 'Expense Breakdown by Category',
      font: {
        size: 16,
        family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
      },
      padding: { top: 10, bottom: 10 }
    },
    tooltip: {
      backgroundColor: '#4A4A4A', // Darker tooltip background
      titleFont: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', size: 14 },
      bodyFont: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', size: 12 },
      callbacks: {
        label: function(context) {
          let label = context.label || '';
          if (label) {
            label += ': ';
          }
          if (context.parsed !== null) {
            // Assuming formatCurrency is accessible here or define a similar one
            label += new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(context.parsed);
          }
          return label;
        }
      }
    }
  }
});

const barChartData = computed(() => {
  if (!incomeExpenseTrend.data || incomeExpenseTrend.data.length === 0) {
    return null;
  }
  // Ensure field names match the backend (Month, TotalIncome, TotalExpenses)
  return {
    labels: incomeExpenseTrend.data.map(item => item.month),
    datasets: [
      {
        label: 'Total Income',
        backgroundColor: '#28a745', // Green (from new palette for success/income)
        borderColor: '#28a745',
        borderWidth: 1,
        data: incomeExpenseTrend.data.map(item => item.total_income),
      },
      {
        label: 'Total Expenses',
        backgroundColor: '#dc3545', // Red (from new palette for error/expense)
        borderColor: '#dc3545',
        borderWidth: 1,
        data: incomeExpenseTrend.data.map(item => item.total_expenses),
      },
    ],
  };
});

const barChartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'top',
      labels: {
        font: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif' },
        padding: 10,
      }
    },
    title: {
      display: true,
      text: 'Monthly Income vs. Expense Trend',
      font: {
        size: 16,
        family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif'
      },
      padding: { top: 10, bottom: 20 } // More bottom padding for title
    },
    tooltip: {
      backgroundColor: '#4A4A4A',
      titleFont: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', size: 14 },
      bodyFont: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif', size: 12 },
      callbacks: {
        label: function(context) {
          let label = context.dataset.label || '';
          if (label) {
            label += ': ';
          }
          if (context.parsed.y !== null) {
            label += new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(context.parsed.y);
          }
          return label;
        }
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      grid: {
        color: '#e0e0e0', // Lighter grid lines
        borderColor: '#cccccc' // Border for the axis line
      },
      ticks: {
        font: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif' },
        callback: function(value) { // Format Y-axis ticks as currency
          return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(value);
        }
      }
    },
    x: {
      grid: {
        display: false // Hide vertical grid lines for a cleaner look
      },
      ticks: {
        font: { family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif' }
      }
    }
  }
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

.analytics-section {
  /* Similar to other sections or define new styles */
  background-color: #f0f7ff; /* Light blue background for distinction */
}
.analytics-chart {
  margin-top: 20px;
  padding: 15px; /* Increased padding */
  border: 1px solid #dce9f5; /* Lighter border */
  border-radius: 4px;
  background-color: #fff; /* White background for chart area */
}
.analytics-chart h4 {
  margin-top: 0;
  font-size: 1.1em;
  color: #337ab7; /* Theme color for heading */
}
.loading-message, .error-message {
  padding: 10px;
  border-radius: 4px;
  margin-top: 10px;
}
.loading-message {
  background-color: #e9f5ff;
  color: #31708f;
}
.error-message {
  background-color: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}
/* Placeholder styling for pre tag */
.analytics-chart pre {
  background-color: #f5f5f5;
  padding: 10px;
  border-radius: 3px;
  border: 1px solid #ccc;
  max-height: 200px; /* Limit height */
  overflow-y: auto; /* Add scroll for overflow */
}

</style>