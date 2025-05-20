import { CircleMarker } from "leaflet"

export class StopMarker extends CircleMarker {
  private selected = false

  constructor(lat: number, lon: number, name?: string | null) {
    super([lat, lon], {
      radius: 5,
      fill: true,
      color: "darkblue",
      fillColor: "blue",
      weight: 2,
      opacity: 1,
      fillOpacity: 0.8,
    })
    this.bindTooltip(name ?? "Unknown stop", {
      permanent: false,
      direction: "top",
    })
  }

  public setSelected(selected: boolean) {
    this.selected = selected
    this.setStyle({
      color: selected ? "darkcyan" : "darkblue",
      fillColor: selected ? "cyan" : "blue",
    })
  }
}
