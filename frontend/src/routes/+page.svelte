<script>
  import { onMount } from 'svelte';

  // --- State Management ---
  // These variables hold the current state of our UI.

  let isLoading = true; // Is the page currently loading initial data?
  let errorMessage = ''; // Holds any error message we want to display.
  let currentRule = null; // Will hold the QoS rule object fetched from the backend.

  // These variables are bound to the form inputs.
  let selectedInterface = ''; // The interface selected in the dropdown.
  let inputBandwidth = 50; // The bandwidth value in the input box.

  // --- API Functions ---
  // These functions talk to our Go backend.

  /**
   * Fetches the current QoS rule from the backend and updates the UI.
   */
  async function fetchCurrentRule() {
    isLoading = true;
    errorMessage = '';
    try {
      const response = await fetch('http://localhost:8080/api/qos/rules');
      if (!response.ok) {
        // Handle cases where the API returns an error, e.g., no rule found.
        const errorData = await response.json().catch(() => null);
        throw new Error(errorData?.message || `Failed to fetch rule: ${response.statusText}`);
      }
      currentRule = await response.json();
    } catch (error) {
      console.error(error);
      // Since a 404 might just mean no rule is set, we handle it gracefully.
      currentRule = null; 
      // We could show a more specific error if needed: errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  /**
   * Applies a new QoS rule based on the form inputs.
   */
  async function applyRule() {
    if (!selectedInterface || !inputBandwidth) {
      errorMessage = 'Please select an interface and enter a bandwidth.';
      return;
    }
    
    isLoading = true;
    errorMessage = '';
    
    try {
      const response = await fetch('http://localhost:8080/api/qos/rules', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          interface: selectedInterface,
          algorithm: 'cake', // For now, we hardcode 'cake'
          settings: {
            bandwidth_mbit: Number(inputBandwidth),
          },
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to apply rule.');
      }
      
      // After successfully applying, refresh the status.
      await fetchCurrentRule();

    } catch (error) {
      console.error(error);
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  /**
   * Deletes the QoS rule for the selected interface.
   */
  async function deleteRule() {
    if (!selectedInterface) {
      errorMessage = 'Please select an interface to delete its rule.';
      return;
    }

    isLoading = true;
    errorMessage = '';
    
    try {
      const response = await fetch(`http://localhost:8080/api/qos/rules?interface=${selectedInterface}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to delete rule.');
      }
      
      // After successfully deleting, refresh the status.
      await fetchCurrentRule();

    } catch (error) {
      console.error(error);
      errorMessage = error.message;
    } finally {
      isLoading = false;
    }
  }

  // When the component first loads, fetch the initial QoS status.
  onMount(() => {
    fetchCurrentRule();
  });
</script>

<main class="min-h-screen bg-gray-50 flex flex-col items-center p-4 sm:p-8">
  <div class="w-full max-w-2xl bg-white rounded-xl shadow-lg p-6 sm:p-8 space-y-6">
    
    <!-- Header Section -->
    <header class="text-center border-b pb-4">
      <h1 class="text-3xl font-bold text-gray-800">✈️ NetPilot</h1>
      <p class="text-md text-gray-500 mt-1">CAKE QoS Control Panel</p>
    </header>

    <!-- Status Display Section -->
    <section>
      <h2 class="text-xl font-semibold text-gray-700 mb-3">Current Status</h2>
      <div class="bg-gray-100 p-4 rounded-lg text-gray-600 min-h-[6rem] flex items-center justify-center">
        {#if isLoading}
          <p>Loading status...</p>
        {:else if errorMessage}
          <p class="text-red-500 font-medium">{errorMessage}</p>
        {:else if currentRule}
          <div class="text-left w-full space-y-1">
            <p><strong>Interface:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{currentRule.interface}</code></p>
            <p><strong>Algorithm:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{currentRule.algorithm}</code></p>
            <p><strong>Bandwidth:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{currentRule.settings.bandwidth}</code></p>
          </div>
        {:else}
          <p>No active QoS rule found.</p>
        {/if}
      </div>
    </section>

    <!-- Configuration Section -->
    <section>
      <h2 class="text-xl font-semibold text-gray-700 mb-3">Configuration</h2>
      <!-- We use on:submit|preventDefault to stop the browser from reloading the page -->
      <form class="space-y-4" on:submit|preventDefault={applyRule}>
        <div>
          <label for="interface-select" class="block text-sm font-medium text-gray-700">Network Interface</label>
          <!-- `bind:value` creates a two-way data binding with the `selectedInterface` variable -->
          <select id="interface-select" bind:value={selectedInterface} class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
            <option disabled value="">-- Please choose an interface --</option>
            <!-- TODO: This should be populated dynamically from a backend API endpoint -->
            <option value="eth0">eth0</option>
            <option value="enp1s0">enp1s0</option>
          </select>
        </div>

        <div>
          <label for="bandwidth-input" class="block text-sm font-medium text-gray-700">Upload Bandwidth (Mbit/s)</label>
          <input type="number" id="bandwidth-input" bind:value={inputBandwidth} placeholder="e.g., 50" class="mt-1 block w-full pl-3 pr-3 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
        </div>

        <div class="flex flex-col sm:flex-row gap-3 pt-2">
          <button type="submit" class="w-full inline-flex justify-center items-center px-4 py-2 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            Apply Rule
          </button>
          <button type="button" on:click={deleteRule} class="w-full inline-flex justify-center items-center px-4 py-2 border border-gray-300 text-base font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            Delete Rule
          </button>
        </div>
      </form>
    </section>
  </div>
</main>