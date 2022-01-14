package main

func main() {
	f(map[string]int{})
	f(map[string]int{})
}

func f(m map[string]int) {
	println(len(m))
}
