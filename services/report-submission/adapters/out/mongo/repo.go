// Package mongo is the MongoDB ReportRepository. It is used when MONGO_URL is
// set; otherwise the service keeps reports in memory. Collection and indexes:
// docs/design/database-schema.md section 5.
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/jurwaCST/Jurwa/services/report-submission/domain"
)

const collectionName = "reports"

// Repository stores reports in the "reports" collection.
type Repository struct {
	client *mongo.Client
	db     *mongo.Database
	coll   *mongo.Collection
}

// Connect opens the connection and checks it with a ping.
func Connect(ctx context.Context, url, database string) (*Repository, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(url))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	db := client.Database(database)
	return &Repository{client: client, db: db, coll: db.Collection(collectionName)}, nil
}

// EnsureIndexes creates the two indexes from the schema. Safe to run again.
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "reporter_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "category", Value: 1}, {Key: "status", Value: 1}}},
	})
	return err
}

// Save stores the report, replacing any report with the same id.
func (r *Repository) Save(ctx context.Context, report domain.Report) error {
	doc := toDocument(report)
	_, err := r.coll.ReplaceOne(ctx, bson.D{{Key: "_id", Value: doc.ID}}, doc, options.Replace().SetUpsert(true))
	return err
}

// FindByID returns the report or domain.ErrReportNotFound.
func (r *Repository) FindByID(ctx context.Context, id string) (domain.Report, error) {
	var doc document
	err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.Report{}, domain.ErrReportNotFound
	}
	if err != nil {
		return domain.Report{}, err
	}
	return doc.toReport(), nil
}

// Close ends the connection.
func (r *Repository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

// DropDatabase deletes the whole database. Used by tests only.
func (r *Repository) DropDatabase(ctx context.Context) error {
	return r.db.Drop(ctx)
}
