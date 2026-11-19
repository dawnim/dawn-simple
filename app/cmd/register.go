package cmd

import "github.com/prismgo/framework/console"

func CommandFactories() []console.CommandFactory {
	return []console.CommandFactory{func() console.Command { return &SupportDocsCommand{} }}
}
