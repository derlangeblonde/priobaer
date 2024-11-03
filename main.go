package main

import "softbaer.dev/ass/cmd"

func main() {
	err := cmd.Run()

	if err != nil {
		panic(err)
	}
}
