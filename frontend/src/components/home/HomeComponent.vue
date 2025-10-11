<script setup lang="ts">
import CityDialogComponent from "@components/home/CityDialogComponent.vue"
import MenuBackgroundIMG from "@assets/images/menu-background.jpg"
import { onMounted, ref } from "vue"
import { GetCities } from "@wails/go/api/APIClient"
import { api } from "@wails/go/models"

const cities = ref<api.CityInfo[]>([])

const backgroundImageURL = `url(${MenuBackgroundIMG})`

onMounted(async () => {
  cities.value = await GetCities()
})
</script>

<template>
  <div class="background">
    <v-container class="my-sm-10">
      <v-row>
        <v-col v-for="city in cities" cols="12" sm="6" xl="4">
          <CityDialogComponent :city="city"></CityDialogComponent>
        </v-col>
      </v-row>
    </v-container>
  </div>
</template>

<style>
.background {
  position: absolute;
  width: 100vw;
  min-height: 100vh;
  background: v-bind(backgroundImageURL) no-repeat center center fixed;
  background-size: cover;
}
</style>
