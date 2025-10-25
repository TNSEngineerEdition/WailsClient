import { modifiedNodes as reactiveModifiedNodes } from "@composables/store"
import { GetBounds, GetCityRectangles } from "@wails/go/city/City"
import { city } from "@wails/go/models"
import L, {
  LatLngBounds,
  LatLngExpression,
  Map as LMap,
  tileLayer,
} from "leaflet"

//
// TYPES
//
type GraphNeighbor = {
  id: number
  distance: number
  azimuth: number
  max_speed: number
}

type GraphNode = {
  details: {
    id: number
    lat: number
    lon: number
    neighbors: Record<number, GraphNeighbor>

    // tram stop specific
    gtfs_stop_ids?: string[]
    name?: string
    node_type?: string
  }
}

type CityRectangles = {
  bounds: city.CityRectangle["bounds"]
  nodes_by_id: Record<number, GraphNode>
}

//
// CLASS
//
export class LeafletCustomizeMap {
  private tracksLayer = L.layerGroup()
  private selectedRectangle: L.Rectangle | null = null

  private selectedStart: number | null = null
  private selectedNodes: number[] = []

  //
  // CONSTRUCTOR
  //
  constructor(
    private map: LMap,
    private modifiedNodes: typeof reactiveModifiedNodes,
  ) {
    this.tracksLayer.addTo(this.map)
  }

  //
  // METHODS
  //
  static async init(
    mapHTMLElement: HTMLElement,
    modifiedNodes: typeof reactiveModifiedNodes,
  ) {
    const leafletCustomizeMap = new LeafletCustomizeMap(
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
      modifiedNodes,
    )

    const rectangles: CityRectangles[] = await GetCityRectangles()
    for (const { bounds, nodes_by_id: nodes } of rectangles) {
      leafletCustomizeMap.drawRectangle(bounds, nodes)
    }

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(leafletCustomizeMap.map)

    return leafletCustomizeMap
  }

  private drawRectangle(
    { minLat, minLon, maxLat, maxLon }: city.LatLonBounds,
    nodes: Record<number, GraphNode>,
  ) {
    const rectangle = L.rectangle(
      [
        [minLat, minLon],
        [maxLat, maxLon],
      ],
      {
        fillColor: "white",
        fillOpacity: 0.3,
        color: "black",
        weight: 1,
        opacity: 1,
      },
    )

    rectangle.on("click", () => {
      if (this.selectedRectangle) this.map.addLayer(this.selectedRectangle)

      this.map.removeLayer(rectangle)
      this.selectedRectangle = rectangle

      this.tracksLayer.clearLayers()
      this.drawTracks(nodes)
    })

    rectangle.addTo(this.map)
  }

  // Draws tracks on the selected rectangle
  private drawTracks(nodes: Record<number, GraphNode>) {
    this.selectedStart = null
    this.selectedNodes = []

    for (const [nodeIDstr, nodeObj] of Object.entries(nodes)) {
      const nodeID = Number(nodeIDstr)
      const node = nodeObj.details

      for (const neighborIDstr in node.neighbors) {
        if (!(neighborIDstr in nodes)) continue

        const neighborID = Number(neighborIDstr)
        const neighbor = nodes[neighborID].details

        const latlngs = [
          [node.lat, node.lon],
          [neighbor.lat, neighbor.lon],
        ] satisfies LatLngExpression[]

        const maxSpeed =
          this.modifiedNodes[nodeID]?.neighborMaxSpeed[neighborID] ??
          node.neighbors[neighborID].max_speed

        const polylineOptions = {
          weight: 4,
          color: this.getColorForSpeed(maxSpeed),
          fill: false,
          smoothFactor: 3,
        } satisfies L.PolylineOptions

        const polyline = L.polyline(latlngs, polylineOptions)

        polyline.on("click", () => {
          this.onPolylineClick(nodeID, nodes)
          polyline.setStyle({
            weight: 6,
            color: "red",
          })
        })

        polyline.addTo(this.tracksLayer)
      }
    }
  }

  // Travels through the graph to find the second selected node
  // and determines the path to highlight.
  // It stops at tram stops, switches and the end of the rectangle
  private findSelectionEndAndPath(
    selectedNodeID: number | string,
    nodes: Record<number, GraphNode>,
  ) {
    if (!this.selectedStart)
      throw new Error("First node of selection is not selected")

    selectedNodeID = Number(selectedNodeID)
    this.selectedNodes = []

    let node = nodes[Number(this.selectedStart)].details

    while (true) {
      this.selectedNodes.push(node.id)

      // 1 - the same node
      if (node.id === selectedNodeID) {
        console.log("cond 1")
        break
      }

      // 2 - switch
      if (
        Object.keys(node.neighbors).length !== 1 &&
        node.id !== this.selectedStart
      ) {
        console.log("cond 2")
        break
      }

      // 3- tram stop
      if (node.node_type === "stop" && node.id !== this.selectedStart) {
        console.log("cond 3", "details below")
        console.log(node)
        console.log(this.selectedStart)
        break
      }

      // 4 - out of rectangle bounds
      const neighborID = Object.keys(node.neighbors)[0]
      if (!(neighborID in nodes)) {
        console.log("cond 4")
        break
      }

      node = nodes[Number(neighborID)].details
    }
  }

  private highlightSelectedPath(nodes: Record<number, GraphNode>) {
    this.tracksLayer.clearLayers()

    const latlngs: LatLngExpression[] = this.selectedNodes.map(nodeID => [
      nodes[nodeID].details.lat,
      nodes[nodeID].details.lon,
    ])

    const polyline = L.polyline(latlngs, {
      weight: 6,
      color: "red",
    })

    polyline.bindPopup(() => {
      const div = L.DomUtil.create("div")
      div.innerHTML = `
    <label>Max speed (km/h): </label>
    <input id="maxSpeedInput" type="number" value="50" step="5" style="width:80px"/>
    <button id="saveSpeedBtn">Save</button>
  `
      return div
    })

    polyline.on("popupopen", event => {
      this.onPolylinePopupOpen(event, nodes)
    })

    polyline.addTo(this.tracksLayer)
  }

  private onPolylineClick(nodeID: number, nodes: Record<number, GraphNode>) {
    if (!this.selectedStart) {
      this.selectedStart = nodeID
      return
    }

    this.findSelectionEndAndPath(nodeID, nodes)

    if (this.selectedNodes.length > 0) {
      this.highlightSelectedPath(nodes)
    }
  }

  private onPolylinePopupOpen(
    event: L.PopupEvent,
    nodes: Record<number, GraphNode>,
  ) {
    const container = event.popup.getElement()
    if (!container) return

    const input = container.querySelector<HTMLInputElement>("#maxSpeedInput")
    const button = container.querySelector<HTMLButtonElement>("#saveSpeedBtn")

    button?.addEventListener("click", () => {
      if (!input) return

      const newMaxSpeedValue = parseFloat(input.value) / 3.6
      if (!isNaN(newMaxSpeedValue)) {
        this.saveNewMaxSpeed(newMaxSpeedValue)
      }

      this.tracksLayer.clearLayers()
      this.drawTracks(nodes)
    })
  }

  private saveNewMaxSpeed(newMaxSpeed: number) {
    for (let i = 0; i < this.selectedNodes.length - 1; i++) {
      const nodeID = Number(this.selectedNodes[i])
      const nextNodeID = Number(this.selectedNodes[i + 1])

      if (!this.modifiedNodes[nodeID]) {
        this.modifiedNodes[nodeID] = {
          neighborMaxSpeed: {},
        }
      }
      this.modifiedNodes[nodeID].neighborMaxSpeed[nextNodeID] = newMaxSpeed
    }
  }

  private getColorForSpeed(speedMS: number): string {
    const speed = speedMS * 3.59 // convert m/s to km/h
    if (speed <= 10) return "#9100FF"
    if (speed <= 20) return "#7D88FF"
    if (speed <= 30) return "#05B6FC"
    if (speed <= 40) return "#1772FC"
    if (speed <= 50) return "#3D00FF"
    if (speed <= 60) return "#00b52aff"
    return "#00e636ff"
  }
}
