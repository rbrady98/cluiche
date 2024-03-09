package main

func TestBit(val byte, bit int) bool {
	return (val>>bit)&0x1 == 1
}

func ResetBit(val byte, bit byte) byte {
	return val & ^(1 << bit)
}

func SetBit(val byte, bit byte) byte {
	return val | (1 << bit)
}
