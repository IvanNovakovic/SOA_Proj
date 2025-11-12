package repository

import (
	"context"
	"log"
	"time"

	"stakeholders-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	r := &UserRepository{coll: db.Collection("users")}
	// Ensure indexes once at startup (non-fatal if this fails)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.ensureIndexes(ctx); err != nil {
		log.Printf("user repository: ensure indexes: %v", err)
	}
	return r
}

// Create a new user; returns inserted ObjectID.
func (r *UserRepository) Create(ctx context.Context, u *model.User) (primitive.ObjectID, error) {
	res, err := r.coll.InsertOne(ctx, u)
	if err != nil {
		return primitive.NilObjectID, err
	}
	id, _ := res.InsertedID.(primitive.ObjectID)
	return id, nil
}

// GetAll returns every user (use List for pagination/filters).
func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	cur, err := r.coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []model.User
	for cur.Next(ctx) {
		var u model.User
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, cur.Err()
}

// List with optional filter and pagination.
func (r *UserRepository) List(ctx context.Context, filter bson.M, skip, limit int64) ([]model.User, error) {
	opts := options.Find()
	if skip > 0 {
		opts.SetSkip(skip)
	}
	if limit > 0 {
		opts.SetLimit(limit)
	}
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var users []model.User
	for cur.Next(ctx) {
		var u model.User
		if err := cur.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, cur.Err()
}

// GET METHODS -----------------------------------------------------

func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (model.User, error) {
	var u model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	return u, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&u)
	return u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var u model.User
	err := r.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	return u, err
}

// UUPDATE METHODS ---------------------------------------------

// UpdateFields updates arbitrary fields (except _id). Prefer this for profile updates.
// Example fields: bson.M{"name": "Ivan", "surname": "Novakovic", "roles": []string{"user"}}
func (r *UserRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
	delete(fields, "_id")
	_, err := r.coll.UpdateByID(ctx, id, bson.M{"$set": fields})
	return err
}

// UpdatePassword updates only the password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, id primitive.ObjectID, passwordHash string) error {
	_, err := r.coll.UpdateByID(ctx, id, bson.M{"$set": bson.M{"password": passwordHash}})
	return err
}

// DeleteByID performs a hard delete. If you want soft-delete, add a Deleted flag to the model.
func (r *UserRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// HELPERS -----------------------------------------------------

// ensureIndexes creates helpful indexes (idempotent).
func (r *UserRepository) ensureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
		// Geo index on address.location if you plan geo queries
		{
			// Only index documents where location exists.
			Keys: bson.D{{Key: "address.location", Value: "2dsphere"}},
			Options: options.Index().
				SetPartialFilterExpression(bson.D{{Key: "address.location", Value: bson.D{{Key: "$exists", Value: true}}}}),
		},
	}
	_, err := r.coll.Indexes().CreateMany(ctx, models)
	return err
}
