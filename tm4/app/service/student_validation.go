package service

import (
	"errors"
	"strings"

	"latihan-fiber/tm4/app/model"
)

// 1. Validasi untuk POST (Create)
func ValidateCreateStudent(req model.Student) error {
	if strings.TrimSpace(req.Nim) == "" {
		return errors.New("NIM tidak boleh kosong")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("nama tidak boleh kosong")
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		return errors.New("grade/IPK harus berada di rentang 0.0 - 4.0")
	}
	return nil
}

// 2. Validasi untuk PUT (Update)
func ValidateUpdateStudent(req model.Student) error {
	if strings.TrimSpace(req.Nim) == "" {
		return errors.New("NIM wajib diisi untuk pembaruan data")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("nama wajib diisi untuk pembaruan data")
	}
	if req.Grade < 0 || req.Grade > 4.0 {
		return errors.New("grade/IPK harus berada di rentang 0.0 - 4.0")
	}
	return nil
}

// 3. Penerapan Perubahan untuk PATCH (Partial Update)
func ApplyPatchStudent(existing model.Student, patch model.Student) (model.Student, error) {
	updated := existing

	if strings.TrimSpace(patch.Nim) != "" {
		updated.Nim = patch.Nim
	}
	if strings.TrimSpace(patch.Name) != "" {
		updated.Name = patch.Name
	}
	if patch.Grade >= 0 && patch.Grade <= 4.0 {
		updated.Grade = patch.Grade
	}

	if err := ValidateUpdateStudent(updated); err != nil {
		return existing, err
	}

	return updated, nil
}