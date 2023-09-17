<script lang="ts">
  import { onMount } from "svelte";
  import type {DashboardSettings} from "../../lib/settings"

  let container: HTMLDivElement;
  onMount(() => {
    container.addEventListener('click', e => {
      e.stopImmediatePropagation()

    })
  })
  export let show: boolean, settings: DashboardSettings;
</script>

<div class="background" class:hidden={!show} on:click={() => { show = false} }>
  <div class="container" bind:this={container}>
    <h2 class="title">Settings</h2>
    <div class="disable404 setting">
      <div class="setting-label">
        Disable 404
      </div>
      <input type="checkbox" name="disable404" id="checkbox" bind:value={settings.disable404}>
    </div>
  </div>
</div>

<style scoped>
  .background {
    background: rgba(0, 0, 0, 0.6);
    height: 100vh;
    width: 100%;
    display: grid;
    place-items: center;
    position: absolute;
    top: 0;
    cursor: pointer;
    z-index: 10;
  }
  .container {
    background: var(--background);
    border-radius: 6px;
    width: 30vw;
    height: 30vh;
    border: 1px solid #2E2E2E;
    color: var(--faded-text);
    z-index: 20; 
    position: absolute;
    cursor: default !important;
    pointer-events: bounding-box;
    padding: 30px 50px 50px;
  }
  .title {
    font-size: 1.8em;
    font-weight: 600;
    text-align: left
  }
  .hidden {
    display: none;
  }
  .setting {
    display: flex;
  }
  .setting-label {
    margin-right: 10px;
  }
  
  #checkbox {
    height: 15px;
    width: 15px;
  }

</style>