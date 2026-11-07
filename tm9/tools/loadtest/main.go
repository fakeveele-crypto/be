// Alat uji beban Modul 9. BUKAN bagian aplikasi: tidak ikut di-deploy.
//
// Pemakaian (server harus sedang berjalan di terminal lain, dari root repo):
//
//	go run ./tm9/tools/loadtest -mode crowd        # 60 mahasiswa berebut IF399 (kuota 5)
//	go run ./tm9/tools/loadtest -mode credits      # 10 mahasiswa x 12 kelas (batas 24 SKS)
//	go run ./tm9/tools/loadtest -mode swap         # 20 pasang pindah kelas berlawanan arah
//	go run ./tm9/tools/loadtest -mode sv-assign    # D.3: 30 penetapan ke pembimbing berkuota 3
//	go run ./tm9/tools/loadtest -mode sv-switch    # D.2: 20 pasang ganti pembimbing berlawanan arah
//	go run ./tm9/tools/loadtest -mode token -user mhs001   # cetak token untuk dipakai di Postman
//	go run ./tm9/tools/loadtest -mode token -user loadadmin
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"latihan-fiber/tm9/app/model"
	"latihan-fiber/tm9/config"
	"latihan-fiber/tm9/database"
	"latihan-fiber/tm9/helper"
)

type runner struct {
	ctx    context.Context
	pool   *pgxpool.Pool
	jwt    *helper.JWTManager
	base   string
	client *http.Client
}

func main() {
	mode := flag.String("mode", "crowd", "crowd | credits | swap | sv-assign | sv-switch")
	base := flag.String("base", "http://localhost:3000", "alamat dasar server")
	user := flag.String("user", "mhs001", "username untuk mode token")
	flag.Parse()

	config.LoadEnv()
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	secret := config.GetEnv("JWT_SECRET", "")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET tidak diisi atau terlalu pendek")
	}
	r := &runner{
		ctx:  ctx,
		pool: pool,
		jwt: helper.NewJWTManager(secret, config.GetEnv("JWT_ISSUER", "praktikum-backend"),
			time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute),
		base: strings.TrimRight(*base, "/"),
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        256,
				MaxIdleConnsPerHost: 256,
			},
		},
	}

	switch *mode {
	case "crowd":
		err = r.crowd()
	case "credits":
		err = r.credits()
	case "swap":
		err = r.swap()
	case "sv-assign":
		err = r.svAssign()
	case "sv-switch":
		err = r.svSwitch()
	case "token":
		err = r.printToken(*user)
	default:
		err = fmt.Errorf("mode tidak dikenal: %s", *mode)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// fire menjalankan n pekerjaan secara serentak dan menghitung hasilnya.
//
// Setiap goroutine menunggu di depan "gerbang" start yang sama. Gerbang
// baru dibuka setelah SELURUH goroutine siap, sehingga request benar-benar
// berangkat hampir bersamaan, bukan bergiliran sesuai urutan dibuatnya.
func fire(n int, job func(i int) string) map[string]int {
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, out chan<- string) {
			defer wg.Done()
			<-start
			out <- job(i)
		}(i, results)
	}

	close(start) // buka gerbang
	wg.Wait()    // tunggu SEMUA goroutine selesai
	close(results)

	tally := map[string]int{}
	for r := range results {
		tally[r]++
	}
	return tally
}

// ---------- pembantu HTTP ----------

func (r *runner) token(u model.User) (string, error) {
	// Alat ini mencetak token sendiri karena ia memegang JWT_SECRET.
	// Itulah sebabnya JWT_SECRET tidak boleh bocor: siapa pun yang
	// memegangnya dapat mencetak token atas nama siapa saja.
	return r.jwt.GenerateAccess(u)
}

// call mengirim satu request dan mengembalikan label hasil, misalnya
// "201 OK" atau "409 COURSE_FULL".
func (r *runner) call(method, path, token string, body any) string {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, r.base+path, rd)
	if err != nil {
		return "ERR request"
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// RequireJSON mewajibkan Content-Type JSON pada POST/PUT/PATCH,
	// termasuk POST tanpa body seperti /courses/:id/enroll.
	if method != "GET" && method != "DELETE" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "ERR " + shorten(err.Error())
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var parsed struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(data, &parsed)
	label := "OK"
	if !parsed.Success && resp.StatusCode >= 300 {
		label = parsed.Code
		if label == "" {
			label = "(tanpa kode)"
		}
	}
	return strconv.Itoa(resp.StatusCode) + " " + label
}

func shorten(s string) string {
	if i := strings.Index(s, ": "); i > 0 && len(s) > 60 {
		s = s[i+2:]
	}
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

func printTally(t map[string]int) {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println("hasil request:")
	for _, k := range keys {
		fmt.Printf("  %-34s %d\n", k, t[k])
	}
}

// ---------- pembantu database ----------

func (r *runner) loadStudents(n int) ([]model.User, error) {
	rows, err := r.pool.Query(r.ctx,
		`SELECT id, username, email, role FROM users
		  WHERE username LIKE 'mhs%' ORDER BY username LIMIT $1`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if len(users) < n {
		return nil, fmt.Errorf("hanya ada %d akun mhs%%, butuh %d (jalankan migration 008)", len(users), n)
	}
	return users, rows.Err()
}

func (r *runner) courseID(code string) (int, error) {
	var id int
	err := r.pool.QueryRow(r.ctx, "SELECT id FROM courses WHERE code = $1", code).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("kelas %s tidak ada: %w", code, err)
	}
	return id, nil
}

func (r *runner) deadlocks() int {
	var n int
	_ = r.pool.QueryRow(r.ctx,
		"SELECT deadlocks FROM pg_stat_database WHERE datname = current_database()").Scan(&n)
	return n
}

// resetMahasiswa sama dengan reset.sql.
func (r *runner) resetMahasiswa() error {
	if _, err := r.pool.Exec(r.ctx,
		`DELETE FROM enrollments WHERE user_id IN (SELECT id FROM users WHERE username LIKE 'mhs%')`); err != nil {
		return err
	}
	_, err := r.pool.Exec(r.ctx,
		`UPDATE courses c SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)`)
	return err
}

// reportCourses mencetak kesesuaian enrolled_count dengan jumlah baris
// enrollments, dan mengembalikan true bila semuanya sama.
func (r *runner) reportCourses(codes ...string) bool {
	ok := true
	for _, code := range codes {
		var quota, count, rows int
		err := r.pool.QueryRow(r.ctx,
			`SELECT quota, enrolled_count,
			        (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)
			   FROM courses c WHERE code = $1`, code).Scan(&quota, &count, &rows)
		if err != nil {
			fmt.Printf("%s: gagal membaca (%v)\n", code, err)
			return false
		}
		status := "KONSISTEN"
		if count != rows || count > quota {
			status = "TIDAK KONSISTEN"
			ok = false
		}
		fmt.Printf("%s kuota=%d enrolled_count=%d baris enrollments=%d -> %s\n",
			code, quota, count, rows, status)
	}
	return ok
}

// ---------- mode: token ----------

// printToken mencetak access token milik satu akun, untuk dipakai di
// Postman atau curl. Untuk akun admin, jalankan dulu mode sv-assign
// sekali (ia membuat akun loadadmin), lalu: -mode token -user loadadmin
func (r *runner) printToken(username string) error {
	var u model.User
	err := r.pool.QueryRow(r.ctx,
		`SELECT id, username, email, role FROM users WHERE LOWER(username) = LOWER($1)`,
		username).Scan(&u.ID, &u.Username, &u.Email, &u.Role)
	if err != nil {
		return fmt.Errorf("akun %s tidak ditemukan: %w", username, err)
	}
	t, err := r.token(u)
	if err != nil {
		return err
	}
	fmt.Println(t)
	return nil
}

// ---------- mode: crowd ----------

func (r *runner) crowd() error {
	users, err := r.loadStudents(60)
	if err != nil {
		return err
	}
	id, err := r.courseID("IF399")
	if err != nil {
		return err
	}
	tokens, err := r.tokens(users)
	if err != nil {
		return err
	}

	fmt.Println("60 mahasiswa berebut kelas IF399 secara serentak...")
	begin := time.Now()
	tally := fire(60, func(i int) string {
		return r.call("POST", fmt.Sprintf("/api/v1/courses/%d/enroll", id), tokens[i], nil)
	})
	elapsed := time.Since(begin)

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
	r.reportCourses("IF399")
	return nil
}

func (r *runner) tokens(users []model.User) ([]string, error) {
	out := make([]string, len(users))
	for i, u := range users {
		t, err := r.token(u)
		if err != nil {
			return nil, err
		}
		out[i] = t
	}
	return out, nil
}

// ---------- mode: credits ----------

var creditCourses = []string{
	"IF101", "IF102", "IF301", "IF302", "IF303", "IF304",
	"IF305", "IF306", "IF307", "IF308", "IF309", "IF310",
}

func (r *runner) credits() error {
	const students = 10
	users, err := r.loadStudents(students)
	if err != nil {
		return err
	}
	tokens, err := r.tokens(users)
	if err != nil {
		return err
	}
	ids := make([]int, len(creditCourses))
	for i, code := range creditCourses {
		if ids[i], err = r.courseID(code); err != nil {
			return err
		}
	}

	total := students * len(ids)
	fmt.Printf("%d mahasiswa, masing-masing mendaftar ke %d kelas 3 SKS secara serentak (batas 24 SKS)...\n",
		students, len(ids))
	begin := time.Now()
	tally := fire(total, func(i int) string {
		s, c := i/len(ids), i%len(ids)
		return r.call("POST", fmt.Sprintf("/api/v1/courses/%d/enroll", ids[c]), tokens[s], nil)
	})
	elapsed := time.Since(begin)

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())

	ok := true
	for _, u := range users {
		var sks int
		if err := r.pool.QueryRow(r.ctx,
			`SELECT COALESCE(SUM(c.credits),0) FROM enrollments e
			   JOIN courses c ON c.id = e.course_id WHERE e.user_id = $1`, u.ID).Scan(&sks); err != nil {
			return err
		}
		if sks > 24 {
			ok = false
			fmt.Printf("  %s mengambil %d SKS (MELEBIHI 24)\n", u.Username, sks)
		}
	}
	// enrolled_count harus selalu sama dengan jumlah baris enrollments.
	var bad int
	_ = r.pool.QueryRow(r.ctx,
		`SELECT COUNT(*) FROM courses c WHERE enrolled_count <>
		   (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)`).Scan(&bad)
	if bad > 0 {
		ok = false
		fmt.Printf("  %d kelas enrolled_count-nya tidak sama dengan jumlah baris\n", bad)
	}
	if ok {
		fmt.Printf("seluruh %d mahasiswa <= 24 SKS -> KONSISTEN\n", students)
	} else {
		fmt.Println("-> TIDAK KONSISTEN")
	}
	return nil
}

// ---------- mode: swap ----------

func (r *runner) swap() error {
	const pairs = 20
	users, err := r.loadStudents(pairs * 2)
	if err != nil {
		return err
	}
	tokens, err := r.tokens(users)
	if err != nil {
		return err
	}
	a, err := r.courseID("IF101")
	if err != nil {
		return err
	}
	b, err := r.courseID("IF102")
	if err != nil {
		return err
	}

	// Keadaan awal: mhs001..020 di IF101, mhs021..040 di IF102.
	if err := r.resetMahasiswa(); err != nil {
		return err
	}
	for i, u := range users {
		course := a
		if i >= pairs {
			course = b
		}
		if _, err := r.pool.Exec(r.ctx,
			"INSERT INTO enrollments (course_id, user_id) VALUES ($1, $2)", course, u.ID); err != nil {
			return err
		}
	}
	if _, err := r.pool.Exec(r.ctx,
		`UPDATE courses c SET enrolled_count = (SELECT COUNT(*) FROM enrollments e WHERE e.course_id = c.id)`); err != nil {
		return err
	}

	fmt.Printf("%d pasang mahasiswa pindah kelas berlawanan arah secara serentak...\n", pairs)
	before := r.deadlocks()
	begin := time.Now()
	tally := fire(pairs*2, func(i int) string {
		from, to := a, b
		if i >= pairs {
			from, to = b, a
		}
		return r.call("POST", "/api/v1/enrollments/swap", tokens[i],
			map[string]int{"from_course_id": from, "to_course_id": to})
	})
	elapsed := time.Since(begin)

	time.Sleep(1500 * time.Millisecond) // statistik PostgreSQL diperbarui dengan jeda
	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", r.deadlocks()-before)
	r.reportCourses("IF101", "IF102")
	return nil
}

// ---------- D.2 / D.3: pembimbing ----------

func (r *runner) adminToken() (string, error) {
	var u model.User
	err := r.pool.QueryRow(r.ctx,
		`INSERT INTO users (username, email, password, role)
		 VALUES ('loadadmin', 'loadadmin@student.unair.ac.id', '!', 'admin')
		 ON CONFLICT DO NOTHING RETURNING id, username, email, role`).
		Scan(&u.ID, &u.Username, &u.Email, &u.Role)
	if err != nil { // sudah ada
		err = r.pool.QueryRow(r.ctx,
			`SELECT id, username, email, role FROM users WHERE LOWER(username) = 'loadadmin'`).
			Scan(&u.ID, &u.Username, &u.Email, &u.Role)
		if err != nil {
			return "", err
		}
	}
	if u.Role != "admin" {
		return "", fmt.Errorf("akun loadadmin bukan admin")
	}
	return r.token(u)
}

// testStudents memastikan ada n baris students uji (NIM 990000001...).
func (r *runner) testStudents(n int) ([]int, error) {
	var owner int
	if err := r.pool.QueryRow(r.ctx,
		"SELECT id FROM users WHERE username = 'mhs001'").Scan(&owner); err != nil {
		return nil, fmt.Errorf("akun mhs001 tidak ada (migration 008): %w", err)
	}
	for i := 1; i <= n; i++ {
		if _, err := r.pool.Exec(r.ctx,
			`INSERT INTO students (nim, name, grade, owner_id)
			 SELECT $1::varchar, $2::varchar, 3.00, $3::int
			  WHERE NOT EXISTS (SELECT 1 FROM students WHERE nim = $1::varchar)`,
			fmt.Sprintf("99%07d", i), fmt.Sprintf("Mahasiswa Uji %03d", i), owner); err != nil {
			return nil, err
		}
	}
	rows, err := r.pool.Query(r.ctx,
		"SELECT id FROM students WHERE nim LIKE '99%' ORDER BY nim LIMIT $1", n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// clearSupervisions mengosongkan pembimbing milik mahasiswa uji, lalu
// menghitung ulang supervised_count dari baris yang sebenarnya.
func (r *runner) clearSupervisions(studentIDs []int) error {
	if _, err := r.pool.Exec(r.ctx,
		"DELETE FROM supervisions WHERE student_id = ANY($1)", studentIDs); err != nil {
		return err
	}
	_, err := r.pool.Exec(r.ctx,
		`UPDATE supervisors sv SET supervised_count =
		   (SELECT COUNT(*) FROM supervisions s WHERE s.supervisor_id = sv.id)`)
	return err
}

func (r *runner) newSupervisor(name string, quota int) (int, error) {
	if _, err := r.pool.Exec(r.ctx, "DELETE FROM supervisors WHERE name = $1", name); err != nil {
		return 0, err
	}
	var id int
	err := r.pool.QueryRow(r.ctx,
		"INSERT INTO supervisors (name, quota) VALUES ($1, $2) RETURNING id", name, quota).Scan(&id)
	return id, err
}

func (r *runner) reportSupervisor(id int) (count, rows, quota int) {
	_ = r.pool.QueryRow(r.ctx,
		`SELECT quota, supervised_count,
		        (SELECT COUNT(*) FROM supervisions s WHERE s.supervisor_id = sv.id)
		   FROM supervisors sv WHERE id = $1`, id).Scan(&quota, &count, &rows)
	status := "KONSISTEN"
	if count != rows || count > quota {
		status = "TIDAK KONSISTEN"
	}
	fmt.Printf("pembimbing id=%d kuota=%d supervised_count=%d baris supervisions=%d -> %s\n",
		id, quota, count, rows, status)
	return
}

func (r *runner) svAssign() error {
	const students, quota = 30, 3
	admin, err := r.adminToken()
	if err != nil {
		return err
	}
	ids, err := r.testStudents(students)
	if err != nil {
		return err
	}
	if err := r.clearSupervisions(ids); err != nil {
		return err
	}
	sv, err := r.newSupervisor("Dosen Uji Beban", quota)
	if err != nil {
		return err
	}

	fmt.Printf("%d penetapan serentak ke pembimbing berkuota %d...\n", students, quota)
	begin := time.Now()
	tally := fire(students, func(i int) string {
		return r.call("POST", fmt.Sprintf("/api/v1/supervisors/%d/assign", sv), admin,
			map[string]int{"student_id": ids[i]})
	})
	elapsed := time.Since(begin)

	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
	count, rows, _ := r.reportSupervisor(sv)
	// LULUS hanya bila hasilnya PERSIS: quota yang berhasil, sisanya
	// SUPERVISOR_FULL, dan tidak ada hasil jenis lain. Kegagalan lain
	// (misalnya 500 karena CHECK constraint) berarti kode salah, walau
	// database berhasil menahan kelebihannya.
	if tally["201 OK"] == quota && tally["409 SUPERVISOR_FULL"] == students-quota &&
		len(tally) == 2 && count == quota && rows == quota {
		fmt.Printf("tepat %d berhasil, %d SUPERVISOR_FULL -> LULUS\n", quota, students-quota)
	} else {
		fmt.Printf("hasil menyimpang dari yang diharapkan (%d berhasil + %d SUPERVISOR_FULL) -> GAGAL\n", quota, students-quota)
	}
	return nil
}

func (r *runner) svSwitch() error {
	const pairs = 20
	admin, err := r.adminToken()
	if err != nil {
		return err
	}
	ids, err := r.testStudents(pairs * 2)
	if err != nil {
		return err
	}
	if err := r.clearSupervisions(ids); err != nil {
		return err
	}
	a, err := r.newSupervisor("Dosen Uji Tukar A", pairs*2)
	if err != nil {
		return err
	}
	b, err := r.newSupervisor("Dosen Uji Tukar B", pairs*2)
	if err != nil {
		return err
	}
	for i, sid := range ids {
		sv := a
		if i >= pairs {
			sv = b
		}
		if _, err := r.pool.Exec(r.ctx,
			"INSERT INTO supervisions (student_id, supervisor_id) VALUES ($1, $2) ON CONFLICT (student_id) DO UPDATE SET supervisor_id = EXCLUDED.supervisor_id",
			sid, sv); err != nil {
			return err
		}
	}
	if _, err := r.pool.Exec(r.ctx,
		`UPDATE supervisors sv SET supervised_count =
		   (SELECT COUNT(*) FROM supervisions s WHERE s.supervisor_id = sv.id) WHERE id = ANY($1)`,
		[]int{a, b}); err != nil {
		return err
	}

	fmt.Printf("%d pasang mahasiswa ganti pembimbing berlawanan arah secara serentak...\n", pairs)
	before := r.deadlocks()
	begin := time.Now()
	tally := fire(pairs*2, func(i int) string {
		to := b
		if i >= pairs {
			to = a
		}
		return r.call("POST", "/api/v1/supervisions/switch", admin,
			map[string]int{"student_id": ids[i], "to_supervisor_id": to})
	})
	elapsed := time.Since(begin)

	time.Sleep(1500 * time.Millisecond)
	printTally(tally)
	fmt.Printf("waktu total : %dms\n", elapsed.Milliseconds())
	fmt.Printf("deadlock terdeteksi PostgreSQL selama uji: %d\n", r.deadlocks()-before)
	r.reportSupervisor(a)
	r.reportSupervisor(b)
	return nil
}

var _ = os.Stdout
