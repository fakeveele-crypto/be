package main

import "fmt"

func main() {

	var nama string = "Vale"
	var semester int = 4
	var ipk float64 = 3.85
	var isAktif bool = true
	var daftarNilai []int = []int{90, 85, 88}

	fmt.Println("Variabel Mahasiswa")
	fmt.Println("Nama Mahasiswa :", nama)
	fmt.Println("Semester       :", semester)
	fmt.Println("IPK            :", ipk)
	fmt.Println("Status Aktif   :", isAktif)
	fmt.Println("Daftar Nilai   :", daftarNilai)
	fmt.Println()

	dataMahasiswa := make(map[string]int)

	dataMahasiswa["Alisya"] = 85
	dataMahasiswa["Lailia"] = 92
	dataMahasiswa["Iqbal"] = 78

	fmt.Println("Menambahkan Data Mahasiswa")
	fmt.Println("Data Mahasiswa Saat Ini:")
	for kunci, isi := range dataMahasiswa {
		fmt.Println("- Nama:", kunci, "; Nilai:", isi)
	}
	fmt.Println()

	fmt.Println("Mencari Data Mahasiswa")
	namaCari := "Lailia"
	if nilai, ada := dataMahasiswa[namaCari]; ada {
		fmt.Println("Nilai", namaCari, "ditemukan, yaitu", nilai)
	} else {
		fmt.Println("Data", namaCari, "tidak ditemukan")
	}
	fmt.Println()

	fmt.Println("Menghapus Data Mahasiswa")
	delete(dataMahasiswa, "Iqbal")
	fmt.Println("Data Iqbal telah dihapus dari sistem")
	fmt.Println()

	fmt.Println("Data Seluruh Mahasiswa:")
	for kunci, isi := range dataMahasiswa {
		fmt.Println("- Nama:", kunci, "; Nilai:", isi)
	}
}