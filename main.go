package main

import (
	"embed"

	"github.com/entro314-labs/update-ng/cmd"
)

//go:embed bin/*
var scripts embed.FS

func main() {
	cmd.Execute(scripts)
}
