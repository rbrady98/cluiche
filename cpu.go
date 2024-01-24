package main

import "fmt"

type CPU struct {
	registers *Registers
	memory    *Memory

	pc uint16
	sp uint16

	halted            bool
	interruptsEnabled bool
}

func NewCPU(mem *Memory) *CPU {
	return &CPU{
		registers: &Registers{},
		memory:    mem,
		pc:        0x0100,
		sp:        0xFFFE,
	}
}

func (c *CPU) tick() {
	opcode := c.readNext()
	prefixed := opcode == 0xCB
	if prefixed {
		opcode = c.readNext()
	}

	if prefixed {
		fmt.Printf("opcode: 0xCB 0x%x\n", opcode)
	} else {
		fmt.Printf("opcode: 0x%x\n", opcode)
	}

	nextAddr := c.execute(opcode, prefixed)
	c.pc = nextAddr
}

// execute matches an opcode to an instruction
func (c *CPU) execute(op byte, prefixed bool) uint16 {
	// instructions which are not prefixed with 0xCB
	if !prefixed {
		switch op {
		case 0x00: // NOP
		case 0x10: // STOP
			c.halted = true
			c.readNext()
		case 0x2F: // CPL
			c.registers.a = ^c.registers.a
			c.registers.flags.Subtract = true
			c.registers.flags.HalfCarry = true
		case 0x01: // LD BC,nn
			c.registers.setBC(c.readNext16())
		case 0x11: // LD DE,nn
			c.registers.setDE(c.readNext16())
		case 0x21: // LD HL,nn
			c.registers.setHL(c.readNext16())
		case 0x31: // LD SP,nn
			c.sp = c.readNext16()
		case 0x7F: // LD A,A
			// self assign just skip
		case 0x47: // LD B,A
			c.registers.b = c.registers.a
		case 0x4F: // LD C,A
			c.registers.c = c.registers.a
		case 0x57: // LD D,A
			c.registers.d = c.registers.a
		case 0x5F: // LD E,A
			c.registers.e = c.registers.a
		case 0x67: // LD H,A
			c.registers.h = c.registers.a
		case 0x6F: // LD L,A
			c.registers.l = c.registers.a
		case 0x02: // LD (BC),A
			c.memory.Write(c.registers.getBC(), c.registers.a)
		case 0x12: // LD (DE),A
			c.memory.Write(c.registers.getDE(), c.registers.a)
		case 0x77: // LD HL,A
			c.memory.Write(c.registers.getHL(), c.registers.a)
		case 0xEA: // LD nn,A
			c.memory.Write(c.readNext16(), c.registers.a)
		case 0x2A: // LD A,HL+
			c.registers.a = c.memory.Read(c.registers.getHL())
			c.registers.setHL(c.inc16(c.registers.getHL()))
		case 0x3E: // LD A,#
			c.registers.a = c.readNext()
		case 0xE0: // LDH n,A = LD (0xFF00+n),A
			n := c.readNext()
			c.memory.Write(0xFF00+uint16(n), c.registers.a)
		case 0xF0: // LDH A,n = LD A,(0xFF00+n)
			n := c.readNext()
			c.registers.a = c.memory.Read(0xFF00 + uint16(n))
		case 0xC3: // JP nn
			c.jump(c.readNext16())
		case 0x18: // JR n
			c.jump(c.pc + uint16(c.readNext()))
		case 0x20: // JR NZ
			if !c.registers.flags.Zero {
				c.jump(c.pc + uint16(c.readNext()))
			}
		case 0x28: // JR Z
			if c.registers.flags.Zero {
				c.jump(c.pc + uint16(c.readNext()))
			}
		case 0x30: // JR NC
			if !c.registers.flags.Carry {
				c.jump(c.pc + uint16(c.readNext()))
			}
		case 0x38: // JR C
			if c.registers.flags.Carry {
				c.jump(c.pc + uint16(c.readNext()))
			}
		case 0xf3: // DI
			c.interruptsEnabled = false
		case 0xfB: // DI
			c.interruptsEnabled = true
		case 0x3C: // INC A
			c.registers.a = c.inc(c.registers.a)
		case 0x04: // INC B
			c.registers.b = c.inc(c.registers.b)
		case 0x0C: // INC C
			c.registers.c = c.inc(c.registers.c)
		case 0x14: // INC D
			c.registers.d = c.inc(c.registers.d)
		case 0x1C: // INC E
			c.registers.e = c.inc(c.registers.e)
		case 0x24: // INC H
			c.registers.h = c.inc(c.registers.h)
		case 0x2C: // INC L
			c.registers.l = c.inc(c.registers.l)
		case 0x34: // INC (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.inc(v))
		case 0xc9: // RET
			c.ret()
		case 0xF5: // PUSH AF
			c.push(c.registers.getAF())
		case 0xC5: // PUSH BC
			c.push(c.registers.getBC())
		case 0xD5: // PUSH DE
			c.push(c.registers.getDE())
		case 0xE5: // PUSH HL
			c.push(c.registers.getHL())
		case 0xF1: // POP AF
			c.registers.setAF(c.pop())
		case 0xC1: // POP BC
			c.registers.setBC(c.pop())
		case 0xD1: // POP DE
			c.registers.setDE(c.pop())
		case 0xE1: // POP HL
			c.registers.setHL(c.pop())
		case 0x03: // INC BC
			c.registers.setBC(c.inc16(c.registers.getBC()))
		case 0x13: // INC DE
			c.registers.setDE(c.inc16(c.registers.getDE()))
		case 0x23: // INC HL
			c.registers.setHL(c.inc16(c.registers.getHL()))
		case 0x33: // INC SP
			c.sp = (c.inc16(c.sp))
		case 0xA0: // AND A,A
			c.registers.a = c.and(c.registers.a, c.registers.a)
		case 0xAF: // XOR A
			c.registers.a = c.xor(c.registers.a, c.registers.a)
		case 0xA8: // XOR B
			c.registers.a = c.xor(c.registers.a, c.registers.b)
		case 0xA9: // XOR C
			c.registers.a = c.xor(c.registers.a, c.registers.c)
		case 0xAA: // XOR D
			c.registers.a = c.xor(c.registers.a, c.registers.d)
		case 0xAB: // XOR E
			c.registers.a = c.xor(c.registers.a, c.registers.e)
		case 0xAC: // XOR H
			c.registers.a = c.xor(c.registers.a, c.registers.h)
		case 0xAD: // XOR L
			c.registers.a = c.xor(c.registers.a, c.registers.l)
		case 0xAE: // XOR (HL)
			v := c.memory.Read(c.registers.getHL())
			c.registers.a = c.xor(c.registers.a, v)
		case 0xEE: // XOR n
			c.registers.a = c.xor(c.registers.a, c.readNext())
		case 0x87: // ADD A,A
			c.registers.a = c.add(c.registers.a, c.registers.a)
		case 0x80: // ADD A,B
			c.registers.a = c.add(c.registers.a, c.registers.b)
		case 0x81: // ADD A,C
			c.registers.a = c.add(c.registers.a, c.registers.c)
		case 0x82: // ADD A,D
			c.registers.a = c.add(c.registers.a, c.registers.d)
		case 0x83: // ADD A,E
			c.registers.a = c.add(c.registers.a, c.registers.e)
		case 0x84: // ADD A,H
			c.registers.a = c.add(c.registers.a, c.registers.h)
		case 0x85: // ADD A,L
			c.registers.a = c.add(c.registers.a, c.registers.l)
		case 0x86: // ADD A,HL
			c.registers.a = c.add(c.registers.a, c.memory.Read(c.registers.getHL()))
		case 0xC6: // ADD A,#
			c.registers.a = c.add(c.registers.a, c.readNext())
		case 0x8F: // ADC A,A
			c.registers.a = c.addC(c.registers.a, c.registers.a)
		case 0x88: // ADC A,B
			c.registers.a = c.addC(c.registers.a, c.registers.b)
		case 0x89: // ADC A,C
			c.registers.a = c.addC(c.registers.a, c.registers.c)
		case 0x8A: // ADC A,D
			c.registers.a = c.addC(c.registers.a, c.registers.d)
		case 0x8B: // ADC A,E
			c.registers.a = c.addC(c.registers.a, c.registers.e)
		case 0x8C: // ADC A,H
			c.registers.a = c.addC(c.registers.a, c.registers.h)
		case 0x8D: // ADC A,L
			c.registers.a = c.addC(c.registers.a, c.registers.l)
		case 0x8E: // ADC A,HL
			c.registers.a = c.addC(c.registers.a, c.memory.Read(c.registers.getHL()))
		case 0xCE: // ADC A,#
			c.registers.a = c.addC(c.registers.a, c.readNext())
		case 0x09: // ADD HL,BC
			c.registers.setHL(c.add16(c.registers.getHL(), c.registers.getBC()))
		case 0x19: // ADD HL,DE
			c.registers.setHL(c.add16(c.registers.getHL(), c.registers.getDE()))
		case 0x29: // ADD HL,HL
			c.registers.setHL(c.add16(c.registers.getHL(), c.registers.getHL()))
		case 0x39: // ADD HL,SP
			c.registers.setHL(c.add16(c.registers.getHL(), c.sp))
		case 0x3D: // DEC A
			c.registers.a = c.dec(c.registers.a)
		case 0x05: // DEC B
			c.registers.b = c.dec(c.registers.b)
		case 0x0D: // DEC C
			c.registers.c = c.dec(c.registers.c)
		case 0x15: // DEC D
			c.registers.d = c.dec(c.registers.d)
		case 0x1D: // DEC E
			c.registers.e = c.dec(c.registers.e)
		case 0x25: // DEC H
			c.registers.h = c.dec(c.registers.h)
		case 0x2D: // DEC L
			c.registers.l = c.dec(c.registers.l)
		case 0x35: // DEC (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.dec(v))
		case 0x0B: // DEC BC
			c.registers.setBC(c.dec16(c.registers.getBC()))
		case 0x1B: // DEC DE
			c.registers.setDE(c.dec16(c.registers.getDE()))
		case 0x2B: // DEC HL
			c.registers.setHL(c.dec16(c.registers.getHL()))
		case 0x3B: // DEC SP
			c.sp = c.dec16(c.sp)
		case 0xC7: // RST 0x00
			c.rst(0x00)
		case 0xCF: // RST 0x08
			c.rst(0x08)
		case 0xD7: // RST 0x10
			c.rst(0x10)
		case 0xDF: // RST 0x18
			c.rst(0x18)
		case 0xE7: // RST 0x20
			c.rst(0x20)
		case 0xEF: // RST 0x28
			c.rst(0x28)
		case 0xF7: // RST 0x30
			c.rst(0x30)
		case 0xFF: // RST 0x38
			c.rst(0x38)
		case 0xCD: // CALL nn
			c.call()
		case 0x06: // LD B,n
			c.registers.b = c.readNext()
		case 0x0E: // LD C,n
			c.registers.c = c.readNext()
		case 0x16: // LD D,n
			c.registers.d = c.readNext()
		case 0x1E: // LD E,n
			c.registers.e = c.readNext()
		case 0x26: // LD H,n
			c.registers.h = c.readNext()
		case 0x2E: // LD L,n
			c.registers.l = c.readNext()
		case 0x78: // LD A,B
			c.registers.a = c.registers.b
		case 0x79: // LD A,C
			c.registers.a = c.registers.c
		case 0x7A: // LD A,D
			c.registers.a = c.registers.d
		case 0x7B: // LD A,E
			c.registers.a = c.registers.e
		case 0x7C: // LD A,H
			c.registers.a = c.registers.h
		case 0x7D: // LD A,L
			c.registers.a = c.registers.l
		case 0x7E: // LD A,HL
			c.registers.a = c.memory.Read(c.registers.getHL())
		case 0x40: // LD B,B
			// ignore self assign
		case 0x41: // LD B,C
			c.registers.b = c.registers.c
		case 0x42: // LD B,D
			c.registers.b = c.registers.d
		case 0x43: // LD B,E
			c.registers.b = c.registers.e
		case 0x44: // LD B,H
			c.registers.b = c.registers.h
		case 0x45: // LD B,L
			c.registers.b = c.registers.l
		case 0x46: // LD B,HL
			c.registers.b = c.memory.Read(c.registers.getHL())
		case 0x48: // LD C,B
			c.registers.c = c.registers.b
		case 0x49: // LD C,C
			// ignore self assign
		case 0x4A: // LD C,D
			c.registers.c = c.registers.d
		case 0x4B: // LD C,E
			c.registers.c = c.registers.e
		case 0x4C: // LD C,H
			c.registers.c = c.registers.h
		case 0x4D: // LD C,L
			c.registers.c = c.registers.l
		case 0x4E: // LD C,HL
			c.registers.c = c.memory.Read(c.registers.getHL())
		case 0x50: // LD D,B
			c.registers.d = c.registers.b
		case 0x51: // LD D,C
			c.registers.d = c.registers.c
		case 0x52: // LD D,D
			// ignore self assign
		case 0x53: // LD D,E
			c.registers.d = c.registers.e
		case 0x54: // LD D,H
			c.registers.d = c.registers.h
		case 0x55: // LD D,L
			c.registers.d = c.registers.l
		case 0x56: // dD D,HL
			c.registers.d = c.memory.Read(c.registers.getHL())
		case 0x58: // LD E,B
			c.registers.e = c.registers.b
		case 0x59: // LD E,C
			c.registers.e = c.registers.c
		case 0x5A: // LD E,D
			c.registers.e = c.registers.d
		case 0x5B: // LD E,E
			// ignore self assign
		case 0x5C: // LD E,H
			c.registers.e = c.registers.h
		case 0x5D: // LD E,L
			c.registers.e = c.registers.l
		case 0x5E: // LD E,HL
			c.registers.e = c.memory.Read(c.registers.getHL())
		case 0x60: // LD H,B
			c.registers.h = c.registers.b
		case 0x61: // LD H,C
			c.registers.h = c.registers.c
		case 0x62: // LD H,D
			c.registers.h = c.registers.d
		case 0x63: // LD H,E
			c.registers.h = c.registers.e
		case 0x64: // LD H,H
			// ignore self assign
		case 0x65: // LD H,L
			c.registers.h = c.registers.l
		case 0x66: // LD H,HL
			c.registers.h = c.memory.Read(c.registers.getHL())
		case 0x68: // LD L,B
			c.registers.l = c.registers.b
		case 0x69: // LD L,C
			c.registers.l = c.registers.c
		case 0x6A: // LD L,D
			c.registers.l = c.registers.d
		case 0x6B: // LD L,E
			c.registers.l = c.registers.e
		case 0x6C: // LD L,H
			c.registers.l = c.registers.h
		case 0x6D: // LD L,L
			// ignore self assign
		case 0x6E: // LD L,HL
			c.registers.l = c.memory.Read(c.registers.getHL())
		case 0x70: // LD HL,B
			c.memory.Write(c.registers.getHL(), c.registers.b)
		case 0x71: // LD HL,C
			c.memory.Write(c.registers.getHL(), c.registers.c)
		case 0x72: // LD HL,D
			c.memory.Write(c.registers.getHL(), c.registers.d)
		case 0x73: // LD HL,E
			c.memory.Write(c.registers.getHL(), c.registers.e)
		case 0x74: // LD HL,H
			c.memory.Write(c.registers.getHL(), c.registers.h)
		case 0x75: // LD HL,L
			c.memory.Write(c.registers.getHL(), c.registers.l)
		case 0x36: // LD HL,n
			c.memory.Write(c.registers.getHL(), c.readNext())
		case 0x1F:
			c.registers.a = c.rra()
		default:
			panic(fmt.Sprintf("unimplemented opcode: %#2x\n", op))
		}
	} else {
		switch op {
		case 0x00:
			fmt.Println("no op")
		case 0x3F: // SRL A
			c.registers.a = c.srl(c.registers.a)
		case 0x38: // SRL B
			c.registers.b = c.srl(c.registers.b)
		case 0x39: // SRL C
			c.registers.c = c.srl(c.registers.c)
		case 0x3A: // SRL D
			c.registers.d = c.srl(c.registers.d)
		case 0x3B: // SRL E
			c.registers.e = c.srl(c.registers.e)
		case 0x3C: // SRL H
			c.registers.h = c.srl(c.registers.h)
		case 0x3D: // SRL L
			c.registers.l = c.srl(c.registers.l)
		case 0x3E: // SRL HL
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.srl(v))
		case 0x1F: // RR A
			c.registers.a = c.rr(c.registers.a)
		case 0x18: // RR B
			c.registers.b = c.rr(c.registers.b)
		case 0x19: // RR C
			c.registers.c = c.rr(c.registers.c)
		case 0x1A: // RR D
			c.registers.d = c.rr(c.registers.d)
		case 0x1B: // RR E
			c.registers.e = c.rr(c.registers.e)
		case 0x1C: // RR H
			c.registers.h = c.rr(c.registers.h)
		case 0x1D: // RR L
			c.registers.l = c.rr(c.registers.l)
		case 0x1E: // RR (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.rr(v))
		default:
			panic(fmt.Sprintf("unimplemented opcode: %#2x\n", op))
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
