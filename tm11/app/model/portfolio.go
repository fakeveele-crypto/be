package model

import "time"

// Portfolio adalah portofolio SATU mahasiswa. Data mahasiswanya tetap di
// PostgreSQL (tabel students); StudentID hanya merujuk ke students.id.
//
// Seperti Review, tidak ada tipe driver MongoDB di sini. Bentuk dokumen
// sebagaimana tersimpan didefinisikan di repository.
type Portfolio struct {
	ID           string        `json:"id"`
	StudentID    int           `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Technologies []string `json:"technologies"`
	URL          string   `json:"url"`
	Year         int      `json:"year"`
}

// Certificate.Date memakai teks YYYY-MM-DD di sisi API. Di MongoDB ia
// disimpan sebagai tipe date BSON; penerjemahannya tugas repository.
type Certificate struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

// MaxProjectsPerPortfolio dan MaxCertificatesPerPortfolio membatasi
// pertumbuhan array. Satu dokumen MongoDB maksimal 16 MB.
const (
	MaxProjectsPerPortfolio     = 20
	MaxCertificatesPerPortfolio = 30
)

type ProjectInput struct {
	Title        string   `json:"title" validate:"required,min=2,max=120"`
	Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=1,max=30"`
	URL          string   `json:"url" validate:"omitempty,url,max=300"`
	Year         int      `json:"year" validate:"required,min=1990,max=2100"`
}

type CertificateInput struct {
	Name   string `json:"name" validate:"required,min=2,max=120"`
	Issuer string `json:"issuer" validate:"required,min=2,max=120"`
	Date   string `json:"date" validate:"required,datetime=2006-01-02"`
}

// ReplacePortfolioRequest dipakai PUT: seluruh isi portofolio diganti.
type ReplacePortfolioRequest struct {
	Headline     string             `json:"headline" validate:"required,min=3,max=120"`
	Bio          string             `json:"bio" validate:"max=1000"`
	Projects     []ProjectInput     `json:"projects" validate:"max=20,dive"`
	Certificates []CertificateInput `json:"certificates" validate:"max=30,dive"`
}

type TechnologyCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}
