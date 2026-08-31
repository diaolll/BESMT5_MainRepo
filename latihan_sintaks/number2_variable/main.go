package main

import "fmt"

func main() {
	var namaMhs string = "Diaul"
	var semester int = 4
	var ipk float64 = 3.75
	var statusAktif bool = true
	skill := []string{"UI/UX", "Figma", "Go"}

	fmt.Println("Nama    :", namaMhs)
	fmt.Println("Semester:", semester)
	fmt.Println("IPK     :", ipk)
	fmt.Println("Aktif   :", statusAktif)
	fmt.Println("Skill   :", skill)

	dataMahasiswa := make(map[string]float64)
	dataMahasiswa["Diaul"] = 92.0
	dataMahasiswa["Sinta"] = 87.5
	dataMahasiswa["Rafi"] = 75.0

	fmt.Println("\nData mahasiswa awal:", dataMahasiswa)

	nilai, ketemu := dataMahasiswa["Diaul"]
	if ketemu {
		fmt.Println("Nilai Diaul ditemukan:", nilai)
	}

	_, ketemuGak := dataMahasiswa["Andi"]
	if !ketemuGak {
		fmt.Println("Andi tidak ada di data")
	}

	delete(dataMahasiswa, "Rafi")
	fmt.Println("\nData setelah Rafi dihapus:")
	for nama, nilai := range dataMahasiswa {
		fmt.Printf("- %s: %.1f\n", nama, nilai)
	}
}