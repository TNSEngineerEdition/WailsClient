<script setup lang="ts">
import SidebarComponent from "@components/simulation/sidebar/SidebarComponent.vue"
import VehicleControlButtonComponent from "@components/simulation/sidebar/VehicleControlButtonComponent.vue"
import { Time } from "@classes/Time"
import { vehicle } from "@wails/go/models"
import {
  GetVehicleDetails,
  StopResumeVehicle,
} from "@wails/go/simulation/Simulation"
import { computed, ref, watch } from "vue"
import { VehicleMarker } from "@classes/VehicleMarker"

const model = defineModel<boolean>({ required: true })

const props = defineProps<{
  vehicleId?: number
  vehicleMarker?: VehicleMarker
  currentTime: number
  followVehicle: boolean
}>()

const vehicleDetails = ref<vehicle.VehicleDetails>()

const headers = [
  { title: "Stop name", key: "stop", align: "center", sortable: false },
  { title: "Departure", key: "time", align: "center", sortable: false },
  { title: "Arrival", key: "arrival", align: "center", sortable: false },
  { title: "Departure", key: "departure", align: "center", sortable: false },
] as const

const stopsTableData = computed(
  () =>
    vehicleDetails.value?.stop_names.map((stop, index) => {
      const time = vehicleDetails.value?.stops[index]?.time ?? 0
      const tripIndex = vehicleDetails.value?.trip_index ?? 0

      return {
        stop,
        time,
        arrival:
          index <= tripIndex
            ? (vehicleDetails.value?.arrivals[index] ?? 0) - time
            : null,
        departure:
          index <= tripIndex - 1
            ? (vehicleDetails.value?.departures[index] ?? 0) - time
            : null,
        id: vehicleDetails.value?.stops[index]?.id,
      }
    }) ?? [],
)

const isVehicleRunning = computed(
  () =>
    vehicleDetails.value?.state !== vehicle.VehicleState.STOPPED &&
    vehicleDetails.value?.state !== vehicle.VehicleState.STOPPING,
)

const isVehicleDisabled = computed(() => {
  return (
    !props.vehicleId ||
    vehicleDetails.value?.state === vehicle.VehicleState.TRIP_FINISHED ||
    vehicleDetails.value?.state === vehicle.VehicleState.TRIP_NOT_STARTED
  )
})

const emit = defineEmits(["stopSelected", "centerVehicle", "followVehicle"])

function getRowProps(data: any) {
  if (data.index === vehicleDetails.value?.trip_index)
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

function onCenterVehicleClick() {
  emit("centerVehicle")
}

async function stopResumeVehicle() {
  if (isVehicleDisabled.value) return

  const updated = await StopResumeVehicle(props.vehicleId!)
  vehicleDetails.value = updated

  if (props.vehicleMarker) {
    const isStopped =
      updated.state === vehicle.VehicleState.STOPPED ||
      updated.state === vehicle.VehicleState.STOPPING
    props.vehicleMarker.setStopped(isStopped)
  }
}

watch(
  () => props.vehicleId,
  async id => {
    if (id) {
      vehicleDetails.value = await GetVehicleDetails(id)
    } else {
      vehicleDetails.value = undefined
      model.value = false
    }
  },
  { immediate: true },
)

watch(
  () => props.currentTime,
  async () => {
    if (props.vehicleId) {
      vehicleDetails.value = await GetVehicleDetails(props.vehicleId)
    }
  },
  { immediate: true },
)

watch(
  () => isVehicleDisabled.value,
  disabled => {
    if (disabled && props.followVehicle) {
      emit("followVehicle", false)
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
      vehicleDetails
        ? `${vehicleDetails.route} ➡ ${vehicleDetails.trip_head_sign}`
        : 'Loading data...'
    "
    title-icon="mdi-tram"
  >
    <template #title-actions>
      <v-btn
        icon="mdi-crosshairs-gps"
        variant="text"
        density="compact"
        :disabled="!props.vehicleId || isVehicleDisabled"
        @click="onCenterVehicleClick"
      />
    </template>
    <div class="section">
      <div class="label">
        <v-icon icon="mdi-identifier" class="mr-2"></v-icon>
        Vehicle ID
      </div>
      <div class="value">{{ props.vehicleId }}</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-speedometer" class="mr-2"></v-icon>
        Speed
      </div>
      <div class="value">{{ vehicleDetails?.speed }} km/h</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-account-group" class="mr-2"></v-icon>
        Passenger count
      </div>
      <div class="value">{{ vehicleDetails?.passengers_count }}</div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi-radar" class="mr-2" />
        Follow vehicle
      </div>

      <div class="value">
        <v-switch
          :disabled="isVehicleDisabled"
          color="info"
          density="compact"
          hide-details
          @update:model-value="value => emit('followVehicle', value)"
        />
      </div>
    </div>

    <div class="section">
      <div class="label">
        <v-icon icon="mdi mdi-wrench-cog" class="mr-2"></v-icon>
        Simulate failure
      </div>
      <VehicleControlButtonComponent
        :running="isVehicleRunning"
        :disabled="isVehicleDisabled"
        @click="stopResumeVehicle"
      ></VehicleControlButtonComponent>
    </div>

    <div class="section" style="margin-bottom: 0px">
      <div class="label">
        <v-icon icon="mdi-map-marker-path" class="mr-2"></v-icon>
        Stops
      </div>
    </div>
    <div class="scrollable">
      <v-data-table-virtual
        v-if="vehicleDetails?.stop_names.length"
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

<style lang="scss">
.v-switch .v-selection-control {
  min-height: unset !important;
}
</style>
