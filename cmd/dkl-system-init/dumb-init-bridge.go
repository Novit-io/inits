package main

// void handleSignals();
import "C"

func handleChildren() {
	C.handleSignals()
}
