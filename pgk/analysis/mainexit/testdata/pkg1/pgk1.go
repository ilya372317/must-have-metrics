package main

import "os"

func main() {
	os.Exit(1)       // want "calling Exit function of os package not recomended"
	defer os.Exit(1) // want "calling Exit function of os package not recomended"
	go os.Exit(1)    // want "calling Exit function of os package not recomended"
}
