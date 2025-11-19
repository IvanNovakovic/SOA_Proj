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
    // create simple index on authorId and tourId
    _, _ = col.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "authorId", Value: 1}},
    })
    _, _ = kpCol.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "tourId", Value: 1}},
    })
    return &TourRepository{client: client, col: col, kpCol: kpCol}, nil
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

// KeyPoint methods
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
    cur, err := r.kpCol.Find(ctx, filter)
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
