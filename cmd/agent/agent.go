package main

import (
	"github.com/farmer-hutao/k6s/pkg/agent"
)

func main() {
	app := agent.NewGinEngine()
	go agent.TryCheck()
	app.Run(":3335")
}
