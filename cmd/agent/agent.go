package main

import (
	"farmer-hutao/k6s/pkg/agent/pkg"
)

func main() {
	router := pkg.NewGinEngine()
	router.Run(":23456")
}
