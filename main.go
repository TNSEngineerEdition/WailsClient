package main

import (
	"context"
	"embed"

	"github.com/TNSEngineerEdition/WailsClient/pkg/api"
	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation/tram"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	apiClient := api.NewAPIClient()

	city := city.City{}
	simulation := simulation.NewSimulation(&apiClient, &city)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TNSEngineerEdition",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			simulation.SetContext(ctx)
		},
		Bind: []any{
			&apiClient,
			&city,
			&simulation,
		},
		EnumBind: []any{
			api.Weekdays,
			tram.TramStates,
		},
		LogLevel: logger.WARNING,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
