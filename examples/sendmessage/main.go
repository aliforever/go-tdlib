package main

func main() {
	err := NewClient("Examples").Start()

	panic(err)
}
