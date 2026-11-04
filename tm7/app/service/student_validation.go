package service

import "latihan-fiber/tm7/app/model"

// ApplyPatchStudent menggabungkan field yang dikirim pada PATCH ke data
// yang sudah ada. Pemeriksaan bentuk sudah dikerjakan tag sebelum fungsi
// ini dipanggil, jadi tugasnya tinggal satu: menggabungkan.
func ApplyPatchStudent(existing model.Student, patch model.PatchStudentRequest) model.Student {
	if patch.NIM != nil {
		existing.Nim = *patch.NIM
	}
	if patch.Name != nil {
		existing.Name = *patch.Name
	}
	if patch.Grade != nil {
		existing.Grade = *patch.Grade
	}
	if patch.IsActive != nil {
		existing.IsActive = *patch.IsActive
	}
	return existing
}

// IsEmptyPatchStudent memeriksa body PATCH yang tidak berisi field apa pun.
// Aturan ini tidak dapat ditulis sebagai tag, karena ia menyangkut
// hubungan antar field: setidaknya satu di antara mereka harus ada.
func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}