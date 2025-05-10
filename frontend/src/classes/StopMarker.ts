import { CircleMarker, CircleMarkerOptions } from "leaflet";

export class StopMarker extends CircleMarker {
    private selected = false;

    constructor(lat: number, lon: number, name: string | null | undefined) {
        super([lat, lon], {
        radius: 5,
        fill: true,
        color: "darkblue",
        fillColor: "blue",
        weight: 1,
        opacity: 1,
        fillOpacity: 0.8,
      })
      this.bindTooltip(name ?? "Unknown stop", {
            permanent: false,
            direction: 'top',
        });
    }

    public onStopClick(callback: () => void) {
        this.on("click", callback);
    }

    public setSelected(value: boolean) {
        this.selected = value;
        this.setStyle({
        fillColor: value ? "green" : "blue",
        color:     value ? "darkgreen" : "darkblue",
        });
        if (value) this.bringToFront();
    }

    public isSelected() {
        return this.selected;
    }
}
