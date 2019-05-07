package main

import (
	"github.com/farmer-hutao/k6s/pkg/agent"
)

func main() {
	router := agent.NewGinEngine()
	router.Run(":23456")
}
