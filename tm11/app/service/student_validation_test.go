package service

import (
	"testing"

	"latihan-fiber/tm11/app/model"
	"latihan-fiber/tm11/helper"
)

func TestApplyPatchStudent(t *testing.T) {
	existing := model.Student{Nim: "434241088", Name: "Nama Lama", Grade: 3.00}
	name := "Nama Baru"

	updated := ApplyPatchStudent(existing, model.PatchStudentRequest{Name: &name})

	if updated.Name != "Nama Baru" {
		t.Errorf("nama seharusnya berubah, dapat %q", updated.Name)
	}
	if updated.Nim != "434241088" {
		t.Errorf("NIM tidak boleh berubah bila tidak dikirim")
	}
}

func TestIsEmptyPatchStudent(t *testing.T) {
	if !IsEmptyPatchStudent(model.PatchStudentRequest{}) {
		t.Error("patch tanpa field harus dianggap kosong")
	}
	name := "x"
	if IsEmptyPatchStudent(model.PatchStudentRequest{Name: &name}) {
		t.Error("patch dengan satu field tidak boleh dianggap kosong")
	}
}

func TestPatchStudentTags(t *testing.T) {
	empty := ""
	if errs := helper.ValidateStruct(model.PatchStudentRequest{Name: &empty}); errs["name"] == "" {
		t.Error("name kosong harus ditolak")
	}
	if errs := helper.ValidateStruct(model.PatchStudentRequest{}); errs != nil {
		t.Errorf("body tanpa name harus lolos, dapat %v", errs)
	}
}

func TestCreateStudentTags(t *testing.T) {
	ok := model.CreateStudentRequest{NIM: "434241088", Name: "Valerina", Grade: 3.5}
	if errs := helper.ValidateStruct(ok); errs != nil {
		t.Errorf("data sah ditolak: %v", errs)
	}
	badNIM := model.CreateStudentRequest{NIM: "12345", Name: "Valerina", Grade: 3.5}
	if errs := helper.ValidateStruct(badNIM); errs["nim"] == "" {
		t.Error("NIM 5 digit harus ditolak")
	}
	badGrade := model.CreateStudentRequest{NIM: "434241088", Name: "Valerina", Grade: 4.5}
	if errs := helper.ValidateStruct(badGrade); errs["grade"] == "" {
		t.Error("grade 4.5 harus ditolak")
	}
}
