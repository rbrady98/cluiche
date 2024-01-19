package main

// Perform ADD instruction
func (c *CPU) add(reg1 byte, reg2 byte) byte {
	total := int16(reg1) + int16(reg2)

	c.registers.flags.Zero = (byte(total) == 0)
	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = ((reg1&0xF)+(reg2&0xF) > 0xF)
	c.registers.flags.Carry = total > 0xFF

	return byte(total)
}

// Perform ADD instruction and set the carry flag
func (c *CPU) addC(reg1 byte, reg2 byte) byte {
	carry := int16(boolToByte(c.registers.flags.Carry))
	total := int16(reg1) + int16(reg2) + carry

	c.registers.flags.Zero = (byte(total) == 0)
	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = ((reg1&0xF)+(reg2&0xF) > 0xF)
	c.registers.flags.Carry = total > 0xFF

	return byte(total)
}

func (c *CPU) sub(reg1 byte, reg2 byte) byte {
	total := int16(reg1) - int16(reg2)

	c.registers.flags.Zero = (total == 0)
	c.registers.flags.Subtract = true
	c.registers.flags.HalfCarry = int16(reg1&0x0F)-int16(reg2&0xF) < 0
	c.registers.flags.Carry = total < 0

	return byte(total)
}

func (c *CPU) subC(reg1 byte, reg2 byte) byte {
	carry := int16(boolToByte(c.registers.flags.Carry))
	total := int16(reg1) - int16(reg2) - carry

	c.registers.flags.Zero = (total == 0)
	c.registers.flags.Subtract = true
	c.registers.flags.HalfCarry = int16(reg1&0x0F)-int16(reg2&0xF)-carry < 0
	c.registers.flags.Carry = total < 0

	return byte(total)
}

func (c *CPU) and(reg1 byte, reg2 byte) byte {
	total := reg1 & reg2

	c.registers.flags.Zero = (total == 0)
	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = true
	c.registers.flags.Carry = false

	return total
}

func (c *CPU) or(reg1 byte, reg2 byte) byte {
	total := reg1 | reg2

	c.registers.flags.Zero = (total == 0)
	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = false
	c.registers.flags.Carry = false

	return total
}

func (c *CPU) cp(reg1 byte, reg2 byte) {
	c.registers.flags.Zero = reg1 == reg2
	c.registers.flags.Subtract = true
	c.registers.flags.HalfCarry = false
	c.registers.flags.Carry = reg1 < reg2
}

func (c *CPU) inc(reg byte) byte {
	total := reg + 1

	c.registers.flags.Zero = total == 0
	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = ((reg&0xF)+1 > 0xF)

	return total
}

func (c *CPU) dec(reg byte) byte {
	total := reg - 1

	c.registers.flags.Zero = total == 0
	c.registers.flags.Subtract = true
	c.registers.flags.HalfCarry = reg&0x0F == 0

	return total
}

func (c *CPU) add16(reg1 uint16, reg2 uint16) uint16 {
	total := int32(reg1) + int32(reg2)

	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = int32(reg1&0xFFF) > (total & 0xFFF)
	c.registers.flags.Carry = total > 0xFFFF

	return uint16(total)
}

func (c *CPU) signedAdd16(reg1 uint16, reg2 uint16) uint16 {
	total := int32(reg1) + int32(reg2)

	c.registers.flags.Subtract = false
	c.registers.flags.HalfCarry = int32(reg1&0xFFF) > (total & 0xFFF)
	c.registers.flags.Carry = total > 0xFFFF

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

func (c *CPU) call() {
	next := c.readNext16()
	c.push(c.pc)
	c.jump(next)
}

func (c *CPU) ret() {
	c.jump(c.pop())
}

func (c *CPU) halt() {
	c.halted = true
}

func boolToByte(b bool) byte {
	if b {
		return byte(1)
	} else {
		return byte(0)
	}
}
