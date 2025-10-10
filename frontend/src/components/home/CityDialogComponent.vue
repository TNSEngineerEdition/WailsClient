<script setup lang="ts">
import { api } from "@wails/go/models"
import { computed, ref } from "vue"

const props = defineProps<{
  city: api.CityInfo
}>()

const date = ref<string>()
const weekday = ref<string>()
const customSchedule = ref<File>()

const cityName = computed(
  () =>
    `${props.city.cityConfiguration.city}, ${props.city.cityConfiguration.country}`,
)

const disableDate = computed(() => !!(weekday.value || customSchedule.value))
const disableWeekday = computed(() => !!date.value)
const disableCustomSchedule = computed(() => !!date.value)
</script>

<template>
  <v-dialog max-width="400">
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
          <v-switch
            class="d-flex flex-column align-center"
            label="Customize speed limits"
            disabled
          ></v-switch>

          <v-select
            v-model="date"
            :items="props.city.availableDates"
            :disabled="disableDate"
            prepend-icon="mdi-calendar"
            label="Schedule date"
            clearable
          ></v-select>

          <v-select
            v-model="weekday"
            :items="[
              'Monday',
              'Tuesday',
              'Wednesday',
              'Thursday',
              'Friday',
              'Saturday',
              'Sunday',
            ]"
            :disabled="disableWeekday"
            prepend-icon="mdi-view-week"
            label="Weekday"
            clearable
          ></v-select>

          <v-file-input
            v-model="customSchedule"
            :disabled="disableCustomSchedule"
            accept="application/zip"
            prepend-icon="mdi-invoice-text-clock"
            label="Custom GTFS Schedule file"
          ></v-file-input>

          <v-file-input
            accept="text/csv"
            prepend-icon="mdi-transit-transfer"
            label="Passenger model"
          ></v-file-input>
        </v-form>
      </v-card-text>

      <template v-slot:actions>
        <v-btn text="Start" to="/simulation" block></v-btn>
      </template>
    </v-card>
  </v-dialog>
</template>
