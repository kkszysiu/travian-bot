<script lang="ts">
  import MainLayout from './lib/components/layout/MainLayout.svelte'
  import WaitingOverlay from './lib/components/layout/WaitingOverlay.svelte'
  import { waiting, type WaitingState } from './lib/stores/waiting'

  let waitingState = $state<WaitingState>({ visible: false, message: '' })
  waiting.subscribe(v => waitingState = v)
</script>

<MainLayout />
<WaitingOverlay visible={waitingState.visible} message={waitingState.message} />
