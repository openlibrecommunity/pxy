package main

import (
	"log/slog"
	"os"

	"github.com/openlibrecommunity/pxy/internal/pxyapp"
)

func main() {
	if err := pxyapp.Run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
