import { api } from "@wails/go/models"
import { CircleMarker } from "leaflet"

export class StopMarker extends CircleMarker {
  private selected = false
  private stop?: api.ResponseGraphTramStop

  constructor(stop: api.ResponseGraphTramStop) {
    super([stop.lat, stop.lon], {
      radius: 5,
      fill: true,
      color: "darkblue",
      fillColor: "blue",
      weight: 2,
      opacity: 1,
      fillOpacity: 0.8,
    })
    this.bindTooltip(stop.name ?? "Unknown stop", {
      permanent: false,
      direction: "top",
    })
    this.stop = stop
  }

  public setSelected(selected: boolean) {
    this.selected = selected
    this.setStyle({
      color: selected ? "darkcyan" : "darkblue",
      fillColor: selected ? "cyan" : "blue",
    })
  }

  public getStop(): api.ResponseGraphTramStop | undefined {
    return this.stop
  }
}
