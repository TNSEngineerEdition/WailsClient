<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import TramControlButtonComponent from "@components/simulation/sidebar/TramControlButtonComponent.vue"
import { Time } from "@classes/Time"
import { tram } from "@wails/go/models"
import { GetTramDetails, StopResumeTram } from "@wails/go/simulation/Simulation"
import { computed, ref, watch } from "vue"
import { TramMarker } from "@classes/TramMarker"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  tramId?: number
  tramMarker?: TramMarker
  currentTime: number
  followTram: boolean
}>()

const tramDetails = ref<tram.TramDetails>()

const headers = [
  { title: "Stop name", key: "stop", align: "center", sortable: false },
  { title: "Departure", key: "time", align: "center", sortable: false },
  { title: "Arrival", key: "arrival", align: "center", sortable: false },
  { title: "Departure", key: "departure", align: "center", sortable: false },
] as const

const stopsTableData = computed(
  () =>
    tramDetails.value?.stop_names.map((stop, index) => {
      const time = tramDetails.value?.stops[index]?.time ?? 0
      const tripIndex = tramDetails.value?.trip_index ?? 0

      return {
        stop,
        time,
        arrival:
          index <= tripIndex
            ? (tramDetails.value?.arrivals[index] ?? 0) - time
            : null,
        departure:
          index <= tripIndex - 1
            ? (tramDetails.value?.departures[index] ?? 0) - time
            : null,
        id: tramDetails.value?.stops[index]?.id,
      }
    }) ?? [],
)

const isTramRunning = computed(
  () =>
    tramDetails.value?.state !== tram.TramState.STOPPED &&
    tramDetails.value?.state !== tram.TramState.STOPPING,
)

const isTramDisabled = computed(() => {
  return (
    !props.tramId ||
    tramDetails.value?.state === tram.TramState.TRIP_FINISHED ||
    tramDetails.value?.state === tram.TramState.TRIP_NOT_STARTED
  )
})

const emit = defineEmits(["stopSelected", "centerTram", "followTram"])

function getRowProps(data: any) {
  if (data.index === tramDetails.value?.trip_index)
    return {
      style:
        "background-color: rgba(40, 150, 241, 0.2); transition: background-color 0.3s ease, font-weight 0.3s ease;",
    }
  else
    return {
      style: "transition: background-color 0.3s ease;",
    }
}

function getDelayTextColorClass(delay: number) {
  if (delay > 0) {
    return "text-red font-weight-bold"
  } else if (delay < 0) {
    return "text-info font-weight-bold"
  } else {
    return ""
  }
}

function onStopClick(_: MouseEvent, row: { item: any }) {
  emit("stopSelected", row.item.id)
}

function onCenterTramClick() {
  if (props.tramId) emit("centerTram")
}

async function stopResumeTram() {
  if (isTramDisabled.value) return

  const updated = await StopResumeTram(props.tramId!)
  tramDetails.value = updated

  if (props.tramMarker) {
    const isStopped =
      updated.state === tram.TramState.STOPPED ||
      updated.state === tram.TramState.STOPPING
    props.tramMarker.setStopped(isStopped)
  }
}

watch(
  () => props.tramId,
  async id => {
    if (id) {
      tramDetails.value = await GetTramDetails(id)
    } else {
      tramDetails.value = undefined
      model.value = false
    }
  },
  { immediate: true },
)

watch(
  () => props.currentTime,
  async () => {
    if (props.tramId) {
      tramDetails.value = await GetTramDetails(props.tramId)
    }
  },
  { immediate: true },
)
</script>

<template>
  <SidebarComponent
    v-model="model"
    position="left"
    :title="
      tramDetails
        ? `${tramDetails.route} ➡ ${tramDetails.trip_head_sign}`
        : 'Loading data...'
    "
    title-icon="mdi-tram"
  >
    <template #title-actions>
      <v-btn
        icon="mdi-crosshairs-gps"
        variant="text"
        density="compact"
        :disabled="!props.tramId"
        @click="onCenterTramClick"
      />
    </template>
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-identifier" class="mr-2"></v-icon>
        Tram ID
      </div>
      <div class="value">{{ props.tramId }}</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-speedometer" class="mr-2"></v-icon>
        Speed
      </div>
      <div class="value">{{ tramDetails?.speed }} km/h</div>
    </div>
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-radar" class="mr-2"></v-icon>
        Follow tram
      </div>

      <div class="value">
        <v-btn
          icon
          variant="text"
          size="x-small"
          class="mini-checkbox"
          :color="followTram ? 'primary' : undefined"
          @click="() => emit('followTram', !followTram)"
        >
          <v-icon size="16">
            {{
              followTram ? "mdi-checkbox-marked" : "mdi-checkbox-blank-outline"
            }}
          </v-icon>
        </v-btn>
      </div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-account-group" class="mr-2"></v-icon>
        Passenger count
      </div>
      <div class="value">{{ tramDetails?.passengers_count }}</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi mdi-wrench-cog" class="mr-2"></v-icon>
        Simulate failure
      </div>
      <TramControlButtonComponent
        :running="isTramRunning"
        :disabled="isTramDisabled"
        @click="stopResumeTram"
      ></TramControlButtonComponent>
    </div>
    <div class="section" style="margin-bottom: 0px">
      <div class="label">
        <v-icon icon="mdi-map-marker-path" class="mr-2"></v-icon>
        Stops
      </div>
    </div>
    <div class="scrollable">
      <v-data-table-virtual
        v-if="tramDetails?.stop_names.length"
        :headers="headers"
        :header-props="{
          style: 'font-weight: bold;',
        }"
        :items="stopsTableData"
        :row-props="getRowProps"
        class="stops-table"
        density="compact"
        hide-default-footer
        hover
        @click:row="onStopClick"
      >
        <template v-slot:item.time="{ item }">
          {{ new Time(item.time).toShortMinuteString() }}
        </template>

        <template v-slot:item.arrival="{ item }">
          <span
            v-if="item.arrival != null"
            :class="getDelayTextColorClass(item.arrival)"
          >
            {{ new Time(item.arrival, true).toShortSecondString() }}
          </span>
        </template>

        <template v-slot:item.departure="{ item }">
          <span
            v-if="item.departure != null"
            :class="getDelayTextColorClass(item.departure)"
          >
            {{ new Time(item.departure, true).toShortSecondString() }}
          </span>
        </template>
      </v-data-table-virtual>
    </div>
  </SidebarComponent>
</template>

<style scoped lang="scss">
.scrollable {
  overflow-y: auto;
  max-height: 40vh;
}

.scrollable::-webkit-scrollbar {
  width: 6px;
}

.scrollable::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.stops-table {
  width: 100%;
  background-color: transparent;
}
</style>
