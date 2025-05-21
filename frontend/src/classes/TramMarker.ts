import { CircleMarker, CircleMarkerOptions } from "leaflet"
import { LeafletMap } from "@classes/LeafletMap"

export class TramMarker extends CircleMarker {
  private isOnMap = false

  constructor(
    private leafletMap: LeafletMap,
    options: CircleMarkerOptions,
    private onClickHandler: () => void
  ) {
    super([0, 0], options)

    this.leafletMap = leafletMap

    this.on('click', () => {
      this.setStyle({color: "orange"})
      this.onClickHandler()
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
