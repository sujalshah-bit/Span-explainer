package main

import (
	"fmt"

	"github.com/sujalshah-bit/span-explainer/internal/store"
)

func main() {
	_ = store.NewStore()
	fmt.Print("Hello world")
}
