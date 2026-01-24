package main

import (
	"github.com/goyek/goyek/v3"
	"github.com/goyek/x/boot"
)

func main() {
	goyek.SetDefault(tests)
	boot.Main()
}
