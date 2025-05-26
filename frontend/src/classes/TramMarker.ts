import { CircleMarker, CircleMarkerOptions } from "leaflet"
import { LeafletMap } from "@classes/LeafletMap"

export class TramMarker extends CircleMarker {
  private isOnMap = false
  private isSelected = false

  constructor(private leafletMap: LeafletMap) {
    super([0, 0], {
      radius: 5,
      fill: true,
      color: "red",
    })
    this.leafletMap = leafletMap
  }

  public setSelected(isSelected: boolean) {
    this.isSelected = isSelected
    this.setStyle({
      color: isSelected ? "orange" : "red"
    })
  }

  public updateCoordinates(lat: number, lon: number) {
    if (!this.isOnMap) {
      this.leafletMap.addTram(this)
      this.isOnMap = true
    }

    this.setLatLng([lat, lon])
  }

  public removeFromMap() {
    if (!this.isOnMap) return

    this.leafletMap.removeTram(this)
    this.isOnMap = false
  }
}
