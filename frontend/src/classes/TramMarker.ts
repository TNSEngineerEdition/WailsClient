import { Marker, DivIcon, LatLngExpression } from "leaflet"
import { LeafletMap } from "@classes/LeafletMap"

export class TramMarker extends Marker {
  private isOnMap = false
  private isSelected = false
  private route: string
  private azimuth: number

  constructor(
    private leafletMap: LeafletMap,
    route: string,
  ) {
    const position: LatLngExpression = [0, 0]
    const icon = TramMarker.createIcon(route, 0, false)

    super(position, { icon })
    this.route = route
    this.azimuth = 0
  }

  private static createIcon(
    route: string,
    rotateDeg: number,
    selected: boolean,
  ): DivIcon {
    return new DivIcon({
      className: "",
      html: `
        <div class="tram-marker ${selected ? "selected" : ""}">
          <div class="tm-circle-arrow" style="transform: rotate(${135 + rotateDeg}deg);"></div>
          <div class="tm-circle"></div>
          <div class="tm-route-label">${route}</div>
        </div>
      `,
      iconSize: [24, 24],
      iconAnchor: [12, 12],
    })
  }

  public setSelected(isSelected: boolean) {
    this.isSelected = isSelected
    this.setIcon(TramMarker.createIcon(this.route, this.azimuth, isSelected))
  }

  public updateCoordinates(lat: number, lon: number, azimuth: number) {
    if (!this.isOnMap) {
      this.leafletMap.addTram(this)
      this.isOnMap = true
    }

    this.setLatLng([lat, lon])
    this.azimuth = azimuth
    this.setIcon(TramMarker.createIcon(this.route, azimuth, this.isSelected))
  }

  public removeFromMap() {
    if (!this.isOnMap) return

    this.leafletMap.removeTram(this)
    this.isOnMap = false
  }
}
