<script>
  import { onMount } from 'svelte';

  // --- State Management ---
  // We use a single object to hold the UI's state for clarity.
  let status = {
    isLoading: true,
    message: 'Initializing...',
    isError: false,
    rule: null, // This will hold the active QoS rule object if it exists
  };
  
  // --- Form Inputs ---
  let selectedInterface = '';
  let inputBandwidth = 50;
  let availableInterfaces = [];

  // --- API Functions ---

  async function fetchInterfaces() {
    try {
      const response = await fetch('http://localhost:8080/api/interfaces');
      if (!response.ok) {
        throw new Error(`Failed to fetch interfaces: ${response.statusText}`);
      }
      
      availableInterfaces = await response.json();
      
      if (availableInterfaces.length > 0) {
        // Automatically select the first interface found
        selectedInterface = availableInterfaces[0];
      } else {
        // Handle the case where no network interfaces are found
        status = { isLoading: false, message: 'No suitable network interfaces found.', isError: true, rule: null };
      }
    } catch (error) {
      console.error('Fetch Interfaces Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  async function fetchRuleForInterface(iface) {
    if (!iface) return;

    status = { isLoading: true, message: `Checking status for ${iface}...`, isError: false, rule: null };
    
    try {
      const response = await fetch(`http://localhost:8080/api/qos/rules?interface=${iface}`);
      
      if (response.status === 204) {
        // 204 No Content is a success signal that means no rule is currently set.
        status = { isLoading: false, message: `No active QoS rule on ${iface}.`, isError: false, rule: null };
      } else if (response.ok) {
        const rule = await response.json();
        status = { isLoading: false, message: `Active rule found on ${iface}.`, isError: false, rule: rule };
        // Update the input box to reflect the current setting
        if (rule.settings && rule.settings.bandwidth) {
          inputBandwidth = parseInt(rule.settings.bandwidth) || 50;
        }
      } else {
        const errorText = await response.text();
        throw new Error(`API Error: ${response.status} - ${errorText}`);
      }
    } catch (error) {
      console.error('Fetch Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  async function applyRule() {
    if (!selectedInterface || !inputBandwidth || inputBandwidth <= 0) {
      status = { isLoading: false, message: 'Interface and a positive bandwidth are required.', isError: true, rule: status.rule };
      return;
    }

    status = { isLoading: true, message: `Applying ${inputBandwidth}Mbit rule to ${selectedInterface}...`, isError: false, rule: status.rule };
    
    try {
      const response = await fetch('http://localhost:8080/api/qos/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          interface: selectedInterface,
          algorithm: 'cake',
          settings: { bandwidth_mbit: Number(inputBandwidth) },
        }),
      });
      
      const result = await response.json();
      if (!response.ok) throw new Error(result.message || 'Failed to apply rule');

      // After applying, immediately refresh the status to show the new state
      await fetchRuleForInterface(selectedInterface);
    } catch (error) {
      console.error('Apply Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  async function deleteRule() {
    if (!selectedInterface) return;

    status = { isLoading: true, message: `Deleting rule from ${selectedInterface}...`, isError: false, rule: status.rule };
    
    try {
      const response = await fetch(`http://localhost:8080/api/qos/rules?interface=${selectedInterface}`, {
        method: 'DELETE',
      });

      const result = await response.json();
      if (!response.ok) throw new Error(result.message || 'Failed to delete rule');
      
      // After deleting, immediately refresh the status
      await fetchRuleForInterface(selectedInterface);
    } catch (error) {
      console.error('Delete Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  // --- Lifecycle & Reactivity ---

  // When the component is first mounted in the browser, run the full initialization sequence.
  onMount(async () => {
    await fetchInterfaces();
    // After interfaces are fetched, the selectedInterface will be set,
    // which will then trigger the reactive statement below to fetch the rule.
  });
  
  // This reactive statement triggers whenever `selectedInterface` changes.
  // This handles both the initial setting in onMount and subsequent user selections.
  $: if (selectedInterface) {
      fetchRuleForInterface(selectedInterface);
  }

</script>

<main class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
  <div class="w-full max-w-lg bg-white rounded-xl shadow-lg p-8 space-y-6">
    
    <header class="text-center">
      <h1 class="text-3xl font-bold text-gray-900 flex items-center justify-center gap-2">
        ✈️ NetPilot
      </h1>
      <p class="text-gray-500 mt-1">CAKE QoS Control Panel</p>
    </header>

    <div class="border-t"></div>

    <!-- Configuration Section -->
    <section class="space-y-4">
      <div>
        <label for="interface-select" class="block text-sm font-medium text-gray-700">Network Interface</label>
        <select 
          id="interface-select" 
          bind:value={selectedInterface} 
          class="mt-1 block w-full pl-3 pr-10 py-2 border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500" 
          disabled={availableInterfaces.length === 0}
        >
          {#if availableInterfaces.length === 0}
            <option value="">-- Loading interfaces... --</option>
          {:else}
            <option disabled value="">-- Select an interface --</option>
            {#each availableInterfaces as iface}
              <option value={iface}>{iface}</option>
            {/each}
          {/if}
        </select>
      </div>

      <div>
        <label for="bandwidth-input" class="block text-sm font-medium text-gray-700">Bandwidth (Mbit/s)</label>
        <input 
          type="number" 
          id="bandwidth-input" 
          bind:value={inputBandwidth} 
          class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500" 
          min="1"
        >
      </div>
    </section>

    <!-- Status & Actions Section -->
    <section class="bg-gray-50 p-4 rounded-lg min-h-[120px] flex flex-col justify-center">
      <h3 class="text-center font-semibold text-gray-700 mb-2">Status</h3>
      <div class="text-center text-sm">
        {#if status.isLoading}
          <p class="text-yellow-600">{status.message}</p>
        {:else if status.isError}
          <p class="text-red-600 font-semibold">{status.message}</p>
        {:else if status.rule}
          <div class="text-green-700 space-y-1">
            <p>Active rule found on <strong>{status.rule.interface}</strong>.</p>
            <p>Bandwidth is set to <strong>{status.rule.settings.bandwidth}</strong>.</p>
          </div>
        {:else}
          <p class="text-blue-700">{status.message}</p>
        {/if}
      </div>
    </section>

    <!-- Buttons are now shown based on the status -->
    <section class="pt-2">
      {#if !status.isLoading}
        {#if status.rule}
          <button on:click={deleteRule} class="w-full py-2 px-4 rounded-md font-medium text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500">
            Delete Rule
          </button>
        {:else if !status.isError}
          <button on:click={applyRule} class="w-full py-2 px-4 rounded-md font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500">
            Apply Rule
          </button>
        {/if}
      {/if}
    </section>

  </div>
</main>