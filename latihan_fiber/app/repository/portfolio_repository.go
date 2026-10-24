package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"latihan_fiber/app/model"
)

var (
	ErrInvalidID       = errors.New("format id tidak valid")
	ErrProjectLimit    = errors.New("jumlah proyek sudah mencapai batas")
	ErrProjectNotFound = errors.New("proyek tidak ditemukan")
)

const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	GetByStudent(ctx context.Context, studentID int) (model.Portfolio, error)
	Replace(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) (model.Portfolio, error)
	TopTechnologies(ctx context.Context, limit int) ([]model.TechnologyCount, error)
}

// portfolioDocument adalah bentuk dokumen yang benar-benar disimpan di
// MongoDB. Struct model tetap bebas dari tipe driver.
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
	Link         string        `bson:"link"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	Name   string `bson:"name"`
	Issuer string `bson:"issuer"`
	Date   string `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		technologies := p.Technologies
		if technologies == nil {
			technologies = []string{}
		}
		projects = append(projects, model.Project{
			ID: p.ID.Hex(), Title: p.Title, Technologies: technologies,
			Link: p.Link, Year: p.Year,
		})
	}
	certificates := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certificates = append(certificates, model.Certificate{
			Name: c.Name, Issuer: c.Issuer, Date: c.Date,
		})
	}
	return model.Portfolio{
		StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certificates,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type portfolioMongoRepository struct{ coll *mongo.Collection }

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

func portfolioNow() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

// EnsureIndexes membuat aturan "satu mahasiswa satu portofolio" di level
// database. Index unik tidak dapat dilewati oleh request yang serentak.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_unique"),
		},
		{
			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("project_technologies"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) GetByStudent(
	ctx context.Context, studentID int,
) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) Replace(
	ctx context.Context, p model.Portfolio,
) (model.Portfolio, error) {
	t := portfolioNow()
	projects := make([]projectDocument, 0, len(p.Projects))
	for _, item := range p.Projects {
		projects = append(projects, projectDocument{
			ID: bson.NewObjectID(), Title: item.Title,
			Technologies: item.Technologies, Link: item.Link, Year: item.Year,
		})
	}
	certificates := make([]certificateDocument, 0, len(p.Certificates))
	for _, item := range p.Certificates {
		certificates = append(certificates, certificateDocument{
			Name: item.Name, Issuer: item.Issuer, Date: item.Date,
		})
	}

	update := bson.M{
		"$set": bson.M{
			"headline": p.Headline, "bio": p.Bio,
			"projects": projects, "certificates": certificates,
			"updated_at": t,
		},
		"$setOnInsert": bson.M{"created_at": t},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var doc portfolioDocument
	if err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": p.StudentID}, update, opts,
	).Decode(&doc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Portfolio{}, ErrDuplicate
		}
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// AddProject memakai syarat "projects.19 belum ada" agar pemeriksaan batas dan
// penambahan terjadi dalam satu operasi atomik. Pola ini sama dengan batas
// balasan pada Langkah 6 modul.
func (r *portfolioMongoRepository) AddProject(
	ctx context.Context, studentID int, p model.Project,
) (model.Portfolio, error) {
	doc := projectDocument{
		ID: bson.NewObjectID(), Title: p.Title,
		Technologies: p.Technologies, Link: p.Link, Year: p.Year,
	}
	limitField := fmt.Sprintf("projects.%d", MaxProjectsPerPortfolio-1)

	var out portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{
			"student_id": studentID,
			limitField:   bson.M{"$exists": false},
		},
		bson.M{
			"$push": bson.M{"projects": doc},
			"$set":  bson.M{"updated_at": portfolioNow()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&out)
	if err == nil {
		return out.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(
	ctx context.Context, studentID int, projectID string,
) (model.Portfolio, error) {
	oid, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return model.Portfolio{}, ErrInvalidID
	}

	var out portfolioDocument
	err = r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, "projects._id": oid},
		bson.M{
			"$pull": bson.M{"projects": bson.M{"_id": oid}},
			"$set":  bson.M{"updated_at": portfolioNow()},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&out)
	if err == nil {
		return out.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menghapus proyek: %w", err)
	}

	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectNotFound
}

// TopTechnologies menghitung teknologi paling banyak dipakai. Dua $unwind
// diperlukan: satu untuk memecah array projects, satu lagi untuk memecah array
// technologies di dalam setiap proyek.
func (r *portfolioMongoRepository) TopTechnologies(
	ctx context.Context, limit int,
) ([]model.TechnologyCount, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$projects"}},
		bson.D{{Key: "$unwind", Value: "$projects.technologies"}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": "$projects.technologies", "total": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1}, {Key: "_id", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: limit}},
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
		result = append(result, model.TechnologyCount{
			Technology: row.Technology, Total: row.Total,
		})
	}
	return result, nil
}
