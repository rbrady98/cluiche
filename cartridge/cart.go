package cartridge

import (
	"fmt"
)

type BankController interface {
	Read(addr uint16) byte
	WriteROM(addr uint16, value byte)
	WriteRAM(addr uint16, value byte)
}

type Cart struct {
	BankController
	title string
}

func NewCart(rom []byte) (*Cart, error) {
	var cart Cart

	// check what the cartridge type is
	cartType := rom[0x147]
	switch cartType {
	case 0x00:
		// create a rom only bank controller
		cart.BankController = NewROM(rom)
	case 0x01:
		// create a rom only bank controller
		cart.BankController = NewMBC1(rom)
	default:
		return nil, fmt.Errorf("unsupported rom type: %02X", cartType)
	}

	return &cart, nil
}

func (c *Cart) Title() string {
	if c.title != "" {
		return c.title
	}

	for i := uint16(0x134); i < 0x143; i++ {
		v := c.Read(i)
		if v == 0x00 {
			continue
		}

		c.title += string(v)
	}

	return c.title
}
