package main

import "os"

func main() {
	os.Exit(1)       // want "calling Exit function of os package not recommended"
	defer os.Exit(1) // want "calling Exit function of os package not recommended"
	go os.Exit(1)    // want "calling Exit function of os package not recommended"
	someFunc()
}

func someFunc() {}
