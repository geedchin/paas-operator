package main

import (
	. "github.com/farmer-hutao/k6s/pkg/agent"
)

func main() {
	app := NewGinEngine()
	app.Run(":3335")
}
