package main

import (
	"embed"

	"github.com/openlibrecommunity/pxy/internal/guiservice"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := guiservice.New()
	err := wails.Run(&options.App{
		Title:  "pxy",
		Width:  1180,
		Height: 820,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 10, B: 13, A: 1},
		OnStartup:        app.Startup,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		println("err:", err.Error())
	}
}
