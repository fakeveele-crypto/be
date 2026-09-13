package service

import (
	"testing"

	"latihan-fiber/tm5/app/model"
)

func TestValidateRegister_PasswordTerlaluPendek(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "abc123",
	}
	errs := ValidateRegister(req)
	if _, ada := errs["password"]; !ada {
		t.Error("harusnya ada error password karena kurang dari 8 karakter")
	}
}

func TestValidateRegister_PasswordTanpaAngka(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "hanyahuruf",
	}
	errs := ValidateRegister(req)
	if _, ada := errs["password"]; !ada {
		t.Error("harusnya ada error password karena tidak ada angka")
	}
}

func TestValidateRegister_EmailTidakValid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "bukan-email",
		Password: "rahasia123",
	}
	errs := ValidateRegister(req)
	if _, ada := errs["email"]; !ada {
		t.Error("harusnya ada error email karena tidak mengandung @")
	}
}

func TestValidateRegister_DataValid(t *testing.T) {
	req := model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "rahasia123",
	}
	errs := ValidateRegister(req)
	if len(errs) != 0 {
		t.Errorf("harusnya tidak ada error, tapi dapat: %v", errs)
	}
}