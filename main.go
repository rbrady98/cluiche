package main

func main() {
	gb, err := NewGameboy("./roms/cpu_instrs.gb")
	if err != nil {
		panic(err.Error())
	}

	gb.Run()
}
