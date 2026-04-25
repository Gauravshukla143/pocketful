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

const kycCollection = "kyc"

// KYCRepository handles database operations for KYC sessions.
type KYCRepository struct {
	collection *mongo.Collection
}

// NewKYCRepository creates a new KYCRepository instance.
func NewKYCRepository() *KYCRepository {
	return &KYCRepository{
		collection: db.GetCollection(kycCollection),
	}
}

// Create inserts a new KYC session into the database.
func (r *KYCRepository) Create(ctx context.Context, kyc *models.KYC) (*models.KYC, error) {
	kyc.ID = primitive.NewObjectID()
	kyc.CreatedAt = time.Now()
	kyc.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, kyc)
	if err != nil {
		return nil, err
	}
	return kyc, nil
}

// FindByUserID retrieves the most recent KYC session for a user.
func (r *KYCRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) (*models.KYC, error) {
	var kyc models.KYC
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&kyc)
	if err != nil {
		return nil, err
	}
	return &kyc, nil
}

// FindByID retrieves a KYC session by its ID.
func (r *KYCRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.KYC, error) {
	var kyc models.KYC
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&kyc)
	if err != nil {
		return nil, err
	}
	return &kyc, nil
}

// UpdateStatus updates the status of a KYC session.
func (r *KYCRepository) UpdateStatus(ctx context.Context, kycID primitive.ObjectID, status models.KYCStatus, reviewedBy primitive.ObjectID, rejectionNote string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":         status,
			"rejection_note": rejectionNote,
			"reviewed_by":    reviewedBy,
			"reviewed_at":    now,
			"updated_at":     now,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": kycID}, update)
	return err
}

// UpdateStatusOnly updates only the status (used by async worker).
func (r *KYCRepository) UpdateStatusOnly(ctx context.Context, kycID primitive.ObjectID, status models.KYCStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": kycID}, update)
	return err
}

// ExistsByUserID checks if a KYC session already exists for a user.
func (r *KYCRepository) ExistsByUserID(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
