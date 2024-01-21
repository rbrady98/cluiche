package main

import "fmt"

type CPU struct {
	registers *Registers
	memory    *Memory

	pc uint16
	sp uint16

	halted bool
}

func (c *CPU) tick() {
	opcode := c.readNext()
	prefixed := opcode == 0xCB
	if prefixed {
		opcode = c.readNext()
	}

	if prefixed {
		fmt.Printf("opcode: 0xCB 0x%x", opcode)
	} else {
		fmt.Printf("opcode: 0x%x", opcode)
	}

	nextAddr := c.execute(opcode, prefixed)
	c.pc = nextAddr
}

func NewCPU(mem *Memory) *CPU {
	return &CPU{
		registers: &Registers{},
		memory:    mem,
	}
}

// execute matches an opcode to an instruction
func (c *CPU) execute(op byte, prefixed bool) uint16 {
	// instructions which are not prefixed with 0xCB
	if !prefixed {
		switch op {
		case 0x00:
			fmt.Println("no op")
		case 0xA0: // AND A,A
			c.registers.a = c.and(c.registers.a, c.registers.a)
			return c.pc + 1
		default:
			fmt.Printf("unimplemented opcode: %#2x\n", op)
		}
	} else {
		switch op {
		case 0x00:
			fmt.Println("no op")
		case 0xA0: // AND A,A
			c.registers.a = c.and(c.registers.a, c.registers.a)
			return c.pc + 1
		default:
			fmt.Printf("unimplemented opcode: cb %#2x\n", op)
			panic("boom!")
		}
	}

	return c.pc
}

// readNext reads the opcode at the program counter and increments the program counter
func (c *CPU) readNext() byte {
	fmt.Printf("Reading memory at 0x%x\n", c.pc)
	op := c.memory.Read(c.pc)
	c.pc++
	return op
}

// readNext16 reads the value at the next two addresses and increments the program counter
func (c *CPU) readNext16() uint16 {
	op1 := uint16(c.readNext())
	op2 := uint16(c.readNext())
	// immediate value are little endian
	return op2<<8 | op1
}
