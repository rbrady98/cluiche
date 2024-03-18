package cartridge

type MBC3 struct {
	rom     []byte
	romBank int

	ram        []byte
	ramBank    int
	ramEnabled bool
}

func NewMBC3(rom []byte) *MBC3 {
	return &MBC3{
		rom:     rom,
		ram:     make([]byte, 4*ramBankSize),
		romBank: 1,
	}
}

func (m *MBC3) Read(addr uint16) byte {
	switch {
	case addr < 0x4000: // fixed bank 0
		return m.rom[addr]
	case addr < 0x8000: // variable rom bank
		offset := uint32(m.romBank*romBankSize) - romOffset
		return m.rom[uint32(addr)+offset]
	default: // reading from the ram bank
		offset := uint32(m.ramBank*ramBankSize) - ramOffset
		return m.ram[uint32(addr)+offset]
	}
}

func (m *MBC3) WriteROM(addr uint16, value byte) {
	switch {
	case addr < mbc1RAMEnableRegister:
		m.ramEnabled = (value & 0xF) == 0xA

	case addr < mbc1ROMBankRegister:
		// TODO: handle large roms with added banking
		if value == 0x00 {
			value = 0x01
		}

		m.romBank = int(value)

	case addr < mbc1RAMBankRegister:
		m.ramBank = int(value & 0x3)
	}
}

func (m *MBC3) WriteRAM(addr uint16, value byte) {
	if m.ramEnabled {
		offset := uint32(m.ramBank*ramBankSize) - ramOffset
		m.ram[uint32(addr)+offset] = value
	}
}
