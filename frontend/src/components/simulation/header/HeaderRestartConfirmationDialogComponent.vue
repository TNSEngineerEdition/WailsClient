<script setup lang="ts">
import { ref, watch } from "vue"
import HeaderIconButtonComponent from "@components/simulation/header/HeaderIconButtonComponent.vue"

const props = defineProps<{
  disabled: boolean
}>()

const emit = defineEmits<{
  click: []
  reset: []
}>()

const dialog = ref(false)

function confirm() {
  emit("reset")
  dialog.value = false
}

watch(dialog, value => {
  if (value) {
    emit("click")
  }
})
</script>

<template>
  <v-dialog v-model="dialog" width="unset">
    <template v-slot:activator="{ props: dialogProps }">
      <HeaderIconButtonComponent
        :parent-props="dialogProps"
        :disabled="props.disabled"
        description="Restart"
        icon="mdi-replay"
      ></HeaderIconButtonComponent>
    </template>

    <v-card>
      <v-card-title class="text-center pb-0 pt-4">
        Restart simulation
      </v-card-title>

      <v-card-text class="text-h7 mx-5 pb-3">
        Are you sure you want to restart the simulation?
      </v-card-text>

      <v-card-actions class="d-flex justify-space-around">
        <v-btn
          class="mx-1"
          text="No"
          color="red"
          width="85"
          @click="dialog = false"
        ></v-btn>

        <v-btn
          class="mx-1"
          text="Yes"
          color="success"
          width="85"
          @click="confirm"
        ></v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
