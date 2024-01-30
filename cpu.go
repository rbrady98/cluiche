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
		registers: NewRegisters(),
		memory:    mem,
		pc:        0x0100,
		sp:        0xFFFE,
	}
}

// Update ticks the cpu, reading the next instruction and executing it
func (c *CPU) Update() {
	// fmt.Printf("pc: %x\n", c.pc)
	opcode := c.readNext()
	prefixed := opcode == 0xCB
	if prefixed {
		opcode = c.readNext()
	}

	// if prefixed {
	// 	fmt.Printf("opcode: 0xCB 0x%x\n", opcode)
	// } else {
	// 	fmt.Printf("opcode: 0x%x\n", opcode)
	// }

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
			c.registers.f.Subtract = true
			c.registers.f.HalfCarry = true
		case 0x01: // LD BC,nn
			c.registers.setBC(c.readNext16())
		case 0x11: // LD DE,nn
			c.registers.setDE(c.readNext16())
		case 0x21: // LD HL,nn
			c.registers.setHL(c.readNext16())
		case 0x31: // LD SP,nn
			c.sp = c.readNext16()
		case 0xF9: // LD SP,HL
			c.sp = c.registers.getHL()
		case 0xF8: // LD HL, SP=n
			addr := c.sp + uint16(int8(c.readNext()))
			c.registers.setHL(addr)

			c.registers.f.Zero = false
			c.registers.f.Subtract = false
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
		case 0xF2: // LD A,(C)
			n := 0xFF00 + uint16(c.registers.c)
			c.registers.a = c.memory.Read(n)
		case 0xE2: // LD (C),A
			c.memory.Write(0xFF00+uint16(c.registers.c), c.registers.a)
		case 0xC3: // JP nn
			c.jump(c.readNext16())
		case 0xC2: // JP NZ, nn
			addr := c.readNext16()
			if !c.registers.f.Zero {
				c.jump(addr)
			}
		case 0xCA: // JP Z, nn
			addr := c.readNext16()
			if c.registers.f.Zero {
				c.jump(addr)
			}
		case 0xD2: // JP NC, nn
			addr := c.readNext16()
			if !c.registers.f.Carry {
				c.jump(addr)
			}
		case 0xDA: // JP C, nn
			addr := c.readNext16()
			if c.registers.f.Carry {
				c.jump(addr)
			}
		case 0xE9: // JP (HL)
			c.jump(c.registers.getHL())
		case 0x18: // JR n
			addr := int16(c.pc) + int16(int8(c.readNext()))
			c.jump(uint16(addr))
		case 0x20: // JR NZ
			v := int8(c.readNext())
			if !c.registers.f.Zero {
				addr := int16(c.pc) + int16(v)
				c.jump(uint16(addr))
			}
		case 0x28: // JR Z
			v := int8(c.readNext())
			if c.registers.f.Zero {
				addr := int16(c.pc) + int16(v)
				c.jump(uint16(addr))
			}
		case 0x30: // JR NC
			v := int8(c.readNext())
			if !c.registers.f.Carry {
				addr := int16(c.pc) + int16(v)
				c.jump(uint16(addr))
			}
		case 0x38: // JR C
			v := int8(c.readNext())
			if c.registers.f.Carry {
				addr := int16(c.pc) + int16(v)
				c.jump(uint16(addr))
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
		case 0xC9: // RET
			c.ret()
		case 0xC0: // RET NZ
			if !c.registers.f.Zero {
				c.ret()
			}
		case 0xC8: // RET Z
			if c.registers.f.Zero {
				c.ret()
			}
		case 0xD0: // RET NC
			if !c.registers.f.Carry {
				c.ret()
			}
		case 0xD8: // RET C
			if c.registers.f.Carry {
				c.ret()
			}
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
		case 0xA7: // AND A
			c.registers.a = c.and(c.registers.a, c.registers.a)
		case 0xA0: // AND B
			c.registers.a = c.and(c.registers.a, c.registers.b)
		case 0xA1: // AND C
			c.registers.a = c.and(c.registers.a, c.registers.c)
		case 0xA2: // AND D
			c.registers.a = c.and(c.registers.a, c.registers.d)
		case 0xA3: // AND E
			c.registers.a = c.and(c.registers.a, c.registers.e)
		case 0xA4: // AND H
			c.registers.a = c.and(c.registers.a, c.registers.h)
		case 0xA5: // AND L
			c.registers.a = c.and(c.registers.a, c.registers.l)
		case 0xA6: // AND (HL)
			v := c.memory.Read(c.registers.getHL())
			c.registers.a = c.and(c.registers.a, v)
		case 0xE6: // AND n
			c.registers.a = c.and(c.registers.a, c.readNext())
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
		case 0xB7: // OR A
			c.registers.a = c.or(c.registers.a, c.registers.a)
		case 0xB0: // OR B
			c.registers.a = c.or(c.registers.a, c.registers.b)
		case 0xB1: // OR C
			c.registers.a = c.or(c.registers.a, c.registers.c)
		case 0xB2: // OR D
			c.registers.a = c.or(c.registers.a, c.registers.d)
		case 0xB3: // OR E
			c.registers.a = c.or(c.registers.a, c.registers.e)
		case 0xB4: // OR H
			c.registers.a = c.or(c.registers.a, c.registers.h)
		case 0xB5: // OR L
			c.registers.a = c.or(c.registers.a, c.registers.l)
		case 0xB6: // OR (HL)
			v := c.memory.Read(c.registers.getHL())
			c.registers.a = c.or(c.registers.a, v)
		case 0xF6: // OR n
			c.registers.a = c.or(c.registers.a, c.readNext())
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
			next := c.readNext16()
			c.call(next)
		case 0xC4: // CALL NZ,nn
			next := c.readNext16()
			if !c.registers.f.Zero {
				c.call(next)
			}
		case 0xCC: // CALL Z,nn
			next := c.readNext16()
			if c.registers.f.Zero {
				c.call(next)
			}
		case 0xD4: // CALL Nc,nn
			next := c.readNext16()
			if !c.registers.f.Carry {
				c.call(next)
			}
		case 0xDC: // CALL C,nn
			next := c.readNext16()
			if c.registers.f.Carry {
				c.call(next)
			}
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
		case 0x0A: // LD A,BC
			c.registers.a = c.memory.Read(c.registers.getBC())
		case 0x1A: // LD A,DE
			c.registers.a = c.memory.Read(c.registers.getDE())
		case 0x7E: // LD A,HL
			c.registers.a = c.memory.Read(c.registers.getHL())
		case 0xFA: // LD A,nn
			c.registers.a = c.memory.Read(c.readNext16())
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
		case 0x22: // LD (HLI),A
			c.memory.Write(c.registers.getHL(), c.registers.a)
			c.registers.setHL(c.inc16(c.registers.getHL()))
		case 0x32: // LD (HLD),A
			c.memory.Write(c.registers.getHL(), c.registers.a)
			c.registers.setHL(c.dec16(c.registers.getHL()))
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
		case 0x0F:
			c.rrca()
		case 0x1F:
			c.rra()
		case 0x08:
			addr := c.readNext16()
			c.memory.Write(addr, byte(c.sp&0xF))
			c.memory.Write(addr+1, byte(c.sp&0xF0))
		case 0xBF: // CP A
			c.cp(c.registers.a, c.registers.a)
		case 0xB8: // CP B
			c.cp(c.registers.b, c.registers.a)
		case 0xB9: // CP C
			c.cp(c.registers.c, c.registers.a)
		case 0xBA: // CP D
			c.cp(c.registers.d, c.registers.a)
		case 0xBB: // CP E
			c.cp(c.registers.e, c.registers.a)
		case 0xBC: // CP H
			c.cp(c.registers.h, c.registers.a)
		case 0xBD: // CP L
			c.cp(c.registers.l, c.registers.a)
		case 0xBE: // CP L
			v := c.memory.Read(c.registers.getHL())
			c.cp(v, c.registers.a)
		case 0xFE: // CP n
			c.cp(c.readNext(), c.registers.a)
		case 0x97: // SUB A
			c.registers.a = c.sub(c.registers.a, c.registers.a)
		case 0x90: // SUB B
			c.registers.a = c.sub(c.registers.a, c.registers.b)
		case 0x91: // SUB C
			c.registers.a = c.sub(c.registers.a, c.registers.c)
		case 0x92: // SUB D
			c.registers.a = c.sub(c.registers.a, c.registers.d)
		case 0x93: // SUB E
			c.registers.a = c.sub(c.registers.a, c.registers.e)
		case 0x94: // SUB H
			c.registers.a = c.sub(c.registers.a, c.registers.h)
		case 0x95: // SUB L
			c.registers.a = c.sub(c.registers.a, c.registers.l)
		case 0x96: // SUB (HL)
			c.registers.a = c.sub(c.registers.a, c.memory.Read(c.registers.getHL()))
		case 0xD6: // SUB n
			c.registers.a = c.sub(c.registers.a, c.readNext())
		case 0x9F: // SBC A,A
			c.registers.a = c.subC(c.registers.a, c.registers.a)
		case 0x98: // SBC A,B
			c.registers.a = c.subC(c.registers.a, c.registers.b)
		case 0x99: // SBC A,C
			c.registers.a = c.subC(c.registers.a, c.registers.c)
		case 0x9A: // SBC A,D
			c.registers.a = c.subC(c.registers.a, c.registers.d)
		case 0x9B: // SBC A,E
			c.registers.a = c.subC(c.registers.a, c.registers.e)
		case 0x9C: // SBC A,H
			c.registers.a = c.subC(c.registers.a, c.registers.h)
		case 0x9D: // SBC A,L
			c.registers.a = c.subC(c.registers.a, c.registers.l)
		case 0x9E: // SBC A,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.registers.a = c.subC(c.registers.a, v)
		case 0xDE: // SBC A,#
			c.registers.a = c.subC(c.registers.a, c.readNext())
		case 0x27: // DAA
			c.daa()
		case 0x37: // SCF
			c.registers.f.Carry = true
			c.registers.f.Subtract = false
			c.registers.f.HalfCarry = false
			c.registers.f.Carry = true
		case 0x07:
			c.rlca()
		case 0x17:
			c.rla()
		case 0x3F:
			c.ccf()
		default:
			panic(fmt.Sprintf("unimplemented opcode: %#2x\n", op))
		}
	} else {
		switch op {
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
		case 0x37: // SWAO A
			c.registers.a = c.swap(c.registers.a)
		case 0x30: // SWAO B
			c.registers.b = c.swap(c.registers.b)
		case 0x31: // SWAO C
			c.registers.c = c.swap(c.registers.c)
		case 0x32: // SWAO D
			c.registers.d = c.swap(c.registers.d)
		case 0x33: // SWAO E
			c.registers.e = c.swap(c.registers.e)
		case 0x34: // SWAO H
			c.registers.h = c.swap(c.registers.h)
		case 0x35: // SWAO L
			c.registers.l = c.swap(c.registers.l)
		case 0x36: // SWAO (HL)
			c.memory.Write(c.registers.getHL(), c.swap(c.memory.Read(c.registers.getHL())))
		case 0x07: // RLC A
			c.registers.a = c.rlc(c.registers.a)
		case 0x00: // RLC B
			c.registers.b = c.rlc(c.registers.b)
		case 0x01: // RLC C
			c.registers.c = c.rlc(c.registers.c)
		case 0x02: // RLC D
			c.registers.d = c.rlc(c.registers.d)
		case 0x03: // RLC E
			c.registers.e = c.rlc(c.registers.e)
		case 0x04: // RLC H
			c.registers.h = c.rlc(c.registers.h)
		case 0x05: // RLC L
			c.registers.l = c.rlc(c.registers.l)
		case 0x06: // RLC (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.rlc(v))
		case 0x0F: // RRC A
			c.registers.a = c.rrc(c.registers.a)
		case 0x08: // RRC B
			c.registers.b = c.rrc(c.registers.b)
		case 0x09: // RRC C
			c.registers.c = c.rrc(c.registers.c)
		case 0x0A: // RRC D
			c.registers.d = c.rrc(c.registers.d)
		case 0x0B: // RRC E
			c.registers.e = c.rrc(c.registers.e)
		case 0x0C: // RRC H
			c.registers.h = c.rrc(c.registers.h)
		case 0x0D: // RRC L
			c.registers.l = c.rrc(c.registers.l)
		case 0x0E: // RRC (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.rrc(v))
		case 0x17: // RL A
			c.registers.a = c.rl(c.registers.a)
		case 0x10: // RL B
			c.registers.b = c.rl(c.registers.b)
		case 0x11: // RL C
			c.registers.c = c.rl(c.registers.c)
		case 0x12: // RL D
			c.registers.d = c.rl(c.registers.d)
		case 0x13: // RL E
			c.registers.e = c.rl(c.registers.e)
		case 0x14: // RL H
			c.registers.h = c.rl(c.registers.h)
		case 0x15: // RL L
			c.registers.l = c.rl(c.registers.l)
		case 0x16: // RL (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.rl(v))
		case 0x27: // SLA A
			c.registers.a = c.sla(c.registers.a)
		case 0x20: // SLA B
			c.registers.b = c.sla(c.registers.b)
		case 0x21: // SLA C
			c.registers.c = c.sla(c.registers.c)
		case 0x22: // SLA D
			c.registers.d = c.sla(c.registers.d)
		case 0x23: // SLA E
			c.registers.e = c.sla(c.registers.e)
		case 0x24: // SLA H
			c.registers.h = c.sla(c.registers.h)
		case 0x25: // SLA L
			c.registers.l = c.sla(c.registers.l)
		case 0x26: // SLA (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.sla(v))
		case 0x2F: // SRA A
			c.registers.a = c.sra(c.registers.a)
		case 0x28: // SRA B
			c.registers.b = c.sra(c.registers.b)
		case 0x29: // SRA C
			c.registers.c = c.sra(c.registers.c)
		case 0x2A: // SRA D
			c.registers.d = c.sra(c.registers.d)
		case 0x2B: // SRA E
			c.registers.e = c.sra(c.registers.e)
		case 0x2C: // SRA H
			c.registers.h = c.sra(c.registers.h)
		case 0x2D: // SRA L
			c.registers.l = c.sra(c.registers.l)
		case 0x2E: // SRA (HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.sra(v))
		case 0x47: // BIT 0,A
			c.bit(0, c.registers.a)
		case 0x40: // BIT 0,B
			c.bit(0, c.registers.b)
		case 0x41: // BIT 0,C
			c.bit(0, c.registers.c)
		case 0x42: // BIT 0,D
			c.bit(0, c.registers.d)
		case 0x43: // BIT 0,E
			c.bit(0, c.registers.e)
		case 0x44: // BIT 0,H
			c.bit(0, c.registers.h)
		case 0x45: // BIT 0,L
			c.bit(0, c.registers.l)
		case 0x46: // BIT 0,(HL)
			c.bit(0, c.memory.Read(c.registers.getHL()))
		case 0x4F: // BIT 1,A
			c.bit(1, c.registers.a)
		case 0x48: // BIT 1,B
			c.bit(1, c.registers.b)
		case 0x49: // BIT 1,C
			c.bit(1, c.registers.c)
		case 0x4A: // BIT 1,D
			c.bit(1, c.registers.d)
		case 0x4B: // BIT 1,E
			c.bit(1, c.registers.e)
		case 0x4C: // BIT 1,H
			c.bit(1, c.registers.h)
		case 0x4D: // BIT 1,L
			c.bit(1, c.registers.l)
		case 0x4E: // BIT 1,(HL)
			c.bit(1, c.memory.Read(c.registers.getHL()))
		case 0x57: // BIT 2,A
			c.bit(2, c.registers.a)
		case 0x50: // BIT 2,B
			c.bit(2, c.registers.b)
		case 0x51: // BIT 2,C
			c.bit(2, c.registers.c)
		case 0x52: // BIT 2,D
			c.bit(2, c.registers.d)
		case 0x53: // BIT 2,E
			c.bit(2, c.registers.e)
		case 0x54: // BIT 2,H
			c.bit(2, c.registers.h)
		case 0x55: // BIT 2,L
			c.bit(2, c.registers.l)
		case 0x56: // BIT 2,(HL)
			c.bit(2, c.memory.Read(c.registers.getHL()))
		case 0x5F: // BIT 3,A
			c.bit(3, c.registers.a)
		case 0x58: // BIT 3,B
			c.bit(3, c.registers.b)
		case 0x59: // BIT 3,C
			c.bit(3, c.registers.c)
		case 0x5A: // BIT 3,D
			c.bit(3, c.registers.d)
		case 0x5B: // BIT 3,E
			c.bit(3, c.registers.e)
		case 0x5C: // BIT 3,H
			c.bit(3, c.registers.h)
		case 0x5D: // BIT 3,L
			c.bit(3, c.registers.l)
		case 0x5E: // BIT 3,(HL)
			c.bit(3, c.memory.Read(c.registers.getHL()))
		case 0x67: // BIT 4,A
			c.bit(4, c.registers.a)
		case 0x60: // BIT 4,B
			c.bit(4, c.registers.b)
		case 0x61: // BIT 4,C
			c.bit(4, c.registers.c)
		case 0x62: // BIT 4,D
			c.bit(4, c.registers.d)
		case 0x63: // BIT 4,E
			c.bit(4, c.registers.e)
		case 0x64: // BIT 4,H
			c.bit(4, c.registers.h)
		case 0x65: // BIT 4,L
			c.bit(4, c.registers.l)
		case 0x66: // BIT 4,(HL)
			c.bit(4, c.memory.Read(c.registers.getHL()))
		case 0x6F: // BIT 5,A
			c.bit(5, c.registers.a)
		case 0x68: // BIT 5,B
			c.bit(5, c.registers.b)
		case 0x69: // BIT 5,C
			c.bit(5, c.registers.c)
		case 0x6A: // BIT 5,D
			c.bit(5, c.registers.d)
		case 0x6B: // BIT 5,E
			c.bit(5, c.registers.e)
		case 0x6C: // BIT 5,H
			c.bit(5, c.registers.h)
		case 0x6D: // BIT 5,L
			c.bit(5, c.registers.l)
		case 0x6E: // BIT 5,(HL)
			c.bit(5, c.memory.Read(c.registers.getHL()))
		case 0x77: // BIT 6,A
			c.bit(6, c.registers.a)
		case 0x70: // BIT 6,B
			c.bit(6, c.registers.b)
		case 0x71: // BIT 6,C
			c.bit(6, c.registers.c)
		case 0x72: // BIT 6,D
			c.bit(6, c.registers.d)
		case 0x73: // BIT 6,E
			c.bit(6, c.registers.e)
		case 0x74: // BIT 6,H
			c.bit(6, c.registers.h)
		case 0x75: // BIT 6,L
			c.bit(6, c.registers.l)
		case 0x76: // BIT 6,(HL)
			c.bit(6, c.memory.Read(c.registers.getHL()))
		case 0x7F: // BIT 7,A
			c.bit(7, c.registers.a)
		case 0x78: // BIT 7,B
			c.bit(7, c.registers.b)
		case 0x79: // BIT 7,C
			c.bit(7, c.registers.c)
		case 0x7A: // BIT 7,D
			c.bit(7, c.registers.d)
		case 0x7B: // BIT 7,E
			c.bit(7, c.registers.e)
		case 0x7C: // BIT 7,H
			c.bit(7, c.registers.h)
		case 0x7D: // BIT 7,L
			c.bit(7, c.registers.l)
		case 0x7E: // BIT 7,(HL)
			c.bit(7, c.memory.Read(c.registers.getHL()))
		case 0x87: // res 0,A
			c.registers.a = c.res(0, c.registers.a)
		case 0x80: // res 0,B
			c.registers.b = c.res(0, c.registers.b)
		case 0x81: // res 0,C
			c.registers.c = c.res(0, c.registers.c)
		case 0x82: // res 0,D
			c.registers.d = c.res(0, c.registers.d)
		case 0x83: // res 0,E
			c.registers.e = c.res(0, c.registers.e)
		case 0x84: // res 0,H
			c.registers.h = c.res(0, c.registers.h)
		case 0x85: // res 0,L
			c.registers.l = c.res(0, c.registers.l)
		case 0x86: // res 0,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(0, v))
		case 0x8F: // res 1,A
			c.registers.a = c.res(1, c.registers.a)
		case 0x88: // res 1,B
			c.registers.b = c.res(1, c.registers.b)
		case 0x89: // res 1,C
			c.registers.c = c.res(1, c.registers.c)
		case 0x8A: // res 1,D
			c.registers.d = c.res(1, c.registers.d)
		case 0x8B: // res 1,E
			c.registers.e = c.res(1, c.registers.e)
		case 0x8C: // res 1,H
			c.registers.h = c.res(1, c.registers.h)
		case 0x8D: // res 1,L
			c.registers.l = c.res(1, c.registers.l)
		case 0x8E: // res 1,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(1, v))
		case 0x97: // res 2,A
			c.registers.a = c.res(2, c.registers.a)
		case 0x90: // res 2,B
			c.registers.b = c.res(2, c.registers.b)
		case 0x91: // res 2,C
			c.registers.c = c.res(2, c.registers.c)
		case 0x92: // res 2,D
			c.registers.d = c.res(2, c.registers.d)
		case 0x93: // res 2,E
			c.registers.e = c.res(2, c.registers.e)
		case 0x94: // res 2,H
			c.registers.h = c.res(2, c.registers.h)
		case 0x95: // res 2,L
			c.registers.l = c.res(2, c.registers.l)
		case 0x96: // res 2,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(2, v))
		case 0x9F: // res 3,A
			c.registers.a = c.res(3, c.registers.a)
		case 0x98: // res 3,B
			c.registers.b = c.res(3, c.registers.b)
		case 0x99: // res 3,C
			c.registers.c = c.res(3, c.registers.c)
		case 0x9A: // res 3,D
			c.registers.d = c.res(3, c.registers.d)
		case 0x9B: // res 3,E
			c.registers.e = c.res(3, c.registers.e)
		case 0x9C: // res 3,H
			c.registers.h = c.res(3, c.registers.h)
		case 0x9D: // res 3,L
			c.registers.l = c.res(3, c.registers.l)
		case 0x9E: // res 3,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(3, v))
		case 0xA7: // res 4,A
			c.registers.a = c.res(4, c.registers.a)
		case 0xA0: // res 4,B
			c.registers.b = c.res(4, c.registers.b)
		case 0xA1: // res 4,C
			c.registers.c = c.res(4, c.registers.c)
		case 0xA2: // res 4,D
			c.registers.d = c.res(4, c.registers.d)
		case 0xA3: // res 4,E
			c.registers.e = c.res(4, c.registers.e)
		case 0xA4: // res 4,H
			c.registers.h = c.res(4, c.registers.h)
		case 0xA5: // res 4,L
			c.registers.l = c.res(4, c.registers.l)
		case 0xA6: // res 4,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(4, v))
		case 0xAF: // res 5,A
			c.registers.a = c.res(5, c.registers.a)
		case 0xA8: // res 5,B
			c.registers.b = c.res(5, c.registers.b)
		case 0xA9: // res 5,C
			c.registers.c = c.res(5, c.registers.c)
		case 0xAA: // res 5,D
			c.registers.d = c.res(5, c.registers.d)
		case 0xAB: // res 5,E
			c.registers.e = c.res(5, c.registers.e)
		case 0xAC: // res 5,H
			c.registers.h = c.res(5, c.registers.h)
		case 0xAD: // res 5,L
			c.registers.l = c.res(5, c.registers.l)
		case 0xAE: // res 5,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(5, v))
		case 0xB7: // res 6,A
			c.registers.a = c.res(6, c.registers.a)
		case 0xB0: // res 6,B
			c.registers.b = c.res(6, c.registers.b)
		case 0xB1: // res 6,C
			c.registers.c = c.res(6, c.registers.c)
		case 0xB2: // res 6,D
			c.registers.d = c.res(6, c.registers.d)
		case 0xB3: // res 6,E
			c.registers.e = c.res(6, c.registers.e)
		case 0xB4: // res 6,H
			c.registers.h = c.res(6, c.registers.h)
		case 0xB5: // res 6,L
			c.registers.l = c.res(6, c.registers.l)
		case 0xB6: // res 6,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(6, v))
		case 0xBF: // res 7,A
			c.registers.a = c.res(7, c.registers.a)
		case 0xB8: // res 7,B
			c.registers.b = c.res(7, c.registers.b)
		case 0xB9: // res 7,C
			c.registers.c = c.res(7, c.registers.c)
		case 0xBA: // res 7,D
			c.registers.d = c.res(7, c.registers.d)
		case 0xBB: // res 7,E
			c.registers.e = c.res(7, c.registers.e)
		case 0xBC: // res 7,H
			c.registers.h = c.res(7, c.registers.h)
		case 0xBD: // res 7,L
			c.registers.l = c.res(7, c.registers.l)
		case 0xBE: // res 7,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.res(7, v))
		case 0xC7: // set 0,A
			c.registers.a = c.set(0, c.registers.a)
		case 0xC0: // set 0,B
			c.registers.b = c.set(0, c.registers.b)
		case 0xC1: // set 0,C
			c.registers.c = c.set(0, c.registers.c)
		case 0xC2: // set 0,D
			c.registers.d = c.set(0, c.registers.d)
		case 0xC3: // set 0,E
			c.registers.e = c.set(0, c.registers.e)
		case 0xC4: // set 0,H
			c.registers.h = c.set(0, c.registers.h)
		case 0xC5: // set 0,L
			c.registers.l = c.set(0, c.registers.l)
		case 0xC6: // set 0,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(0, v))
		case 0xCF: // set 1,A
			c.registers.a = c.set(1, c.registers.a)
		case 0xC8: // set 1,B
			c.registers.b = c.set(1, c.registers.b)
		case 0xC9: // set 1,C
			c.registers.c = c.set(1, c.registers.c)
		case 0xCA: // set 1,D
			c.registers.d = c.set(1, c.registers.d)
		case 0xCB: // set 1,E
			c.registers.e = c.set(1, c.registers.e)
		case 0xCC: // set 1,H
			c.registers.h = c.set(1, c.registers.h)
		case 0xCD: // set 1,L
			c.registers.l = c.set(1, c.registers.l)
		case 0xCE: // set 1,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(1, v))
		case 0xD7: // set 2,A
			c.registers.a = c.set(2, c.registers.a)
		case 0xD0: // set 2,B
			c.registers.b = c.set(2, c.registers.b)
		case 0xD1: // set 2,C
			c.registers.c = c.set(2, c.registers.c)
		case 0xD2: // set 2,D
			c.registers.d = c.set(2, c.registers.d)
		case 0xD3: // set 2,E
			c.registers.e = c.set(2, c.registers.e)
		case 0xD4: // set 2,H
			c.registers.h = c.set(2, c.registers.h)
		case 0xD5: // set 2,L
			c.registers.l = c.set(2, c.registers.l)
		case 0xD6: // set 2,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(2, v))
		case 0xDF: // set 3,A
			c.registers.a = c.set(3, c.registers.a)
		case 0xD8: // set 3,B
			c.registers.b = c.set(3, c.registers.b)
		case 0xD9: // set 3,C
			c.registers.c = c.set(3, c.registers.c)
		case 0xDA: // set 3,D
			c.registers.d = c.set(3, c.registers.d)
		case 0xDB: // set 3,E
			c.registers.e = c.set(3, c.registers.e)
		case 0xDC: // set 3,H
			c.registers.h = c.set(3, c.registers.h)
		case 0xDD: // set 3,L
			c.registers.l = c.set(3, c.registers.l)
		case 0xDE: // set 3,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(3, v))
		case 0xE7: // set 4,A
			c.registers.a = c.set(4, c.registers.a)
		case 0xE0: // set 4,B
			c.registers.b = c.set(4, c.registers.b)
		case 0xE1: // set 4,C
			c.registers.c = c.set(4, c.registers.c)
		case 0xE2: // set 4,D
			c.registers.d = c.set(4, c.registers.d)
		case 0xE3: // set 4,E
			c.registers.e = c.set(4, c.registers.e)
		case 0xE4: // set 4,H
			c.registers.h = c.set(4, c.registers.h)
		case 0xE5: // set 4,L
			c.registers.l = c.set(4, c.registers.l)
		case 0xE6: // set 4,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(4, v))
		case 0xEF: // set 5,A
			c.registers.a = c.set(5, c.registers.a)
		case 0xE8: // set 5,B
			c.registers.b = c.set(5, c.registers.b)
		case 0xE9: // set 5,C
			c.registers.c = c.set(5, c.registers.c)
		case 0xEA: // set 5,D
			c.registers.d = c.set(5, c.registers.d)
		case 0xEB: // set 5,E
			c.registers.e = c.set(5, c.registers.e)
		case 0xEC: // set 5,H
			c.registers.h = c.set(5, c.registers.h)
		case 0xED: // set 5,L
			c.registers.l = c.set(5, c.registers.l)
		case 0xEE: // set 5,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(5, v))
		case 0xF7: // set 6,A
			c.registers.a = c.set(6, c.registers.a)
		case 0xF0: // set 6,B
			c.registers.b = c.set(6, c.registers.b)
		case 0xF1: // set 6,C
			c.registers.c = c.set(6, c.registers.c)
		case 0xF2: // set 6,D
			c.registers.d = c.set(6, c.registers.d)
		case 0xF3: // set 6,E
			c.registers.e = c.set(6, c.registers.e)
		case 0xF4: // set 6,H
			c.registers.h = c.set(6, c.registers.h)
		case 0xF5: // set 6,L
			c.registers.l = c.set(6, c.registers.l)
		case 0xF6: // set 6,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(6, v))
		case 0xFF: // set 7,A
			c.registers.a = c.set(7, c.registers.a)
		case 0xF8: // set 7,B
			c.registers.b = c.set(7, c.registers.b)
		case 0xF9: // set 7,C
			c.registers.c = c.set(7, c.registers.c)
		case 0xFA: // set 7,D
			c.registers.d = c.set(7, c.registers.d)
		case 0xFB: // set 7,E
			c.registers.e = c.set(7, c.registers.e)
		case 0xFC: // set 7,H
			c.registers.h = c.set(7, c.registers.h)
		case 0xFD: // set 7,L
			c.registers.l = c.set(7, c.registers.l)
		case 0xFE: // set 7,(HL)
			v := c.memory.Read(c.registers.getHL())
			c.memory.Write(c.registers.getHL(), c.set(7, v))
		default:
			panic(fmt.Sprintf("unimplemented opcode: CB %#2x\n", op))
		}
	}

	return c.pc
}

// readNext reads the opcode at the program counter and increments the program counter
func (c *CPU) readNext() byte {
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
