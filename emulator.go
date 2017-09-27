package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/go-gl/glfw/v3.2/glfw"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Filepath of program to be loaded: ")
	filepath, _ := reader.ReadString('\n')

	runtime.LockOSThread() // Always execute in the same OS thread

	window := InitGlfw()
	defer glfw.Terminate()
	program := InitOpenGL()

	cells := MakeCells()

	cp8 := NewChip8()
	LoadProgram(cp8, strings.TrimSpace(filepath))

	for !window.ShouldClose() {
		EmulateCycle(cp8)

		if cp8.drawFlag {
			DrawDisplay(cp8, window, cells, program)
		}

		// Store key press state
	}
}
