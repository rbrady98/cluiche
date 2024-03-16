package cartridge

// ROM is the most basic Gameboy cartridge type,
// it does not have banking or RAM
type ROM struct {
	// TODO: check if we should use an array not a slice
	rom []byte
}

func NewROM(rom []byte) *ROM {
	return &ROM{
		rom: rom,
	}
}

// Read implements BankController.
func (r *ROM) Read(addr uint16) byte {
	return r.rom[addr]
}

// WriteRAM implements BankController.
// Noop as basic ROM does not support RAM
func (r *ROM) WriteRAM(addr uint16, value byte) {}

// WriteROM implements BankController.
// Noop as basic ROM does not support banking
func (r *ROM) WriteROM(addr uint16, value byte) {}
