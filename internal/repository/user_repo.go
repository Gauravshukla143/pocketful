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

const usersCollection = "users"

// UserRepository handles database operations for users.
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository() *UserRepository {
	return &UserRepository{
		collection: db.GetCollection(usersCollection),
	}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsAdminUser checks if any admin user exists in the database.
func (r *UserRepository) ExistsAdminUser(ctx context.Context) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"role": models.RoleAdmin})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateIndex ensures the email field is unique.
func (r *UserRepository) CreateIndex(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: nil,
	}
	_, err := r.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}
