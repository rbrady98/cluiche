package cartridge

const (
	mbc1RAMEnableRegister = 0x2000
	mbc1ROMBankRegister   = 0x4000
	mbc1RAMBankRegister   = 0x6000

	romOffset   = 0x4000
	ramOffset   = 0xA000
	romBankSize = 0x4000
	ramBankSize = 0x2000
)

type MBC1 struct {
	rom     []byte
	romBank int

	ram        []byte
	ramBank    int
	ramEnabled bool
}

func NewMBC1(rom []byte) *MBC1 {
	return &MBC1{
		rom:     rom,
		ram:     make([]byte, 4*ramBankSize),
		romBank: 1,
	}
}

// Read implements BankController.
func (m *MBC1) Read(addr uint16) byte {
	switch {
	case addr < 0x4000: // fixed bank 0
		return m.rom[addr]
	case addr < 0x8000: // variable rom bank
		offset := uint16(m.romBank*romBankSize) - romOffset
		return m.rom[addr+offset]
	default: // reading from the ram bank
		offset := uint16(m.ramBank*ramBankSize) - ramOffset
		return m.ram[addr+offset]
	}
}

// WriteRAM implements BankController.
func (m *MBC1) WriteRAM(addr uint16, value byte) {
	if m.ramEnabled {
		offset := uint16(m.ramBank*ramBankSize) - ramOffset
		m.ram[addr+offset] = value
	}
}

// WriteROM implements BankController.
func (m *MBC1) WriteROM(addr uint16, value byte) {
	switch {
	case addr < mbc1RAMEnableRegister:
		m.ramEnabled = (value & 0xF) == 0xA

	case addr < mbc1ROMBankRegister:
		// TODO: handle large roms with added banking
		b := 0xE0 | (value & 0x1F)
		m.romBank = translateBank(b)

	case addr < mbc1RAMBankRegister:
		m.ramBank = int(value & 0x3)
	}
}

// translateBank handles translating invalid rom bank numbers to valid ones.
func translateBank(bank byte) int {
	if bank == 0x00 || bank == 0x20 || bank == 0x40 || bank == 0x60 {
		return int(bank) + 1
	}

	return int(bank)
}
