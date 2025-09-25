<script>
  import { onMount } from 'svelte';

  // --- State Management ---
  let status = { isLoading: true, message: 'Initializing...', isError: false, rule: null };
  let selectedInterface = '', inputBandwidth = 50, availableInterfaces = [];
  let selectedAlgorithm = 'cake';
  const algorithms = [
    { value: 'cake', label: 'CAKE (Recommended)', needsBandwidth: true },
    { value: 'fq_codel', label: 'FQ Codel', needsBandwidth: false },
    { value: 'sfq', label: 'SFQ (Stochastic Fairness)', needsBandwidth: false },
    { value: 'tbf', label: 'TBF (Simple Rate Limiter)', needsBandwidth: true },
  ];
  let showAdvanced = false;

  // --- API Functions (Defined only ONCE) ---

  async function querySystemState(iface) {
    if (!iface) return;
    status = { isLoading: true, message: `Checking status for ${iface}...`, isError: false, rule: null };
    try {
      const response = await fetch(`/api/qos/rules?interface=${iface}`);
      if (!response.ok && response.status !== 204) {
        const errorText = await response.text();
        throw new Error(`API Error: ${response.status} - ${errorText}`);
      }
      
      if (response.status === 204) {
        status = { isLoading: false, message: `No active rule managed by NetPilot on ${iface}.`, isError: false, rule: null };
      } else {
        const rule = await response.json();
        status = { isLoading: false, message: `Active rule found on ${iface}.`, isError: false, rule: rule };
        if (rule.algorithm && algorithms.some(a => a.value === rule.algorithm)) {
          selectedAlgorithm = rule.algorithm;
        }
        if (rule.settings && rule.settings.bandwidth) {
          inputBandwidth = parseInt(rule.settings.bandwidth) || 50;
        } else if (rule.settings && rule.settings.rate) {
          inputBandwidth = parseInt(rule.settings.rate) || 50;
        }
      }
    } catch (error) {
      console.error('Fetch Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  async function applyRule() {
    if (!selectedInterface) return;
    const currentAlgo = algorithms.find(a => a.value === selectedAlgorithm);
    if (currentAlgo && currentAlgo.needsBandwidth && (!inputBandwidth || inputBandwidth <= 0)) {
       status = { ...status, isLoading: false, message: 'A positive bandwidth is required.', isError: true };
       return;
    }
    status = { isLoading: true, message: `Applying ${selectedAlgorithm} rule...`, isError: false, rule: status.rule };
    try {
      const response = await fetch('/api/qos/rules', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          interface: selectedInterface,
          algorithm: selectedAlgorithm,
          settings: { bandwidth_mbit: Number(inputBandwidth) },
        }),
      });
      const result = await response.json();
      if (!response.ok) throw new Error(result.message || 'Failed to apply rule');
      await querySystemState(selectedInterface);
    } catch (error) {
      console.error('Apply Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  async function resetToDefault() {
    if (!selectedInterface) return;
    status = { isLoading: true, message: `Resetting QoS on ${selectedInterface}...`, isError: false, rule: status.rule };
    try {
      const response = await fetch(`/api/qos/rules?interface=${selectedInterface}`, { method: 'DELETE' });
      const result = await response.json();
      if (!response.ok) throw new Error(result.message || 'Failed to delete rule');
      await querySystemState(selectedInterface);
    } catch (error) {
      console.error('Delete Rule Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  }

  // --- NEW: A helper function to parse CAKE's complex options string ---
  function parseCakeOptions(optionsStr) {
    if (!optionsStr) return null;
    const options = optionsStr.split(/\s+/);
    const details = {
      'Isolation': [],
      'DSCP Handling': [],
      'ACK Filter': [],
      'Special Features': [],
      'Timings & Overhead': [],
    };
    
    for (let i = 0; i < options.length; i++) {
      const opt = options[i];
      if (opt.includes('isolate')) details['Isolation'].push(opt);
      else if (opt.includes('diffserv')) details['DSCP Handling'].push(opt);
      else if (opt.includes('wash')) details['DSCP Handling'].push(opt);
      else if (opt.includes('ack-filter')) details['ACK Filter'].push(opt);
      else if (opt.includes('gso')) details['Special Features'].push(opt);
      else if (opt === 'nat' || opt === 'nonat') details['Isolation'].push(opt);
      else if (opt === 'rtt' || opt === 'overhead') {
        details['Timings & Overhead'].push(`${opt} ${options[++i]}`);
      }
    }
    return details;
  }

  // --- NEW: A computed property that automatically parses cake options ---
  $: cakeDetails = status.rule?.algorithm === 'cake' ? parseCakeOptions(status.rule.settings.options) : null;
  
  // --- Lifecycle & Reactivity ---
  onMount(async () => {
    status = { isLoading: true, message: 'Fetching network interfaces...', isError: false, rule: null };
    try {
      const response = await fetch('/api/interfaces');
      if (!response.ok) throw new Error('Failed to fetch interfaces');
      availableInterfaces = await response.json();
      if (availableInterfaces.length > 0) {
        selectedInterface = availableInterfaces[0];
      } else {
        status = { isLoading: false, message: 'No network interfaces found.', isError: true, rule: null };
      }
    } catch (error) {
      console.error('Fetch Interfaces Error:', error);
      status = { isLoading: false, message: error.message, isError: true, rule: null };
    }
  });
  
  $: if (selectedInterface) {
    querySystemState(selectedInterface);
  }

</script>

<main class="min-h-screen bg-gray-100 flex items-center justify-center p-4">
  <div class="w-full max-w-2xl bg-white rounded-xl shadow-lg p-8 space-y-6">
    
    <!-- ... (Header and Configuration sections are perfect, no changes needed) ... -->
     <header class="text-center">
      <h1 class="text-3xl font-bold text-gray-900 flex items-center justify-center gap-2">✈️ NetPilot</h1>
      <p class="text-gray-500 mt-1">Linux QoS Control Panel</p>
    </header>
    <div class="border-t"></div>
    <section class="space-y-4">
      <div>
        <label for="interface-select" class="block text-sm font-medium text-gray-700">Network Interface</label>
        <select id="interface-select" bind:value={selectedInterface} class="mt-1 block w-full pl-3 pr-10 py-2 border-gray-300 rounded-md" disabled={availableInterfaces.length === 0}>
            {#if availableInterfaces.length === 0}<option value="">Loading...</option>{:else}{#each availableInterfaces as iface}<option value={iface}>{iface}</option>{/each}{/if}
        </select>
      </div>
      <div>
        <label for="algorithm-select" class="block text-sm font-medium text-gray-700">QoS Algorithm to Apply</label>
        <select id="algorithm-select" bind:value={selectedAlgorithm} class="mt-1 block w-full pl-3 pr-10 py-2 border-gray-300 rounded-md">
          {#each algorithms as algo}<option value={algo.value}>{algo.label}</option>{/each}
        </select>
      </div>
      {#if algorithms.find(a => a.value === selectedAlgorithm)?.needsBandwidth}
        <div>
          <label for="bandwidth-input" class="block text-sm font-medium text-gray-700">Bandwidth (Mbit/s)</label>
          <input type="number" id="bandwidth-input" bind:value={inputBandwidth} class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md" min="1">
        </div>
      {/if}
      <div class="relative flex items-start pt-2">
        <div class="flex items-center h-5">
          <input id="advanced-toggle" type="checkbox" bind:checked={showAdvanced} class="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded">
        </div>
        <div class="ml-3 text-sm">
          <label for="advanced-toggle" class="font-medium text-gray-700">Show Advanced Options</label>
        </div>
      </div>
      {#if showAdvanced}
        <div class="mt-4 p-4 border-l-4 border-yellow-400 bg-yellow-50"><p class="text-sm text-yellow-700">Advanced options are not yet implemented.</p></div>
      {/if}
    </section>

    <!-- 【核心进化】Status Section - now with beautiful, structured CAKE details -->
    <section class="bg-gray-50 p-4 rounded-lg min-h-[150px] flex flex-col justify-center">
      <h3 class="text-center font-semibold text-gray-700 mb-3">Current System Status on {selectedInterface || '...'}</h3>
      <div class="text-left text-sm">
        {#if status.isLoading}
          <p class="text-center text-yellow-600">{status.message}</p>
        {:else if status.isError}
          <p class="text-center text-red-600 font-semibold">{status.message}</p>
        {:else if status.rule && status.rule.algorithm}
          <div class="space-y-2">
            <div class="flex items-center">
              <span class="w-28 font-semibold text-gray-600">Algorithm:</span>
              <span class="px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded-full">{status.rule.algorithm}</span>
            </div>

            <!-- This is the "smart" part -->
            {#if status.rule.algorithm === 'cake' && cakeDetails}
              <div class="pt-2">
                <h4 class="font-semibold text-gray-600 mb-1">Parameters:</h4>
                <div class="pl-4 border-l-2 border-gray-200 space-y-1">
                  <div class="flex"><span class="w-24 text-gray-500">Bandwidth:</span><code class="text-indigo-700">{status.rule.settings.bandwidth}</code></div>
                  {#each Object.entries(cakeDetails) as [category, opts]}
                    {#if opts.length > 0}
                      <div class="flex"><span class="w-24 text-gray-500">{category}:</span><code class="text-indigo-700">{opts.join(' ')}</code></div>
                    {/if}
                  {/each}
                </div>
              </div>
            {:else if status.rule.settings && Object.keys(status.rule.settings).length > 0}
              <!-- Fallback for other algorithms like TBF, SFQ -->
              <div class="pt-2">
                <h4 class="font-semibold text-gray-600 mb-1">Parameters:</h4>
                <div class="pl-4 border-l-2 border-gray-200 space-y-1">
                  {#each Object.entries(status.rule.settings) as [key, value]}
                    <div class="flex items-baseline">
                      <span class="w-24 text-gray-500 capitalize">{key.replace('_', ' ')}:</span>
                      <code class="text-indigo-700">{value}</code>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {:else}
          <p class="text-center text-blue-700">{status.message || `System default is active.`}</p>
        {/if}
      </div>
    </section>
    
    <!-- ... (Buttons section is perfect, no changes needed) ... -->
    <section class="pt-2 flex flex-col sm:flex-row gap-3">
      <button on:click={applyRule} class="w-full py-2 px-4 rounded-md font-medium text-white bg-indigo-600 hover:bg-indigo-700" disabled={status.isLoading}>Apply '{selectedAlgorithm}'</button>
      <button on:click={resetToDefault} class="w-full py-2 px-4 rounded-md font-medium text-white bg-gray-500 hover:bg-gray-600" disabled={status.isLoading || !status.rule}>Reset to Default</button>
    </section>
  </div>
</main>