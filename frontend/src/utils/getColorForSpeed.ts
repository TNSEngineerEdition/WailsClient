function getColorForSpeed(speedMS: number): string {
  const speed = speedMS * 3.59 // m/s → km/h
  if (speed <= 10) return "#9100FF"
  if (speed <= 20) return "#7D88FF"
  if (speed <= 30) return "#05B6FC"
  if (speed <= 40) return "#1772FC"
  if (speed <= 50) return "#3D00FF"
  if (speed <= 60) return "#00b52aff"
  return "#00e636ff"
}

export default getColorForSpeed
