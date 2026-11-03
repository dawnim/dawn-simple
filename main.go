package main

import (
	"context"
	"os"

	dawn "github.com/dawnim/dawn-server"
	"github.com/dawnim/dawn-simple/bootstrap"
	"github.com/prismgo/framework/console"
)

func main() {
	if err := dawn.HandleCommand(context.Background(), bootstrap.NewApplication(), os.Args); err != nil {
		console.Exit(err.Error())
	}
}
