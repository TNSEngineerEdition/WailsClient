<script setup lang="ts">
import router from "@plugins/router"
import { api, simulation } from "@wails/go/models"
import {
  InitializeCity,
  InitializeSimulation,
} from "@wails/go/simulation/Simulation"
import { computed, ref } from "vue"
import ErrorDialogComponent from "@components/home/ErrorDialogComponent.vue"

const props = defineProps<{
  city: api.CityInfo
}>()

const date = ref<string>()
const weekday = ref<api.Weekday>()
const customSchedule = ref<File>()

const showError = ref(false)
const error = ref<string>()

const cityName = computed(
  () =>
    `${props.city.cityConfiguration.city}, ${props.city.cityConfiguration.country}`,
)

const loading = ref(false)

const disableDate = computed(() => !!(weekday.value || customSchedule.value))
const disableWeekday = computed(() => !!date.value)
const disableCustomSchedule = computed(() => !!date.value)

const weekdayItems = computed(() => {
  return Object.values(api.Weekday).map(item => ({
    title: item[0].toUpperCase() + item.slice(1),
    value: item,
  }))
})

async function handleButtonClick(isCustomizeMap: boolean) {
  loading.value = true

  const parameters = new simulation.SimulationParameters({
    cityID: props.city.cityID,
    date: date.value,
    weekday: weekday.value?.toLowerCase(),
    customSchedule: Array.from((await customSchedule.value?.bytes()) ?? []),
  })

  const dataErrorMessage = await InitializeCity(parameters)
  if (dataErrorMessage) {
    loading.value = false
    showError.value = true
    error.value = dataErrorMessage
    return
  }

  if (isCustomizeMap) {
    router.push("/customize-map")
    return
  }

  const simulationErrorMessage = await InitializeSimulation(0)
  if (simulationErrorMessage) {
    loading.value = false
    showError.value = true
    error.value = simulationErrorMessage
    return
  }

  router.push("/simulation")
}
</script>

<template>
  <ErrorDialogComponent
    v-model="showError"
    title="Error initializing simulation"
    :error="error"
  ></ErrorDialogComponent>

  <v-dialog max-width="500">
    <template v-slot:activator="{ props: dialogProps }">
      <v-hover v-slot="{ isHovering, props: hoverProps }">
        <v-card
          v-bind="{ ...hoverProps, ...dialogProps }"
          :elevation="isHovering ? 12 : 3"
          style="user-select: none"
        >
          <v-img
            :src="props.city.cityConfiguration.image"
            class="align-end"
            gradient="to bottom, rgba(0,0,0,.1), rgba(0,0,0,.5)"
            height="325px"
            cover
          >
            <v-card-title
              class="text-black text-center"
              style="background-color: white"
            >
              {{ cityName }}
            </v-card-title>
          </v-img>
        </v-card>
      </v-hover>
    </template>

    <v-card>
      <template v-slot:title>
        <span class="text-center">{{ cityName }}</span>
      </template>

      <v-card-text>
        <v-form>
          <v-select
            v-model="date"
            :items="props.city.availableDates"
            :disabled="disableDate || loading"
            prepend-icon="mdi-calendar"
            label="Schedule date"
            clearable
          ></v-select>

          <v-select
            v-model="weekday"
            :items="weekdayItems"
            :disabled="disableWeekday || loading"
            prepend-icon="mdi-view-week"
            label="Weekday"
            clearable
          ></v-select>

          <v-file-input
            v-model="customSchedule"
            :disabled="disableCustomSchedule || loading"
            accept="application/zip"
            prepend-icon="mdi-invoice-text-clock"
            label="Custom GTFS Schedule file"
          ></v-file-input>

          <v-file-input
            accept="text/csv"
            prepend-icon="mdi-transit-transfer"
            label="Passenger model"
            disabled
          ></v-file-input>
        </v-form>
      </v-card-text>

      <template v-slot:actions>
        <div class="btn-container">
          <v-progress-linear
            v-if="loading"
            color="blue"
            height="7"
            indeterminate
          />
          <v-btn
            v-if="!loading"
            text="Customize speeds"
            variant="elevated"
            color="white"
            style="width: 50%"
            @click="
              () => {
                handleButtonClick(true)
              }
            "
          >
          </v-btn>
          <v-btn
            v-if="!loading"
            text="Start"
            variant="elevated"
            color="blue"
            style="width: 50%"
            @click="
              () => {
                handleButtonClick(false)
              }
            "
          ></v-btn>
        </div>
      </template>
    </v-card>
  </v-dialog>
</template>

<style lang="scss" scoped>
.btn-container {
  width: 100%;
  padding: 0 20px 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
</style>
