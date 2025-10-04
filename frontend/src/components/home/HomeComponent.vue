<script setup lang="ts">
import CityDialogComponent from "@components/home/CityDialogComponent.vue"
import MenuBackgroundIMG from "@assets/images/menu-background.jpg"
import { onMounted } from "vue"
// import { FetchCities } from "@wails/go/city/City"
import { city } from "@wails/go/models"

const cities: Record<string, any> = {
  krakow: {
    cityName: "Kraków",
    country: "Poland",
    image: "https://mpk.krakow.pl/media/photos/6941/l.jpg",
  },
}

const backgroundImageURL = `url(${MenuBackgroundIMG})`

onMounted(() => {
  // FetchCities()
})
</script>

<template>
  <div class="background">
    <v-container class="my-sm-10">
      <v-row>
        <v-col v-for="city in cities" cols="12" sm="6" xl="4">
          <CityDialogComponent
            :city-name="`${city.cityName}, ${city.country}`"
            :image="city.image"
          ></CityDialogComponent>
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
