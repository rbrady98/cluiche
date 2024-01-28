package main

// Perform ADD instruction
func (c *CPU) add(reg1 byte, reg2 byte) byte {
	total := int16(reg1) + int16(reg2)

	c.registers.f.Zero = (byte(total) == 0)
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = ((reg1&0xF)+(reg2&0xF) > 0xF)
	c.registers.f.Carry = total > 0xFF

	return byte(total)
}

// Perform ADD instruction and set the carry flag
func (c *CPU) addC(reg1 byte, reg2 byte) byte {
	carry := int16(boolToByte(c.registers.f.Carry))
	total := int16(reg1) + int16(reg2) + carry

	c.registers.f.Zero = (byte(total) == 0)
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = ((reg1&0xF)+(reg2&0xF) > 0xF)
	c.registers.f.Carry = total > 0xFF

	return byte(total)
}

func (c *CPU) sub(reg1 byte, reg2 byte) byte {
	total := int16(reg1) - int16(reg2)

	c.registers.f.Zero = (total == 0)
	c.registers.f.Subtract = true
	c.registers.f.HalfCarry = int16(reg1&0x0F)-int16(reg2&0xF) < 0
	c.registers.f.Carry = total < 0

	return byte(total)
}

func (c *CPU) subC(reg1 byte, reg2 byte) byte {
	carry := int16(boolToByte(c.registers.f.Carry))
	total := int16(reg1) - int16(reg2) - carry

	c.registers.f.Zero = (total == 0)
	c.registers.f.Subtract = true
	c.registers.f.HalfCarry = int16(reg1&0x0F)-int16(reg2&0xF)-carry < 0
	c.registers.f.Carry = total < 0

	return byte(total)
}

func (c *CPU) and(reg1 byte, reg2 byte) byte {
	total := reg1 & reg2

	c.registers.f.Zero = (total == 0)
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = true
	c.registers.f.Carry = false

	return total
}

func (c *CPU) or(reg1 byte, reg2 byte) byte {
	total := reg1 | reg2

	c.registers.f.Zero = (total == 0)
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = false

	return total
}

func (c *CPU) xor(reg1 byte, reg2 byte) byte {
	r := reg1 ^ reg2

	c.registers.f.Zero = r == 0
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = false

	return r
}

func (c *CPU) cp(reg1 byte, reg2 byte) {
	c.registers.f.Zero = reg1 == reg2
	c.registers.f.Subtract = true
	c.registers.f.HalfCarry = (reg1 & 0x0f) > (reg2 & 0x0f)
	c.registers.f.Carry = reg1 > reg2
}

func (c *CPU) inc(reg byte) byte {
	total := reg + 1

	c.registers.f.Zero = total == 0
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = ((reg&0xF)+1 > 0xF)

	return total
}

func (c *CPU) inc16(reg uint16) uint16 {
	return reg + 1
}

func (c *CPU) dec(reg byte) byte {
	total := reg - 1

	c.registers.f.Zero = total == 0
	c.registers.f.Subtract = true
	c.registers.f.HalfCarry = reg&0x0F == 0

	return total
}

func (c *CPU) dec16(reg uint16) uint16 {
	return reg - 1
}

func (c *CPU) add16(reg1 uint16, reg2 uint16) uint16 {
	total := int32(reg1) + int32(reg2)

	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = int32(reg1&0xFFF) > (total & 0xFFF)
	c.registers.f.Carry = total > 0xFFFF

	return uint16(total)
}

func (c *CPU) signedAdd16(reg1 uint16, reg2 uint16) uint16 {
	total := int32(reg1) + int32(reg2)

	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = int32(reg1&0xFFF) > (total & 0xFFF)
	c.registers.f.Carry = total > 0xFFFF

	return uint16(total)
}

func (c *CPU) jump(next uint16) {
	c.pc = next
}

// push pushes to the stack
func (c *CPU) push(value uint16) {
	c.memory.Write(c.sp-1, byte(value&0xFF00>>8))
	c.memory.Write(c.sp-2, byte(value&0xFF))
	c.sp -= 2
}

// pop pops from the stack
func (c *CPU) pop() uint16 {
	b1 := uint16(c.memory.Read(c.sp))
	b2 := uint16(c.memory.Read(c.sp+1)) << 8
	c.sp += 2
	return b1 | b2
}

func (c *CPU) call(next uint16) {
	c.push(c.pc)
	c.jump(next)
}

func (c *CPU) rst(dest uint16) {
	c.push(c.pc)
	c.jump(0x00 + dest)
}

func (c *CPU) ret() {
	c.jump(c.pop())
}

func (c *CPU) halt() {
	c.halted = true
}

// rr rotates val right through carry flag
func (c *CPU) rr(val byte) byte {
	newC := val & 1
	oldC := boolToByte(c.registers.f.Carry)
	r := (val >> 1) | (oldC << 7)

	c.registers.f.Zero = r == 0
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = newC == 1

	return r
}

func (c *CPU) rra() byte {
	var carry byte
	if c.registers.f.Carry {
		carry = 0x80
	}

	r := c.registers.a>>1 | carry

	c.registers.f.Zero = false
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = (c.registers.a & 1) == 1

	return r
}

// srl shift val right into carry, returns 0 MSB val
func (c *CPU) srl(val byte) byte {
	r := val >> 1

	c.registers.f.Zero = r == 0
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = (val & 1) == 1

	return r
}

func (c *CPU) rlca() {
	v := c.registers.a
	res := (v << 1) | (v >> 7)

	c.registers.f.Zero = res == 0
	c.registers.f.Subtract = false
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = v > 0x7F

	c.registers.a = res
}

func (c *CPU) swap(val byte) byte {
	s := ((val & 0xF) << 4) | ((val & 0xF0) >> 4)

	c.registers.f.Zero = s == 0

	return s
}

func (c *CPU) daa() {
	a := c.registers.a
	half := c.registers.f.HalfCarry
	carry := c.registers.f.Carry
	sub := c.registers.f.Subtract

	if !sub {
		if carry || a > 0x99 {
			a += 0x60
			c.registers.f.Carry = true
		}

		if half || (a&0x0F > 0x09) {
			a += 0x6
		}
	} else {
		if carry {
			a -= 0x60
		}

		if half {
			a -= 0x06
		}
	}

	c.registers.f.Zero = a == 0
	c.registers.f.HalfCarry = false
	c.registers.f.Carry = a > 0x99

	c.registers.a = a
}

func boolToByte(b bool) byte {
	if b {
		return byte(1)
	} else {
		return byte(0)
	}
}
