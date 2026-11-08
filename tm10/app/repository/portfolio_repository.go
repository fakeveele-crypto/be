package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"latihan-fiber/tm10/app/model"
)

var (
	ErrProjectLimit    = errors.New("jumlah proyek sudah mencapai batas")
	ErrProjectNotFound = errors.New("proyek tidak ditemukan")
)

const dateLayout = "2006-01-02"

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error)
	// Replace membuat atau mengganti portofolio. Nilai created true bila
	// dokumennya baru dibuat.
	Replace(ctx context.Context, studentID int, req model.ReplacePortfolioRequest) (p model.Portfolio, created bool, err error)
	AddProject(ctx context.Context, studentID int, in model.ProjectInput) (model.Portfolio, error)
	RemoveProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyCount, error)
}

// portfolioDocument adalah bentuk portofolio SEBAGAIMANA TERSIMPAN di MongoDB.
type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	Name   string    `bson:"name"`
	Issuer string    `bson:"issuer"`
	Date   time.Time `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID: p.ID.Hex(), Title: p.Title, Technologies: techs, URL: p.URL, Year: p.Year,
		})
	}
	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			Name: c.Name, Issuer: c.Issuer, Date: c.Date.UTC().Format(dateLayout),
		})
	}
	return model.Portfolio{
		ID: d.ID.Hex(), StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certs,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func newProjectDocument(in model.ProjectInput) projectDocument {
	techs := in.Technologies
	if techs == nil {
		techs = []string{}
	}
	return projectDocument{
		ID: bson.NewObjectID(), Title: in.Title, Technologies: techs, URL: in.URL, Year: in.Year,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes membuat index unik pada student_id: "satu mahasiswa satu
// portofolio" dijaga DATABASE, bukan hanya kode. Index ini juga melayani
// semua endpoint portofolio, karena semuanya mencari lewat student_id.
// Endpoint statistik teknologi memang membaca seluruh collection (pipeline
// $unwind atas semua dokumen), sehingga tidak ada index yang membantunya.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("student_unique"),
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	if err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Replace membuat portofolio bila belum ada, atau mengganti isinya bila sudah.
//
// Satu operasi upsert. Dua PUT pertama yang datang serentak untuk mahasiswa
// yang sama dapat sama-sama tidak menemukan dokumen. Yang memastikan hanya
// SATU dokumen yang lahir adalah index unik student_id, bukan kode ini. Pada
// MongoDB 4.2 ke atas, upsert dengan filter kesamaan pada field ber-index unik
// yang kalah balapan diulang sendiri oleh server. Pengulangan satu kali di
// bawah hanya sabuk pengaman bila suatu saat error duplikat tetap sampai ke
// aplikasi: pada percobaan kedua dokumennya sudah ada, sehingga operasinya
// menjadi update biasa.
func (r *portfolioMongoRepository) Replace(ctx context.Context, studentID int, req model.ReplacePortfolioRequest) (model.Portfolio, bool, error) {
	// Slice nil tersimpan sebagai null di BSON, dan $push ke field null GAGAL.
	// Karena itu array selalu ditulis sebagai array kosong, bukan nil.
	projects := make([]projectDocument, 0, len(req.Projects))
	for _, p := range req.Projects {
		projects = append(projects, newProjectDocument(p))
	}
	certs := make([]certificateDocument, 0, len(req.Certificates))
	for _, c := range req.Certificates {
		d, err := time.Parse(dateLayout, c.Date)
		if err != nil {
			return model.Portfolio{}, false, fmt.Errorf("tanggal sertifikat: %w", err)
		}
		certs = append(certs, certificateDocument{Name: c.Name, Issuer: c.Issuer, Date: d})
	}

	t := now()
	update := bson.M{
		"$set": bson.M{
			"headline": req.Headline, "bio": req.Bio,
			"projects": projects, "certificates": certs,
			"updated_at": t,
		},
		"$setOnInsert": bson.M{"created_at": t},
	}
	opts := options.UpdateOne().SetUpsert(true)

	var res *mongo.UpdateResult
	var err error
	for attempt := 0; attempt < 2; attempt++ {
		res, err = r.coll.UpdateOne(ctx, bson.M{"student_id": studentID}, update, opts)
		if err == nil || !mongo.IsDuplicateKeyError(err) {
			break
		}
	}
	if err != nil {
		return model.Portfolio{}, false, fmt.Errorf("menyimpan portofolio: %w", err)
	}

	p, err := r.FindByStudent(ctx, studentID)
	if err != nil {
		return model.Portfolio{}, false, err
	}
	return p, res.UpsertedCount > 0, nil
}

// AddProject menambahkan satu proyek ke dalam array projects.
//
// Pola yang sama dengan AddReply: syarat "projects.19 tidak ada" berarti
// array belum berisi 20 elemen, dan $push hanya terjadi pada dokumen yang
// cocok. Pemeriksaan dan penambahan satu operasi, sehingga permintaan
// serentak tidak dapat melampaui batas.
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, in model.ProjectInput) (model.Portfolio, error) {
	limitField := "projects." + strconv.Itoa(model.MaxProjectsPerPortfolio-1)

	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		bson.M{
			"$push": bson.M{"projects": newProjectDocument(in)},
			"$set":  bson.M{"updated_at": now()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Tidak ada dokumen yang cocok: portofolionya tidak ada, atau penuh.
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// RemoveProject menghapus satu proyek dengan $pull.
//
// Filter memuat "projects._id", sehingga dokumen hanya cocok bila proyeknya
// memang ada di dalamnya.
func (r *portfolioMongoRepository) RemoveProject(ctx context.Context, studentID int, projectID string) error {
	oid, err := parseID(projectID)
	if err != nil {
		return err
	}
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"student_id": studentID, "projects._id": oid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": oid}},
			"$set":  bson.M{"updated_at": now()},
		},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount > 0 {
		return nil
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return ErrProjectNotFound
}

// TopTechnologies menghitung teknologi yang paling banyak dipakai di seluruh
// portofolio. Dua kali $unwind: pertama memecah array projects (satu dokumen
// per proyek), lalu memecah array technologies di dalamnya (satu dokumen per
// pasangan proyek dan teknologi). Barulah $group dapat menghitungnya.
// Nilai total = jumlah proyek yang memakai teknologi itu.
func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyCount, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}, {Key: "_id", Value: 1}}}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var rows []struct {
		Technology string `bson:"_id"`
		Total      int    `bson:"total"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}
	result := make([]model.TechnologyCount, 0, len(rows))
	for _, row := range rows {
		result = append(result, model.TechnologyCount{Technology: row.Technology, Total: row.Total})
	}
	return result, nil
}
