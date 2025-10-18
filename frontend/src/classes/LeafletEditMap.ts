import { modifiedNodes as reactiveModifiedNodes } from "@composables/store"
import { GetBounds, GetCityRectangles } from "@wails/go/city/City"
import L, {
  LatLngBounds,
  LatLngExpression,
  Map as LMap,
  tileLayer,
} from "leaflet"

type GraphNeighbor = {
  id: number
  distance: number
  azimuth: number
  max_speed: number
}

type GraphNode = {
  id: number
  lat: number
  lon: number
  neighbors: Record<number, GraphNeighbor>
  name?: string
}

export class LeafletEditMap {
  private tracksLayer = L.layerGroup()
  private selectedRectangle: L.Rectangle | null = null

  // path selection
  private selectionStart: number | null = null
  private selectionEnd: number | null = null
  private selectedNodes: number[] = []

  constructor(
    private map: LMap,
    private modifiedNodes: typeof reactiveModifiedNodes,
  ) {
    this.tracksLayer.addTo(this.map)
  }

  static async init(
    mapHTMLElement: HTMLElement,
    modifiedNodes: typeof reactiveModifiedNodes,
  ) {
    const leafletEditMap = new LeafletEditMap(
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

    const rectangles = await GetCityRectangles()

    for (const { bounds, nodes_by_id: nodes } of rectangles) {
      const { minLat, minLon, maxLat, maxLon } = bounds

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
        if (leafletEditMap.selectedRectangle) {
          leafletEditMap.map.addLayer(leafletEditMap.selectedRectangle)
        }

        leafletEditMap.map.removeLayer(rectangle)
        leafletEditMap.selectedRectangle = rectangle

        leafletEditMap.tracksLayer.clearLayers()
        leafletEditMap.drawTracks(nodes)
      })

      rectangle.addTo(leafletEditMap.map)
    }

    tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: `&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>`,
    }).addTo(leafletEditMap.map)

    return leafletEditMap
  }

  private drawTracks(nodes: Record<number, GraphNode>): void {
    this.selectionStart = null
    this.selectionEnd = null
    this.selectedNodes = []

    for (const [nodeIDstr, node] of Object.entries(nodes)) {
      const nodeID = Number(nodeIDstr)

      for (const neighborIDstr in node.neighbors) {
        if (!(neighborIDstr in nodes)) continue

        const neighborID = Number(neighborIDstr)
        const neighbor = nodes[neighborID]

        const latlngs = [
          [node.lat, node.lon],
          [neighbor.lat, neighbor.lon],
        ] as LatLngExpression[]

        const maxSpeed =
          this.modifiedNodes[nodeID]?.neighborMaxSpeed[neighborID] ??
          node.neighbors[neighborID].max_speed

        const polylineOptions = {
          weight: 4,
          color: this.getColorForSpeed(maxSpeed),
          fill: false,
        } satisfies L.PolylineOptions

        const polyline = L.polyline(latlngs, polylineOptions)

        polyline.on("click", () => {
          if (!this.selectionStart) {
            this.selectionStart = nodeID
          } else {
            this.selectionEnd = this.findSelectionEnd(nodeID, nodes)
            this.findPath(nodes)

            if (this.selectedNodes.length > 0) {
              this.highlightPath(nodes)
            }
          }

          polyline.setStyle({
            weight: 6,
            color: "red",
          })
        })

        polyline.addTo(this.tracksLayer)
      }
    }
  }

  private findSelectionEnd(
    selectedNodeID: number | string,
    nodes: Record<number, GraphNode>,
  ): number {
    if (!this.selectionStart)
      throw new Error("First node of selection is not selected")

    selectedNodeID = Number(selectedNodeID)
    let node = nodes[Number(this.selectionStart)]

    while (true) {
      if (node.id === selectedNodeID) {
        console.log("cond 1")
        break
      }
      if (
        Object.keys(node.neighbors).length != 1 &&
        node.id !== this.selectionStart
      ) {
        console.log("cond 2")
        break
      }
      if (node.name !== null && node.id !== this.selectionStart) {
        console.log("cond 3", node.name)
        break
      }

      const neighborID = Object.keys(node.neighbors)[0]
      if (!(neighborID in nodes)) {
        console.log("cond 4")
        break
      }

      node = nodes[Number(neighborID)]
    }

    return node.id
  }

  private findPath(nodes: Record<number, GraphNode>): void {
    if (!this.selectionStart || !this.selectionEnd)
      throw new Error("Nodes are not selected")

    this.selectedNodes = []
    let node = nodes[this.selectionStart]

    if (this.selectionStart === this.selectionEnd) {
      this.selectedNodes = [node.id]
      return
    }

    while (true) {
      this.selectedNodes.push(node.id)

      if (node.id === this.selectionEnd) break

      const nextNodeID = Object.keys(node.neighbors)[0]
      node = nodes[Number(nextNodeID)]
    }
  }

  private highlightPath(nodes: Record<number, GraphNode>): void {
    this.tracksLayer.clearLayers()

    const latlngs: LatLngExpression[] = []

    this.selectedNodes.forEach(nodeID => {
      const node = nodes[nodeID]
      latlngs.push([node.lat, node.lon])
    })

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
      const container = event.popup.getElement()
      if (!container) return

      const input = container.querySelector<HTMLInputElement>("#maxSpeedInput")
      const button = container.querySelector<HTMLButtonElement>("#saveSpeedBtn")

      button?.addEventListener("click", () => {
        if (!input) return
        const newSpeedValue = parseFloat(input.value) / 3.6

        if (!isNaN(newSpeedValue)) {
          for (let i = 0; i < this.selectedNodes.length - 1; i++) {
            const nodeID = Number(this.selectedNodes[i])
            const nextNodeID = Number(this.selectedNodes[i + 1])

            if (!this.modifiedNodes[nodeID]) {
              this.modifiedNodes[nodeID] = { neighborMaxSpeed: {} }
            }
            this.modifiedNodes[nodeID].neighborMaxSpeed[nextNodeID] =
              newSpeedValue
          }
        }

        this.tracksLayer.clearLayers()
        this.drawTracks(nodes)
      })
    })

    polyline.addTo(this.tracksLayer)
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
