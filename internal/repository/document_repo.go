package repository

import (
	"context"
	"time"

	"pocketful/internal/db"
	"pocketful/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const documentsCollection = "documents"

// DocumentRepository handles database operations for KYC documents.
type DocumentRepository struct {
	collection *mongo.Collection
}

// NewDocumentRepository creates a new DocumentRepository instance.
func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		collection: db.GetCollection(documentsCollection),
	}
}

// Create inserts a new document record into the database.
func (r *DocumentRepository) Create(ctx context.Context, doc *models.Document) (*models.Document, error) {
	doc.ID = primitive.NewObjectID()
	doc.UploadedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FindByKycID retrieves all documents associated with a KYC session.
func (r *DocumentRepository) FindByKycID(ctx context.Context, kycID primitive.ObjectID) ([]models.Document, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"kyc_id": kycID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []models.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindByKycIDAndType retrieves a specific document type for a KYC session.
func (r *DocumentRepository) FindByKycIDAndType(ctx context.Context, kycID primitive.ObjectID, docType models.DocumentType) (*models.Document, error) {
	var doc models.Document
	err := r.collection.FindOne(ctx, bson.M{
		"kyc_id": kycID,
		"type":   docType,
	}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// CountByKycID counts documents uploaded for a given KYC session.
func (r *DocumentRepository) CountByKycID(ctx context.Context, kycID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"kyc_id": kycID})
}

// UpsertByKycIDAndType inserts or replaces a document by KYC ID and document type.
func (r *DocumentRepository) UpsertByKycIDAndType(ctx context.Context, doc *models.Document) (*models.Document, error) {
	doc.UploadedAt = time.Now()

	filter := bson.M{"kyc_id": doc.KycID, "type": doc.Type}
	existing, err := r.FindByKycIDAndType(ctx, doc.KycID, doc.Type)
	if err != nil {
		// Not found, create new
		return r.Create(ctx, doc)
	}
	// Update existing
	doc.ID = existing.ID
	update := bson.M{
		"$set": bson.M{
			"file_name":    doc.FileName,
			"file_path":    doc.FilePath,
			"content_type": doc.ContentType,
			"size_bytes":   doc.SizeBytes,
			"uploaded_at":  doc.UploadedAt,
		},
	}
	_, err = r.collection.UpdateOne(ctx, filter, update)
	return doc, err
}
