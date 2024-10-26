package main

import "softbaer.dev/ass/cmd"

func main() {
	err := cmd.Run("db/dev.sqlite")

	if err != nil {
		panic(err)
	}
}
