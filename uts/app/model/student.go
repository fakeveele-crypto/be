package model

// Student adalah data mahasiswa.
// UserID dipakai untuk cek kepemilikan dan tidak ikut dikirim ke client.
type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"`
	Nama     string  `json:"nama"`
	Prodi    string  `json:"prodi"`
	Angkatan int     `json:"angkatan"`
	IPK      float64 `json:"ipk_terakhir"`
	UserID   int     `json:"-"`
}

// CreateStudentRequest adalah body POST /students.
// IPK berupa pointer supaya bisa dibedakan antara "tidak diisi" dan "diisi 0".
type CreateStudentRequest struct {
	NIM      string   `json:"nim"`
	Nama     string   `json:"nama"`
	Email    string   `json:"email"`
	Prodi    string   `json:"prodi"`
	Angkatan int      `json:"angkatan"`
	IPK      *float64 `json:"ipk_terakhir"`
}

// UpdateStudentRequest adalah body PUT /students/:id (NIM tidak boleh diubah).
type UpdateStudentRequest struct {
	Nama     string   `json:"nama"`
	Prodi    string   `json:"prodi"`
	Angkatan int      `json:"angkatan"`
	IPK      *float64 `json:"ipk_terakhir"`
}

// StudentFilter menampung parameter query untuk daftar mahasiswa.
type StudentFilter struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan int // 0 berarti semua angkatan
	Search   string
	Sort     string
}

// EnrolledCourse adalah satu mata kuliah yang diambil mahasiswa.
type EnrolledCourse struct {
	EnrollmentID  int    `json:"enrollment_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	TahunAkademik string `json:"tahun_akademik"`
}

// StudentDetail adalah isi response GET /students/:id.
type StudentDetail struct {
	Student
	MataKuliah []EnrolledCourse `json:"mata_kuliah"`
	TotalSKS   int              `json:"total_sks"`
	BatasSKS   int              `json:"batas_sks"`
}