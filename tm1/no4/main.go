package main

import "fmt"

type Student struct {
	ID       string
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	status := "Tidak Aktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("ID: %s, Nama: %s, Nilai: %.2f, Status: %s", s.ID, s.Name, s.Grade, status)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	mhs := Student{
		ID:       "088",
		Name:     "Vale",
		Grade:    3.5,
		IsActive: false,
	}

	fmt.Println("Mengembalikan informasi lengkap student dalam satu baris teks")
	fmt.Println(mhs.GetInfo())

	fmt.Println("\nMemperbarui nilai")
	mhs.UpdateGrade(3.90)
	fmt.Println(mhs.GetInfo()) 

	fmt.Println("\nMengubah status aktif")

	mhs.Activate()
	fmt.Println(mhs.GetInfo())

	mhs.Deactivate()
	fmt.Println(mhs.GetInfo())
}