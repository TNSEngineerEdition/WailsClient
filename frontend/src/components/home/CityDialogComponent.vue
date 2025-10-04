<script setup lang="ts">
const props = defineProps<{
  cityName: string
  image: string
}>()
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
            :src="props.image"
            class="align-end"
            gradient="to bottom, rgba(0,0,0,.1), rgba(0,0,0,.5)"
            height="325px"
            cover
          >
            <v-card-title
              class="text-black text-center"
              style="background-color: white"
            >
              {{ props.cityName }}
            </v-card-title>
          </v-img>
        </v-card>
      </v-hover>
    </template>

    <v-card :title="props.cityName">
      <template v-slot:title>
        <span class="text-center">{{ props.cityName }}</span>
      </template>

      <v-card-text>
        <v-form>
          <v-switch
            class="d-flex flex-column align-center"
            label="Customize speed limits"
            disabled
          ></v-switch>

          <v-select></v-select>

          <v-file-input
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
