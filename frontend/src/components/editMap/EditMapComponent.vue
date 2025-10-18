<script lang="ts" setup>
import { onMounted, ref, toRaw, useTemplateRef } from "vue"
import { FetchDataWithoutSimulationStart } from "@wails/go/simulation/Simulation"
import { LeafletEditMap } from "@classes/LeafletEditMap"
import EditSidebarComponent from "./EditSidebarComponent.vue"
import { modifiedNodes } from "@composables/store"
import { UpdateTramTrackGraph } from "@wails/go/city/City"

const isEditMap = defineModel<boolean>("is-edit-map", { required: true })
const loading = defineModel<boolean>("loading", { required: true })

const mapHTMLElement = useTemplateRef("edit-map")

const leafletEditMap = ref<LeafletEditMap>()

async function saveChanges() {
  const rawModifiedNodes = toRaw(modifiedNodes)
  console.log("sending to Go:", rawModifiedNodes)

  await UpdateTramTrackGraph(rawModifiedNodes)
  isEditMap.value = false
}

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  await FetchDataWithoutSimulationStart("krakow")
  leafletEditMap.value = await LeafletEditMap.init(
    mapHTMLElement.value,
    modifiedNodes,
  )

  loading.value = false
})
</script>

<template>
  <div class="container">
    <EditSidebarComponent :save-changes="saveChanges" />
    <v-overlay
      v-model="loading"
      opacity="0.2"
      class="d-flex justify-center align-center"
      persistent
      max-width="500px"
    >
      <v-progress-circular indeterminate size="128"></v-progress-circular>
    </v-overlay>
    <div id="edit-map" ref="edit-map"></div>
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

#edit-map {
  width: calc(100vw - 350px);
  height: 100vh;
}
</style>
