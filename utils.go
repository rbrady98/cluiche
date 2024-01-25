package main

func TestBit(val byte, bit int) bool {
	return (val>>bit)&0x1 == 1
}
