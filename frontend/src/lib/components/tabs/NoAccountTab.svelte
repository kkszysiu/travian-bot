<script lang="ts">
  import { t } from '../../i18n'

  let tr = $state<(key: string, params?: Record<string, string | number>) => string>((k) => k)
  t.subscribe(v => tr = v)
</script>

<div class="flex items-center justify-center h-full text-gray-400">
  <p>{tr('noAccount.message')}</p>
</div>
