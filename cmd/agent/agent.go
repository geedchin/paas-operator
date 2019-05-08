package main

import (
	"github.com/farmer-hutao/k6s/pkg/agent"
)

func main() {
	app := agent.NewGinEngine()
	app.Run(":3335")
}
