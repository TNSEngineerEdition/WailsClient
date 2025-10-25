<script lang="ts" setup>
import { onMounted, ref, toRaw, useTemplateRef } from "vue"
import { LeafletCustomizeMap } from "@classes/LeafletCustomizeMap"
import CustomizeSidebarComponent from "./CustomizeSidebarComponent.vue"
import { modifiedNodes } from "@composables/store"
import { UpdateTramTrackGraph } from "@wails/go/city/City"
import router from "@plugins/router"
import { InitializeSimulation } from "@wails/go/simulation/Simulation"
import CustomizeSpeedDialogComponent from "./CustomizeSpeedDialogComponent.vue"

const loading = defineModel<boolean>("loading", { required: true })

const mapHTMLElement = useTemplateRef("customize-map")
const leafletCustomizeMap = ref<LeafletCustomizeMap>()
const speedDialog = ref(false)
const onCancelCallback = ref<() => void>()
const onSpeedSaveCallback = ref<(newMaxSpeed: number) => void>()

async function saveChanges() {
  const rawModifiedNodes = toRaw(modifiedNodes)
  console.log("sending to Go:", rawModifiedNodes)

  await UpdateTramTrackGraph(rawModifiedNodes)

  const simulationErrorMessage = await InitializeSimulation(0)
  if (simulationErrorMessage) {
    loading.value = false
    return
  }

  router.push("/simulation")
}

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  leafletCustomizeMap.value = await LeafletCustomizeMap.init(
    mapHTMLElement.value,
    modifiedNodes,
    ({ onCancel, onSpeedSave }) => {
      onCancelCallback.value = onCancel
      onSpeedSaveCallback.value = onSpeedSave
      speedDialog.value = true
    },
  )

  loading.value = false
})
</script>

<template>
  <div class="container">
    <CustomizeSidebarComponent :save-changes="saveChanges" />
    <CustomizeSpeedDialogComponent
      v-model:speed-dialog="speedDialog"
      v-model:on-cancel="onCancelCallback"
      v-model:on-speed-save="onSpeedSaveCallback"
    />
    <v-overlay
      v-model="loading"
      opacity="0.2"
      class="d-flex justify-center align-center"
      persistent
      max-width="500px"
    >
      <v-progress-circular indeterminate size="128"></v-progress-circular>
    </v-overlay>
    <div id="customize-map" ref="customize-map"></div>
  </div>
</template>

<style lang="scss">
.container {
  height: 100vh;
  width: 100%;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: row;
}

#customize-map {
  width: calc(100vw - 350px);
  height: 100vh;
}
</style>
