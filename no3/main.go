package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func ubahNilai(x int) {
	x = 100
}

func ubahLewatPointer(x *int) {
	*x = 100
}

func main() {
	x, y := 5, 10
	fmt.Println("sebelum swap:", x, y)
	swap(&x, &y)
	fmt.Println("sesudah swap:", x, y)

	daftar := []string{"apel", "jeruk"}
	fmt.Println("sebelum update:", daftar)
	updateSlice(&daftar, "mangga")
	fmt.Println("sesudah update:", daftar)

	num := 42
	ubahNilai(num)
	fmt.Println("setelah mengubah nilai (value):", num) 

	ubahLewatPointer(&num)
	fmt.Println("setelah mengubah pointer:", num) 
}