package main

import "fmt"

const JOYP uint16 = 0xFF00

type Button byte

const (
	ButtonA      Button = 0
	ButtonB      Button = 1
	ButtonSelect Button = 2
	ButtonStart  Button = 3

	ButtonRight Button = 4
	ButtonLeft  Button = 5
	ButtonUp    Button = 6
	ButtonDown  Button = 7
)

func (b Button) String() string {
	switch b {
	case ButtonA:
		return "A"
	case ButtonB:
		return "B"
	case ButtonSelect:
		return "Select"
	case ButtonStart:
		return "Start"
	case ButtonRight:
		return "Right"
	case ButtonLeft:
		return "Left"
	case ButtonUp:
		return "Up"
	case ButtonDown:
		return "Down"
	default:
		return "Unknown Button"
	}
}

// Input contains the input values for the d-pad and buttons
// the d-pad is contained in the lwoer nibble and the buttons are in the upper nibble
type Input byte

func NewInput() *Input {
	v := Input(0xFF)
	return &v
}

func (i *Input) PressButton(cpu *CPU, button Button) {
	*i = Input(ResetBit(byte(*i), byte(button)))
	cpu.requestInterrupt(4)
}

func (i *Input) ReleaseButton(button Button) {
	*i = Input(SetBit(byte(*i), byte(button)))
}

func (i *Input) GetInput(enabled byte) byte {
	switch enabled {
	case 0x10:
		return enabled | byte(*i)&0x0F
	case 0x20:
		return enabled | (byte(*i) & 0xF0 >> 4)
	default:
		return enabled | 0x0F
	}
}
