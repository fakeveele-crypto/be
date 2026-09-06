package service

import (
	"testing"

	"latihan-fiber/tm4/app/model"
)

// Test 1: Validasi POST (Create)
func TestValidateCreateStudent(t *testing.T) {
	// Kasus Valid
	validStudent := model.Student{Nim: "12345", Name: "Mahasiswa A", Grade: 3.75}
	if err := ValidateCreateStudent(validStudent); err != nil {
		t.Errorf("Diharapkan nil, tetapi mendapatkan error: %v", err)
	}

	// Kasus Invalid (NIM Kosong)
	invalidStudent := model.Student{Nim: "", Name: "Mahasiswa B", Grade: 3.50}
	if err := ValidateCreateStudent(invalidStudent); err == nil {
		t.Errorf("Diharapkan error NIM kosong, tetapi mendapatkan nil")
	}
}

// Test 2: Validasi PUT (Update)
func TestValidateUpdateStudent(t *testing.T) {
	// Kasus Invalid Grade (> 4.0)
	invalidGrade := model.Student{Nim: "12345", Name: "Mahasiswa A", Grade: 4.50}
	if err := ValidateUpdateStudent(invalidGrade); err == nil {
		t.Errorf("Diharapkan error Grade > 4.0, tetapi mendapatkan nil")
	}
}

// Test 3: Penerapan PATCH
func TestApplyPatchStudent(t *testing.T) {
	existing := model.Student{Nim: "12345", Name: "Nama Lama", Grade: 3.00}
	patch := model.Student{Name: "Nama Baru"}

	updated, err := ApplyPatchStudent(existing, patch)
	if err != nil {
		t.Fatalf("Patch gagal dengan error: %v", err)
	}

	if updated.Name != "Nama Baru" {
		t.Errorf("Diharapkan Nama 'Nama Baru', tetapi mendapatkan '%s'", updated.Name)
	}
	if updated.Nim != "12345" {
		t.Errorf("NIM tidak boleh berubah jika tidak diisi pada payload patch")
	}
}