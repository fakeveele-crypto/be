package helper

import (
	"strings"
	"testing"

	"latihan-fiber/tm11/app/model"
)

// Batas jumlah item pada slice harus dilaporkan sebagai "item", bukan
// "nilai maksimal" (kekeliruan yang ditemukan saat modul 10 diuji).
func TestValidateStruct_SliceLimitsSaySItem(t *testing.T) {
	errs := ValidateStruct(model.CreateReviewRequest{
		Rating: 4, Comment: "komentar yang cukup panjang",
		Tags: []string{"aa", "bb", "cc", "dd", "ee", "ff"},
	})
	if got := errs["tags"]; got != "maksimal 5 item" {
		t.Fatalf("pesan tags = %q, want %q", got, "maksimal 5 item")
	}
}

func TestValidateStruct_PointerSliceUpdateRequest(t *testing.T) {
	tags := []string{"aa", "bb", "cc", "dd", "ee", "ff"}
	errs := ValidateStruct(model.UpdateReviewRequest{Tags: &tags})
	if got := errs["tags"]; got != "maksimal 5 item" {
		t.Fatalf("pesan tags (pointer) = %q, want %q", got, "maksimal 5 item")
	}
}

// Pada data bersarang, kunci error memuat jalur lengkapnya sehingga jelas
// proyek ke berapa yang bermasalah.
func TestValidateStruct_NestedFieldPath(t *testing.T) {
	errs := ValidateStruct(model.ReplacePortfolioRequest{
		Headline: "Backend dev",
		Projects: []model.ProjectInput{
			{Title: "Proyek valid", Technologies: []string{"go"}, Year: 2026},
			{Title: "a", Technologies: []string{}, URL: "bukan-url", Year: 1800},
		},
		Certificates: []model.CertificateInput{{Name: "Sertifikat", Issuer: "Dicoding", Date: "20-05-2026"}},
	})
	for _, key := range []string{
		"projects[1].title", "projects[1].technologies", "projects[1].url",
		"projects[1].year", "certificates[0].date",
	} {
		if _, ok := errs[key]; !ok {
			t.Errorf("kunci %q tidak ada pada hasil: %v", key, errs)
		}
	}
	if _, ok := errs["projects[0].title"]; ok {
		t.Errorf("proyek pertama valid, tidak boleh ada errornya: %v", errs)
	}
	if !strings.Contains(errs["certificates[0].date"], "YYYY-MM-DD") {
		t.Errorf("pesan tanggal = %q", errs["certificates[0].date"])
	}
}

// Struct datar tetap memakai nama field saja (response endpoint lama
// tidak boleh berubah).
func TestValidateStruct_FlatKeysUnchanged(t *testing.T) {
	errs := ValidateStruct(model.CreateStudentRequest{NIM: "1", Name: ""})
	if _, ok := errs["nim"]; !ok {
		t.Fatalf("kunci nim hilang: %v", errs)
	}
	if _, ok := errs["name"]; !ok {
		t.Fatalf("kunci name hilang: %v", errs)
	}
}
