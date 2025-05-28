import { Marker, DivIcon } from "leaflet"
import { LeafletMap } from "@classes/LeafletMap"

export class TramMarker extends Marker {
  private isOnMap = false

  constructor(
    private leafletMap: LeafletMap,
    route: string,
  ) {
    super([0, 0], { icon: TramMarker.createIcon(route) })
  }

  private static createIcon(route: string): DivIcon {
    return new DivIcon({
      className: "",
      html: `
        <div class="tram-marker">
          <div class="tm-circle-arrow" style="transform: rotate(0);"></div>
          <div class="tm-circle"></div>
          <div class="tm-route-label">${route}</div>
        </div>
      `,
      iconSize: [24, 24],
      iconAnchor: [12, 12],
    })
  }

  private setAzimuth(azimuth: number) {
    const circleArrow =
      this.getElement()?.querySelector<HTMLElement>(".tm-circle-arrow")
    if (!circleArrow) {
      throw new Error("Tram marker arrow not found")
    }

    circleArrow.style.transform = `rotate(${azimuth + 135}deg)`
  }

  public setSelected(isSelected: boolean) {
    const element =
      this.getElement()?.querySelector<HTMLElement>(".tram-marker")

    if (!element) {
      throw new Error("Tram marker not found")
    }

    if (isSelected) {
      element?.classList.add("selected")
    } else {
      element?.classList.remove("selected")
    }
  }

  public updateCoordinates(lat: number, lon: number, azimuth: number) {
    if (!this.isOnMap) {
      this.leafletMap.addTram(this)
      this.isOnMap = true
    }

    this.setLatLng([lat, lon])
    this.setAzimuth(azimuth)
  }

  public removeFromMap() {
    if (!this.isOnMap) return

    this.leafletMap.removeTram(this)
    this.isOnMap = false
  }
}
