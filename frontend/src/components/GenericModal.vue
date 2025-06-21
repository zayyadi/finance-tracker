<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-container"> <!-- Changed from modal-content -->
      <div class="modal-header">
        <h3>{{ modalTitle }}</h3> <!-- Changed h4 to h3 to match new CSS -->
        <button @click="$emit('close')" class="modal-close-btn" aria-label="Close modal">&times;</button>
      </div>

      <div class="modal-body">
        <form @submit.prevent="submitForm" id="modalMainForm">
          <div v-if="internalError" class="error-message modal-error-message">{{ internalError }}</div>

          <!-- Common fields for Income/Expense -->
        <template v-if="mode === 'income' || mode === 'expenses'">
          <div>
            <label :for="mode + '-amount'">Amount:</label>
            <input type="number" :id="mode + '-amount'" v-model="formData.amount" step="0.01" required>
          </div>
          <div>
            <label :for="mode + '-category'">Category:</label>
            <input type="text" :id="mode + '-category'" v-model="formData.category" required>
          </div>
          <div>
            <label :for="mode + '-date'">Date:</label>
            <input type="date" :id="mode + '-date'" v-model="formData.date" required>
          </div>
          <div>
            <label :for="mode + '-note'">Note:</label>
            <textarea :id="mode + '-note'" v-model="formData.note"></textarea>
          </div>
        </template>

        <!-- Savings Fields -->
        <template v-if="mode === 'savings'">
          <div>
            <label :for="mode + '-goal_name'">Goal Name:</label>
            <input type="text" :id="mode + '-goal_name'" v-model="formData.goal_name" required>
          </div>
          <div>
            <label :for="mode + '-goal_amount'">Goal Amount:</label>
            <input type="number" :id="mode + '-goal_amount'" v-model="formData.goal_amount" step="0.01" required>
          </div>
          <div>
            <label :for="mode + '-current_amount'">Current Amount:</label>
            <input type="number" :id="mode + '-current_amount'" v-model="formData.current_amount" step="0.01">
          </div>
          <div>
            <label :for="mode + '-target_date'">Target Date (Optional):</label>
            <input type="date" :id="mode + '-target_date'" v-model="formData.target_date">
          </div>
           <div>
            <label :for="mode + '-notes'">Notes:</label>
            <textarea :id="mode + '-notes'" v-model="formData.notes"></textarea>
          </div>
        </template>

        <!-- Debt Fields -->
        <template v-if="mode === 'debts'">
            <div>
              <label :for="mode + '-debtor_name'">Debtor/Creditor Name:</label>
              <input type="text" :id="mode + '-debtor_name'" v-model="formData.debtor_name" required>
            </div>
            <div>
              <label :for="mode + '-description'">Description:</label>
              <textarea :id="mode + '-description'" v-model="formData.description"></textarea>
            </div>
          <div>
            <label :for="mode + '-amount-debt'">Amount:</label>
            <input type="number" :id="mode + '-amount-debt'" v-model="formData.amount" step="0.01" required>
          </div>
          <div>
            <label :for="mode + '-due_date'">Due Date:</label>
            <input type="date" :id="mode + '-due_date'" v-model="formData.due_date" required>
          </div>
          <div>
            <label :for="mode + '-status'">Status:</label>
            <select :id="mode + '-status'" v-model="formData.status" required>
              <option value="Pending">Pending</option>
              <option value="Paid">Paid</option>
              <option value="Overdue">Overdue</option>
            </select>
          </div>
        </template>

          <!-- Form-level actions moved into modal-footer -->
        </form>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-light" @click="$emit('close')" :disabled="props.processing || localLoading">Cancel</button>
        <button type="submit" form="modalMainForm" class="btn btn-primary" :disabled="props.processing || localLoading">
          <span v-if="props.processing || localLoading">Saving...</span>
          <span v-else>Save</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
// Add id="modalMainForm" to the <form> tag
// The script part does not need changes for this subtask, only template and style import.
import { ref, watch, computed, defineProps, defineEmits, toRefs } from 'vue';

const props = defineProps({
  mode: { type: String, required: true }, // 'income', 'expenses', 'savings', 'debts'
  itemData: { type: Object, default: () => ({}) },
  isEditing: { type: Boolean, default: false },
  processing: { type: Boolean, default: false } // Prop to indicate parent is processing
});

const emit = defineEmits(['close', 'save']);

const formData = ref({});
const loading = ref(false);
const internalError = ref(null);
const localLoading = ref(false); // For any purely internal async work, if needed

const modalTitle = computed(() => {
  const action = props.isEditing ? 'Edit' : 'Add';
  const itemType = props.mode.charAt(0).toUpperCase() + props.mode.slice(1);
  return `${action} ${itemType}`;
});

watch(() => props.itemData, (newData) => {
  // Ensure date fields are correctly formatted for <input type="date">
  const dataToEdit = { ...newData };
  ['date', 'target_date', 'due_date', 'start_date'].forEach(dateField => {
    if (dataToEdit[dateField]) {
      dataToEdit[dateField] = new Date(dataToEdit[dateField]).toISOString().split('T')[0];
    }
  });
  formData.value = { ...dataToEdit };
}, { immediate: true, deep: true });

const submitForm = () => {
  internalError.value = null;
  // localLoading.value = true; // Set this if modal does its own async validation before emitting
  // The actual API call will be handled by the parent (Dashboard.vue)
  // This component just emits the data.
  // Add validation here if needed before emitting.
  
  // Convert amount fields back to numbers if they became strings
  ['amount', 'goal_amount', 'current_amount'].forEach(field => {
    if (formData.value[field] !== undefined && formData.value[field] !== null) {
      formData.value[field] = parseFloat(formData.value[field]);
    }
  });

  emit('save', { ...formData.value });
  // Parent will handle API call and closing modal on success/failure
  // localLoading.value = false;
};

// It's important to add an id to the form tag to link the submit button if it's outside the form.
// e.g., <form @submit.prevent="submitForm" id="modalMainForm">
// This is handled in the template change.
</script>

<style scoped src="../assets/modal-styles.css"></style>
<!-- Removed inlined styles, now using external modal-styles.css -->
```

**i. `frontend/src/services/api.js`**
A simple Axios wrapper for API calls.

```diff
Unchanged linesimport axios from 'axios';

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
