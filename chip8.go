package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
)

var Chip8FontSet = [80]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80}

type Chip8 struct {
	gfx        [2048]uint8
	memory     [4096]uint8
	V          [16]uint8
	I          uint16
	pc         uint16
	sp         uint16
	delayTimer uint8
	soundTimer uint8
	stack      [16]uint16
	key        [16]uint16
	drawFlag   bool
}

func NewChip8() *Chip8 {
	cp8 := new(Chip8)
	cp8.pc = 0x200

	for i := 0; i < 80; i++ {
		cp8.memory[i] = Chip8FontSet[i]
	}

	return cp8
}

func LoadProgram(cp8 *Chip8, progName string) {
	buf, err := ioutil.ReadFile(progName)
	if err != nil {
		fmt.Println("Can't read ROM:", err)
	}

	for i := 0; i < len(buf); i++ {
		cp8.memory[512+i] = buf[i]
	}
}

func EmulateCycle(cp8 *Chip8) {
	opcode := uint16(cp8.memory[cp8.pc])<<8 | uint16(cp8.memory[cp8.pc+1])

	cp8.pc = cp8.pc + 2  // Increment program counter
	cp8.drawFlag = false // Draw flag not set, unless instructions 0x00e0 or 0xdxyn

	x := int((opcode & 0x0f00) >> 8)
	y := int((opcode & 0x00f0) >> 4)
	switch opcode & 0xf000 {
	case 0x0000:
		switch opcode & 0x0fff {
		case 0x00e0:
			for i := 0; i < 2048; i++ {
				cp8.gfx[i] = 0x00
			}
			cp8.drawFlag = true
		case 0x00EE:
			cp8.pc = cp8.stack[cp8.sp]
			cp8.sp = cp8.sp - 1
		}
	case 0x1000:
		cp8.pc = opcode & 0x0fff
	case 0x2000:
		cp8.sp = cp8.sp + 1
		cp8.stack[cp8.sp] = cp8.pc
		cp8.pc = opcode & 0x0fff
	case 0x3000:
		if uint16(cp8.V[x]) == opcode&0x00ff {
			cp8.pc = cp8.pc + 2
		}
	case 0x4000:
		if uint16(cp8.V[x]) != opcode&0x00ff {
			cp8.pc = cp8.pc + 2
		}
	case 0x5000:
		if cp8.V[x] == cp8.V[y] {
			cp8.pc = cp8.pc + 2
		}
	case 0x6000:
		cp8.V[x] = uint8(opcode & 0x00ff)
	case 0x7000:
		cp8.V[x] = cp8.V[x] + uint8(opcode&0x00ff)
	case 0x8000:
		switch opcode & 0x000f {
		case 0x0000:
			cp8.V[x] = cp8.V[y]
		case 0x0001:
			cp8.V[x] = cp8.V[x] | cp8.V[y]
		case 0x0002:
			cp8.V[x] = cp8.V[x] & cp8.V[y]
		case 0x0003:
			cp8.V[x] = cp8.V[x] ^ cp8.V[y]
		case 0x0004:
			cp8.V[x] = uint8((uint16(cp8.V[x]) + uint16(cp8.V[y])) & 0x00ff)
			if uint16(cp8.V[x])+uint16(cp8.V[y]) > 0x00ff {
				cp8.V[int(0xf)] = 1 // Set the VF register
			} else {
				cp8.V[int(0xf)] = 0
			}
		case 0x0005:
			if cp8.V[x] > cp8.V[y] {
				cp8.V[int(0xf)] = 1 // Set the VF register
			} else {
				cp8.V[int(0xf)] = 0
			}
			cp8.V[x] = uint8((int16(cp8.V[x]) - int16(cp8.V[y])) & 0x00ff)
		case 0x0006:
			if cp8.V[x]&0x0001 == 0x0001 {
				cp8.V[int(0xf)] = 1 // Set the VF register
			} else {
				cp8.V[int(0xf)] = 0
			}
			cp8.V[x] = cp8.V[x] >> 1
		case 0x0007:
			if uint16(cp8.V[y]) > uint16(cp8.V[x]) {
				cp8.V[int(0xf)] = 1 // Set the VF register
			} else {
				cp8.V[int(0xf)] = 0
			}
			cp8.V[x] = uint8((int16(cp8.V[y]) - int16(cp8.V[x])) & 0x00ff)
		case 0x000e:
			if cp8.V[x]&0x0080 == 0x0080 {
				cp8.V[int(0xf)] = 1 // Set the VF register
			} else {
				cp8.V[int(0xf)] = 0
			}
			cp8.V[x] = uint8((uint16(cp8.V[x]) << 1) & 0x00ff)
		}
	case 0x9000:
		if cp8.V[x] != cp8.V[y] {
			cp8.pc = cp8.pc + 2
		}
	case 0xa000:
		cp8.I = opcode & 0x0fff
	case 0xb000:
		cp8.pc = (opcode & 0x0fff) + uint16(cp8.V[0])
	case 0xc000:
		rand.Seed(time.Now().Unix())
		temp := uint8(rand.Intn(255))
		cp8.V[x] = temp & uint8(opcode&0x00ff)
	case 0xd000:
		n := int(opcode & 0x000f)
		cp8.V[int(0xf)] = 0
		for i := 0; i < n; i++ {
			for j := 0; j < 8; j++ {
				tempX := (int(cp8.V[x]) + j) % 64 // Handle wrapping of screen
				tempY := (int(cp8.V[y]) + i) % 32
				newVal := ((cp8.memory[int(cp8.I)+i] & uint8(0x01<<(7-uint8(j)))) >> (7 - uint8(j)))
				if cp8.gfx[tempY*64+tempX]^newVal == 0x0001 {
					cp8.V[int(0xf)] = 1 // Set the VF register
				}
				cp8.gfx[tempY*64+tempX] = cp8.gfx[tempY*64+tempX] ^ newVal
			}
		}
		cp8.drawFlag = true
	case 0xe000:
		switch opcode & 0x000f {
		case 0x000e:
			if cp8.key[int(cp8.V[x])] == 0x0001 {
				cp8.pc = cp8.pc + 2
			}
		case 0x0001:
			if cp8.key[int(cp8.V[x])] == 0x0000 {
				cp8.pc = cp8.pc + 2
			}
		}
	case 0xf000:
		switch opcode & 0x00ff {
		case 0x0007:
			cp8.V[x] = cp8.delayTimer
		case 0x000a:
			temp := 0
			for temp == 0 {
				for i := 0; i < 16; i++ {
					if cp8.key[i] == 0x0001 {
						temp = 1
						cp8.V[x] = uint8(i)
						break
					}
				}
			}
		case 0x0015:
			cp8.delayTimer = cp8.V[x]
		case 0x0018:
			cp8.soundTimer = cp8.V[x]
		case 0x001e:
			cp8.I = cp8.I + uint16(cp8.V[x])
		case 0x0029:
			cp8.I = uint16(cp8.V[x] * 5)
		case 0x0033:
			hundreds := int(cp8.V[x]) / 100
			leftover := int(cp8.V[x]) % 100
			tens := leftover / 10
			ones := leftover % 10
			cp8.memory[int(cp8.I)] = uint8(hundreds)
			cp8.memory[int(cp8.I)+1] = uint8(tens)
			cp8.memory[int(cp8.I)+2] = uint8(ones)
		case 0x0055:
			for i := 0; i < x; i++ {
				cp8.memory[int(cp8.I)+i] = cp8.V[i]
			}
		case 0x0065:
			for i := 0; i < x; i++ {
				cp8.V[i] = cp8.memory[int(cp8.I)+i]
			}
		}
	default:
		fmt.Println("Unknown opcode: 0x%x", opcode)
	}

	if cp8.delayTimer > 0 {
		cp8.delayTimer = cp8.delayTimer - 1
	}

	if cp8.soundTimer > 0 {
		if cp8.soundTimer == 1 {
			fmt.Println("BEEP!")
		}
		cp8.soundTimer = cp8.soundTimer - 1
	}
}
