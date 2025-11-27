package repository

import (
	"context"
	"time"

	"tour-service/model"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TourRepository struct {
	client *mongo.Client
	col    *mongo.Collection
	kpCol  *mongo.Collection
	revCol *mongo.Collection
}

func NewTourRepository(ctx context.Context, uri string, dbName string) (*TourRepository, error) {
	clientOpts := options.Client().ApplyURI(uri).SetConnectTimeout(10 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	col := db.Collection("tours")
	kpCol := db.Collection("keypoints")
	revCol := db.Collection("reviews")
	// create simple index on authorId and tourId and indexes for keypoints/reviews
	_, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "authorId", Value: 1}},
	})
	_, _ = kpCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tourId", Value: 1}},
	})
	// reviews: index by tourId to lookup reviews for a tour, and by authorId if needed
	_, _ = revCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "tourId", Value: 1}},
	})
	_, _ = revCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "authorId", Value: 1}},
	})
	return &TourRepository{client: client, col: col, kpCol: kpCol, revCol: revCol}, nil
}

func (r *TourRepository) Close(ctx context.Context) error {
	return r.client.Disconnect(ctx)
}

func (r *TourRepository) CreateTour(ctx context.Context, t *model.Tour) (*model.Tour, error) {
	t.CreatedAt = time.Now().UTC()
	// ensure defaults on creation
	if t.Status == "" {
		t.Status = "draft"
	}
	// initial price should be 0
	t.Price = 0
	res, err := r.col.InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}
	t.ID = res.InsertedID.(primitive.ObjectID)
	return t, nil
}

func (r *TourRepository) GetTourByID(ctx context.Context, tourId string) (*model.Tour, error) {
	objID, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		return nil, err
	}
	var tour model.Tour
	err = r.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

func (r *TourRepository) GetToursByAuthor(ctx context.Context, authorId string) ([]model.Tour, error) {
	filter := bson.M{"authorId": authorId}
	cur, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var tours []model.Tour
	for cur.Next(ctx) {
		var t model.Tour
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		tours = append(tours, t)
	}
	return tours, nil
}

func (r *TourRepository) UpdateTour(ctx context.Context, tourId string, authorId string, updates map[string]interface{}) (*model.Tour, error) {
	objID, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		return nil, err
	}
	// Verify ownership
	filter := bson.M{"_id": objID, "authorId": authorId}
	// Prevent updating certain fields
	delete(updates, "_id")
	delete(updates, "authorId")
	delete(updates, "createdAt")

	update := bson.M{"$set": updates}
	var tour model.Tour
	err = r.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
} // KeyPoint methods
func (r *TourRepository) CreateKeyPoint(ctx context.Context, kp *model.KeyPoint) (*model.KeyPoint, error) {
	kp.CreatedAt = time.Now().UTC()
	// ensure TourID is set
	if kp.TourID.IsZero() {
		return nil, mongo.ErrNilDocument
	}
	res, err := r.kpCol.InsertOne(ctx, kp)
	if err != nil {
		return nil, err
	}
	kp.ID = res.InsertedID.(primitive.ObjectID)
	return kp, nil
}

func (r *TourRepository) GetKeyPointsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.KeyPoint, error) {
	filter := bson.M{"tourId": tourId}
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}, {Key: "createdAt", Value: 1}})
	cur, err := r.kpCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var kps []model.KeyPoint
	for cur.Next(ctx) {
		var kp model.KeyPoint
		if err := cur.Decode(&kp); err != nil {
			return nil, err
		}
		kps = append(kps, kp)
	}
	return kps, nil
}

func (r *TourRepository) UpdateKeyPoint(ctx context.Context, keypointId string, updates map[string]interface{}) (*model.KeyPoint, error) {
	objID, err := primitive.ObjectIDFromHex(keypointId)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": objID}
	// Prevent updating certain fields
	delete(updates, "_id")
	delete(updates, "tourId")
	delete(updates, "createdAt")

	update := bson.M{"$set": updates}
	var kp model.KeyPoint
	err = r.kpCol.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&kp)
	if err != nil {
		return nil, err
	}
	return &kp, nil
}

func (r *TourRepository) DeleteKeyPoint(ctx context.Context, keypointId string) error {
	objID, err := primitive.ObjectIDFromHex(keypointId)
	if err != nil {
		return err
	}
	_, err = r.kpCol.DeleteOne(ctx, bson.M{"_id": objID})
	return err
}

func (r *TourRepository) UpdateKeyPointsOrder(ctx context.Context, tourId primitive.ObjectID, orderedIds []string) error {
	// Update each keypoint with its new order
	for i, idStr := range orderedIds {
		objID, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		filter := bson.M{"_id": objID, "tourId": tourId}
		update := bson.M{"$set": bson.M{"order": i}}
		_, err = r.kpCol.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

// Review methods
func (r *TourRepository) CreateReview(ctx context.Context, rev *model.Review) (*model.Review, error) {
	if rev == nil {
		return nil, mongo.ErrNilDocument
	}
	rev.CreatedAt = time.Now().UTC()
	if rev.Rating < 1 {
		rev.Rating = 1
	}
	if rev.Rating > 5 {
		rev.Rating = 5
	}
	res, err := r.revCol.InsertOne(ctx, rev)
	if err != nil {
		return nil, err
	}
	rev.ID = res.InsertedID.(primitive.ObjectID)
	return rev, nil
}

func (r *TourRepository) GetReviewsByTour(ctx context.Context, tourId primitive.ObjectID) ([]model.Review, error) {
	filter := bson.M{"tourId": tourId}
	cur, err := r.revCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Review
	for cur.Next(ctx) {
		var rev model.Review
		if err := cur.Decode(&rev); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	return out, nil
}

// HasUserReviewedTour checks if a user has already reviewed a tour
func (r *TourRepository) HasUserReviewedTour(ctx context.Context, tourId primitive.ObjectID, authorId string) (bool, error) {
	filter := bson.M{
		"tourId":   tourId,
		"authorId": authorId,
	}
	count, err := r.revCol.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TourRepository) PublishTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error) {
	objID, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filter := bson.M{"_id": objID, "authorId": authorId}
	update := bson.M{
		"$set": bson.M{
			"status":      "published",
			"publishedAt": now,
			"archivedAt":  nil,
		},
	}

	var tour model.Tour
	err = r.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

func (r *TourRepository) ArchiveTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error) {
	objID, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	filter := bson.M{"_id": objID, "authorId": authorId, "status": "published"}
	update := bson.M{
		"$set": bson.M{
			"status":     "archived",
			"archivedAt": now,
		},
	}

	var tour model.Tour
	err = r.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}

func (r *TourRepository) ActivateTour(ctx context.Context, tourId string, authorId string) (*model.Tour, error) {
	objID, err := primitive.ObjectIDFromHex(tourId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objID, "authorId": authorId, "status": "archived"}
	update := bson.M{
		"$set": bson.M{
			"status":     "published",
			"archivedAt": nil,
		},
	}

	var tour model.Tour
	err = r.col.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&tour)
	if err != nil {
		return nil, err
	}
	return &tour, nil
}
