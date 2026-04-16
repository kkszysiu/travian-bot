<script lang="ts">
  let { visible = false, message = '' }: { visible: boolean; message?: string } = $props()
</script>

{#if visible}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="flex flex-col items-center gap-3 rounded-lg bg-white px-8 py-6 shadow-xl">
      <div class="spinner"></div>
      {#if message}
        <p class="text-sm text-gray-600">{message}</p>
      {/if}
    </div>
  </div>
{/if}

<style>
  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
