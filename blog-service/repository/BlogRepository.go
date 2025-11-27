package repository

import (
	"context"
	"log"
	"time"

	"blog-service/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogRepository struct {
	coll *mongo.Collection
}

func NewBlogRepository(db *mongo.Database) *BlogRepository {
	r := &BlogRepository{coll: db.Collection("blogs")}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.ensureIndexes(ctx); err != nil {
		log.Printf("blog repository: ensure indexes: %v", err)
	}
	return r
}

func (r *BlogRepository) Create(ctx context.Context, b *model.Blog) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	result, err := r.coll.InsertOne(ctx, b)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		b.ID = oid
	}
	return nil
}

func (r *BlogRepository) GetAll(ctx context.Context) ([]model.Blog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []model.Blog
	for cur.Next(ctx) {
		var b model.Blog
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, cur.Err()
}

// GetByAuthorIDs returns blogs written by any of the provided author IDs.
// If authors slice is empty, returns an empty list.
func (r *BlogRepository) GetByAuthorIDs(ctx context.Context, authors []string) ([]model.Blog, error) {
	if len(authors) == 0 {
		return []model.Blog{}, nil
	}
	filter := bson.M{"author_id": bson.M{"$in": authors}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []model.Blog
	for cur.Next(ctx) {
		var b model.Blog
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, cur.Err()
}

func (r *BlogRepository) GetByID(ctx context.Context, id string) (*model.Blog, error) {
	var b model.Blog
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.coll.FindOne(ctx, bson.M{"_id": oid}).Decode(&b)
	return &b, err
}

func (r *BlogRepository) ensureIndexes(ctx context.Context) error {
	// Keep a descending index on created_at for listing/recent queries.
	models := []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	}
	_, err := r.coll.Indexes().CreateMany(ctx, models)
	return err
}

// UpdateLikesCount increments/decrements likes_count by delta (use +1 or -1)
func (r *BlogRepository) UpdateLikesCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	_, err := r.coll.UpdateByID(ctx, id, bson.M{"$inc": bson.M{"likes_count": delta}})
	return err
}

// Update updates a blog's title, description, and images
func (r *BlogRepository) Update(ctx context.Context, id string, title, description string, images []string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{
			"title":       title,
			"description": description,
			"images":      images,
		},
	}
	_, err = r.coll.UpdateByID(ctx, oid, update)
	return err
}

// Delete removes a blog by ID
func (r *BlogRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// GetByAuthorID returns blogs written by a specific author
func (r *BlogRepository) GetByAuthorID(ctx context.Context, authorID string) ([]model.Blog, error) {
	filter := bson.M{"author_id": authorID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []model.Blog
	for cur.Next(ctx) {
		var b model.Blog
		if err := cur.Decode(&b); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, cur.Err()
}
