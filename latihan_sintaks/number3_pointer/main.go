package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func swapValue(a, b int) {
	a, b = b, a
}

func tambahItem(daftar *[]string, item string) {
	*daftar = append(*daftar, item)
}

func main() {
	p, q := 5, 15
	fmt.Println("sebelum swapValue:", p, q)
	swapValue(p, q)
	fmt.Println("sesudah swapValue:", p, q, "(tetap sama, cuma dapat salinan)")

	m, n := 5, 15
	fmt.Println("\nsebelum swap:", m, n)
	swap(&m, &n)
	fmt.Println("sesudah swap:", m, n, "(berubah, dapat alamat asli)")

	daftarTugas := []string{"wireframe", "prototype"}
	fmt.Println("\nsebelum tambahItem:", daftarTugas)
	tambahItem(&daftarTugas, "usability test")
	fmt.Println("sesudah tambahItem:", daftarTugas)
}