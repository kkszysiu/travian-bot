<script lang="ts">
  import Sidebar from './Sidebar.svelte'
  import NoAccountTab from '../tabs/NoAccountTab.svelte'
  import AddAccountTab from '../tabs/AddAccountTab.svelte'
  import AddAccountsTab from '../tabs/AddAccountsTab.svelte'
  import EditAccountTab from '../tabs/EditAccountTab.svelte'
  import AccountSettingTab from '../tabs/AccountSettingTab.svelte'
  import VillageTab from '../tabs/VillageTab.svelte'
  import FarmingTab from '../tabs/FarmingTab.svelte'
  import DebugTab from '../tabs/DebugTab.svelte'
  import { selectedAccountId } from '../../stores/accounts'
  import { t } from '../../i18n'

  let activeTab = $state<string>('none')
  let selectedId = $state<number | null>(null)
  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)

  selectedAccountId.subscribe(v => selectedId = v)
  t.subscribe(v => tr = v)

  function handleTabChange(tab: string) {
    activeTab = tab
  }

  const tabIds = ['settings', 'villages', 'farming', 'editAccount', 'debug'] as const
  const tabLabelKeys: Record<string, string> = {
    settings: 'tabs.settings',
    villages: 'tabs.villages',
    farming: 'tabs.farming',
    editAccount: 'tabs.editAccount',
    debug: 'tabs.debug',
  }
</script>

<div class="flex h-screen">
  <!-- Sidebar: ~18% width -->
  <div class="w-56 flex-shrink-0">
    <Sidebar onTabChange={handleTabChange} />
  </div>

  <!-- Main Content: ~82% width -->
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Tab Bar -->
    {#if selectedId != null && !['addAccount', 'addAccounts', 'none'].includes(activeTab)}
      <div class="flex border-b border-gray-200 bg-white">
        {#each tabIds as tabId}
          <button
            class="px-4 py-2 text-sm font-medium transition-colors"
            class:text-blue-600={activeTab === tabId}
            class:border-b-2={activeTab === tabId}
            class:border-blue-600={activeTab === tabId}
            class:text-gray-500={activeTab !== tabId}
            class:hover:text-gray-700={activeTab !== tabId}
            onclick={() => activeTab = tabId}
          >
            {tr(tabLabelKeys[tabId])}
          </button>
        {/each}
      </div>
    {/if}

    <!-- Tab Content -->
    <div class="flex-1 overflow-auto bg-white p-4">
      {#if activeTab === 'none'}
        <NoAccountTab />
      {:else if activeTab === 'addAccount'}
        <AddAccountTab onDone={() => { activeTab = 'none' }} />
      {:else if activeTab === 'addAccounts'}
        <AddAccountsTab onDone={() => { activeTab = 'none' }} />
      {:else if activeTab === 'settings' && selectedId != null}
        <AccountSettingTab accountId={selectedId} />
      {:else if activeTab === 'villages' && selectedId != null}
        <VillageTab accountId={selectedId} />
      {:else if activeTab === 'farming' && selectedId != null}
        <FarmingTab accountId={selectedId} />
      {:else if activeTab === 'editAccount' && selectedId != null}
        <EditAccountTab accountId={selectedId} />
      {:else if activeTab === 'debug' && selectedId != null}
        <DebugTab accountId={selectedId} />
      {:else}
        <NoAccountTab />
      {/if}
    </div>
  </div>
</div>
