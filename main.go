package main

import (
	"embed"

	"github.com/TNSEngineerEdition/WailsClient/pkg/city"
	"github.com/TNSEngineerEdition/WailsClient/pkg/simulation"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	city := city.City{}
	simulation := simulation.NewSimulation(&city)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "TNSEngineerEdition",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
			&city,
			&simulation,
		},
		LogLevel: logger.WARNING,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
