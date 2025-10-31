import getColorForSpeed from "@utils/getColorForSpeed"
import { CityRectangles, GraphNode, ModifiedNodes } from "@utils/types"
import { GetBounds, GetCityRectangles } from "@wails/go/city/City"
import { city } from "@wails/go/models"
import L, {
  LatLngBounds,
  LatLngExpression,
  Map as LMap,
  tileLayer,
} from "leaflet"

export class LeafletCustomizeMap {
  private tracksLayer = L.layerGroup()
  private selectedRectangle: L.Rectangle | null = null
  private selectedStart: { edgeStart: number; edgeEnd: number } | null = null
  private selectedNodes: number[] = []

  constructor(
    private map: LMap,
    private modifiedNodes: ModifiedNodes,
    private onSpeedDialogOpen: (context: {
      onCancel: () => void
      onSpeedSave: (newSpeed: number) => void
    }) => void,
  ) {
    this.tracksLayer.addTo(this.map)
  }

  static async init(
    mapHTMLElement: HTMLElement,
    modifiedNodes: ModifiedNodes,
    onSpeedDialogOpen: (context: {
      onCancel: () => void
      onSpeedSave: (newSpeed: number) => void
    }) => void,
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
      onSpeedDialogOpen,
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

        const polyline = L.polyline(latlngs, {
          weight: 4,
          color: getColorForSpeed(maxSpeed),
          fill: false,
          smoothFactor: 3,
        })

        polyline.on("click", () => {
          this.onPolylineClick(nodeID, neighborID, nodes)
          polyline.setStyle({ weight: 6, color: "red" })
        })

        polyline.addTo(this.tracksLayer)
      }
    }
  }

  // Travels through the graph to find the second selected node
  // and determines the path to highlight.
  // It stops at tram stops, switches and the end of the rectangle
  private findSelectionEndAndPath(
    selectedEdgeEnd: number,
    nodes: Record<number, GraphNode>,
  ) {
    if (!this.selectedStart)
      throw new Error("First edge of selection is not selected")

    this.selectedNodes = [this.selectedStart.edgeStart]
    let node = nodes[Number(this.selectedStart.edgeEnd)].details

    while (true) {
      this.selectedNodes.push(node.id)
      const nodeNeighbors = Object.keys(node.neighbors).map(Number)

      // 1 - reached the end of the selected edge
      if (node.id === selectedEdgeEnd) break

      // 2 - switch or crossing (>1 neighbors ahead)
      if (
        nodeNeighbors.length !== 1 &&
        node.id !== this.selectedStart.edgeStart
      )
        break

      // 3 - tram stop
      if (node.node_type === "stop" && node.id !== this.selectedStart.edgeStart)
        break

      // 4 - out of rectangle bounds
      if (!(nodeNeighbors[0] in nodes)) break

      node = nodes[nodeNeighbors[0]].details
    }
  }

  private highlightSelectedPath(nodes: Record<number, GraphNode>) {
    this.tracksLayer.clearLayers()

    const latlngs: LatLngExpression[] = this.selectedNodes.map(nodeID => [
      nodes[nodeID].details.lat,
      nodes[nodeID].details.lon,
    ])

    const polyline = L.polyline(latlngs, { weight: 6, color: "red" })
    polyline.addTo(this.tracksLayer)
  }

  private onPolylineClick(
    edgeStart: number,
    edgeEnd: number,
    nodes: Record<number, GraphNode>,
  ) {
    if (!this.selectedStart) {
      this.selectedStart = { edgeStart, edgeEnd }
      return
    }

    this.findSelectionEndAndPath(edgeEnd, nodes)

    if (this.selectedNodes.length === 0) return

    this.highlightSelectedPath(nodes)

    this.onSpeedDialogOpen({
      onCancel: () => {
        this.tracksLayer.clearLayers()
        this.drawTracks(nodes)
      },
      onSpeedSave: (newMaxSpeed: number) => {
        this.saveNewMaxSpeed(newMaxSpeed / 3.59) // km/h to m/s
        this.tracksLayer.clearLayers()
        this.drawTracks(nodes)
      },
    })
  }

  private saveNewMaxSpeed(newMaxSpeed: number) {
    for (let i = 0; i < this.selectedNodes.length - 1; i++) {
      const nodeID = Number(this.selectedNodes[i])
      const nextNodeID = Number(this.selectedNodes[i + 1])

      if (!this.modifiedNodes[nodeID]) {
        this.modifiedNodes[nodeID] = { neighborMaxSpeed: {} }
      }
      this.modifiedNodes[nodeID].neighborMaxSpeed[nextNodeID] = newMaxSpeed
    }
  }
}
