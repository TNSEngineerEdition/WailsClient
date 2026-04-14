import { GetBounds, GetStops } from "@wails/go/city/City"
import { LatLngBounds, Map as LMap, tileLayer } from "leaflet"
import { VehicleMarker } from "@classes/VehicleMarker"
import { StopMarker } from "@classes/StopMarker"
import { city, simulation, api } from "@wails/go/models"
import { RouteHighlighter } from "./RouteHighlighter"

export class LeafletMap {
  private entityCount = 0
  public selectedStop?: StopMarker
  public selectedVehicle?: VehicleMarker
  private followVehicle = false
  public selectedRouteName?: string
  public highlightedRouteVehicles?: VehicleMarker[]
  private routeHighlighter: RouteHighlighter
  private stopMarkersById: Record<number, StopMarker> = {}

  constructor(private map: LMap) {
    this.routeHighlighter = new RouteHighlighter(map)
  }

  static async init(
    mapHTMLElement: HTMLElement,
    handleStopSelection: (stop: api.ResponseGraphStop) => void,
  ) {
    const leafletMap = new LeafletMap(
      await GetBounds()
        .then(
          bounds =>
            new LatLngBounds(
              [bounds.minLat, bounds.minLon],
              [bounds.maxLat, bounds.maxLon],
            ),
        )
        .then(
          latLngBounds =>
            new LMap(mapHTMLElement, {
              maxBounds: latLngBounds.pad(1),
              center: latLngBounds.getCenter(),
              zoom: 13,
            }),
        ),
    )

    await leafletMap.makeStops(handleStopSelection)

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(leafletMap.map)

    return leafletMap
  }

  private async makeStops(
    handleStopSelection: (stop: api.ResponseGraphStop) => void,
  ) {
    for (const stop of await GetStops()) {
      const marker = new StopMarker(stop)
      marker.addTo(this.map)
      marker.on("click", () => {
        if (this.selectedStop) {
          this.selectedStop.setSelected(false)
        }
        this.selectedStop = marker
        marker.setSelected(true)
        handleStopSelection(stop)
      })
      this.stopMarkersById[stop.id] = marker
    }
  }

  public highlightVehiclesForRoute(vehicles: VehicleMarker[]) {
    this.highlightedRouteVehicles?.forEach(m => m.setHighlighted(false))
    this.highlightedRouteVehicles = vehicles
    this.highlightedRouteVehicles.forEach(m => m.setHighlighted(true))
  }

  public async highlightRoute(route: city.RouteInfo) {
    this.selectedRouteName = route.name
    await this.routeHighlighter.highlight(route)
  }

  public deselectRoute() {
    this.selectedRouteName = undefined
    this.highlightedRouteVehicles?.forEach(m => m.setHighlighted(false))
    this.highlightedRouteVehicles = undefined
    this.routeHighlighter.clear()
  }

  public deselectStop() {
    if (this.selectedStop) {
      this.selectedStop.setSelected(false)
      this.selectedStop = undefined
    }
  }

  public getVehicleMarkers(
    vehicles: simulation.VehicleIdentifier[],
    onClickHandler: (id: number) => void,
  ) {
    const result: Record<number, VehicleMarker> = {}

    for (const vehicle of vehicles) {
      const marker = new VehicleMarker(this, vehicle.route)
      marker.on("click", () => {
        if (this.selectedVehicle) this.selectedVehicle.setSelected(false)
        this.selectedVehicle = marker
        marker.setSelected(true)
        onClickHandler(vehicle.id)
      })
      result[vehicle.id] = marker
    }

    return result
  }

  public addVehicle(vehicleMarker: VehicleMarker) {
    this.entityCount++
    vehicleMarker.addTo(this.map)
  }

  public removeVehicle(vehicleMarker: VehicleMarker) {
    this.entityCount--
    vehicleMarker.removeFrom(this.map)
  }

  public deselectVehicle() {
    if (this.selectedVehicle) {
      this.selectedVehicle.setSelected(false)
      this.selectedVehicle = undefined
    }
  }

  public getEntityCount() {
    return this.entityCount
  }

  public centerOn(lat: number, lon: number) {
    const z = 17
    this.map.flyTo([lat, lon], z, { animate: true, duration: 0.6 })
  }

  public setFollowVehicle(enabled: boolean) {
    this.followVehicle = enabled
  }

  public followTick() {
    if (!this.followVehicle || !this.selectedVehicle) return
    this.map.panTo(this.selectedVehicle.getLatLng(), {
      animate: true,
      duration: 0.25,
    })
  }

  public getStopMarker(stopId: number): StopMarker {
    return this.stopMarkersById[stopId]
  }

  public getMap(): LMap {
    return this.map
  }
}
