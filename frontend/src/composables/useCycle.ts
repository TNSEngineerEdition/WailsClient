import { Ref } from "vue";

export default function useCycle<T>(values: T[], currentValue: Ref<T>) {
  if (!values.length) {
    throw new Error("Empty cycle is not allowed")
  }

  let index = 0

  function setNextValue() {
    index = (index + 1) % values.length
    currentValue.value = values[index]
  }

  function reset() {
    index = 0
  }

  return { setNextValue, reset }
}
