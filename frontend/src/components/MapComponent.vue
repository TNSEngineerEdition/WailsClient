<script lang="ts" setup>
import { onMounted, ref, useTemplateRef, watch } from "vue"
import { GetTimeBounds } from "@wails/go/city/City"
import {
  GetTramIDs,
  AdvanceTrams,
  FetchData,
  GetTramDetails,
} from "@wails/go/simulation/Simulation"
import { LeafletMap } from "@classes/LeafletMap"
import { TramMarker } from "@classes/TramMarker"
import useTimeUtils from "@composables/useTimeUtils"
import TramSidebarComponent from "@components/sidebar/TramSidebarComponent.vue"
import StopSidebarComponent from "@components/sidebar/StopSidebarComponent.vue"
import { simulation } from "@wails/go/models"

const mapHTMLElement = useTemplateRef("map")

const time = defineModel<number>("time", { required: true })
const loading = defineModel<boolean>("loading", { required: true })

const props = defineProps<{
  speed: number
  isRunning: boolean
  resetCounter: number
}>()

const endTime = ref(0)
const leafletMap = ref<LeafletMap>()
const tramMarkerByID = ref<Record<number, TramMarker>>({})

const tramSidebar = ref(false)
const stopSidebar = ref(false)

const selectedTramID = ref<number | null>(null)
const selectedTramDetails = ref<simulation.TramDetails | null>(null)

const timeUtils = useTimeUtils()

async function handleGetTramDetails(id: number) {
  if (selectedTramID.value !== null && id !== selectedTramID.value)
    tramMarkerByID.value[selectedTramID.value].removeHighlightColor()

  selectedTramID.value = id
  selectedTramDetails.value = await GetTramDetails(id)
  tramSidebar.value = true

  //console.log(leafletMap.value?.getEntityCount())
}

async function reset() {
  tramSidebar.value = false
  stopSidebar.value = false
  loading.value = true

  for (const tramMarker of Object.values(tramMarkerByID.value)) {
    tramMarker.removeFromMap()
  }

  tramMarkerByID.value = await GetTramIDs().then(tramIDs =>
    leafletMap.value!.getTramMarkers(tramIDs, handleGetTramDetails),
  )

  await GetTimeBounds().then(timeBounds => {
    time.value = timeBounds.startTime
    endTime.value = timeBounds.endTime
  })

  loading.value = false
}

watch(() => props.resetCounter, reset)

watch(tramSidebar, (newVal) => {
  if (!newVal) {
    tramMarkerByID.value[selectedTramID.value!].removeHighlightColor()
    selectedTramID.value = null
    selectedTramDetails.value = null
  }
})

onMounted(async () => {
  if (mapHTMLElement.value === null) {
    throw new Error("Map element not found")
  }

  await FetchData("http://localhost:8000/cities/krakow")
  leafletMap.value = await LeafletMap.init(mapHTMLElement.value)

  await reset()

  while (time.value <= endTime.value || leafletMap.value!.getEntityCount() > 0) {
    while (!props.isRunning) {
      await timeUtils.sleep(1)
    }

    await AdvanceTrams(time.value).then(tramPositionChanges => {
      for (const tram of tramPositionChanges) {
        if (tram.lat == 0 && tram.lon == 0) {
          tramMarkerByID.value[tram.id].removeFromMap()
        } else {
          tramMarkerByID.value[tram.id].updateCoordinates(tram.lat, tram.lon)
        }
      }
    })

    if (selectedTramID.value !== null)
      handleGetTramDetails(selectedTramID.value)

    time.value += 1

    await timeUtils.sleep(1000 / props.speed)
  }
})
</script>

<template>
  <v-overlay
    v-model="loading"
    opacity="0"
    class="d-flex justify-center align-center"
    persistent
  >
    <v-progress-circular indeterminate size="128"></v-progress-circular>
  </v-overlay>

  <div id="map" ref="map"></div>

  <TramSidebarComponent
    v-model="tramSidebar"
    v-if="selectedTramID !== null && selectedTramDetails !== null"
    :tram-i-d="selectedTramID"
    :tram-details="selectedTramDetails"
  />

  <StopSidebarComponent v-model="stopSidebar"></StopSidebarComponent>
</template>

<style scoped lang="scss">
#map {
  width: 100%;
  height: calc(100vh - 64px);
}
</style>
