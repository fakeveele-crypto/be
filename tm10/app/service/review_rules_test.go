package service

import (
	"reflect"
	"testing"

	"latihan-fiber/tm10/app/model"
)

func TestNormalizeTags(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil menjadi slice kosong", nil, []string{}},
		{"huruf kecil dan spasi tepi", []string{"Tugas-Banyak", " dosen-jelas "}, []string{"tugas-banyak", "dosen-jelas"}},
		{"duplikat dibuang, urutan pertama dipertahankan",
			[]string{"Tugas-Banyak", " dosenjelas ", "tugas-banyak"}, []string{"tugas-banyak", "dosenjelas"}},
		{"elemen kosong dibuang", []string{"  ", "", "seru"}, []string{"seru"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := NormalizeTags(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("NormalizeTags(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeProject(t *testing.T) {
	p := model.ProjectInput{
		Title:        "  Sistem Ulasan ",
		URL:          " https://github.com/x/y ",
		Technologies: []string{"Go", " go ", "MongoDB", ""},
	}
	normalizeProject(&p)

	if p.Title != "Sistem Ulasan" || p.URL != "https://github.com/x/y" {
		t.Fatalf("title/url belum dirapikan: %+v", p)
	}
	if want := []string{"go", "mongodb"}; !reflect.DeepEqual(p.Technologies, want) {
		t.Fatalf("technologies = %q, want %q", p.Technologies, want)
	}
}
