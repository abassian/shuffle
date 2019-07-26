package main

import (
	cmd "github.com/abassian/shuffle/cmd/shl/commands"
)

func main() {
	rootCmd := cmd.RootCmd
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
