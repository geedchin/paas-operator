package main

import (
	. "github.com/farmer-hutao/k6s/pkg/agent"
)

func main() {
	router := NewGinEngine()
	router.Run(":23456")
}
