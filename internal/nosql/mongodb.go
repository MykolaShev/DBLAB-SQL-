package nosql

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FreelancerProfileDoc represents the denormalized profile document.
// This replaces the need to JOIN users, freelancer_profiles, skills, and reviews.
type FreelancerProfileDoc struct {
	UserID      int      `bson:"user_id"`
	Username    string   `bson:"username"`
	Bio         string   `bson:"bio"`
	HourlyRate  float64  `bson:"hourly_rate"`
	Skills      []string `bson:"skills"`
	ReviewScore float64  `bson:"review_score"`
}

type MongoStore struct {
	client *mongo.Client
	col    *mongo.Collection
}

func NewMongoStore(ctx context.Context, uri, dbName, colName string) (*MongoStore, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	col := client.Database(dbName).Collection(colName)

	// Create an index on user_id for fast read operations
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = col.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &MongoStore{client: client, col: col}, nil
}

func (m *MongoStore) InsertProfile(ctx context.Context, profile FreelancerProfileDoc) error {
	_, err := m.col.InsertOne(ctx, profile)
	return err
}

func (m *MongoStore) GetProfileByID(ctx context.Context, userID int) (*FreelancerProfileDoc, error) {
	var profile FreelancerProfileDoc
	filter := bson.D{{Key: "user_id", Value: userID}}
	err := m.col.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (m *MongoStore) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}