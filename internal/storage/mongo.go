package storage

import (
	"context"
	// "os"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	Client *mongo.Client
	Collection *mongo.Collection
}


func New(uri string, access bool) (*Store, error){
	if(!access){
		return &Store{}, nil
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	// collection := client.Database(os.Getenv("MONGO_DB")).Collection(os.Getenv("MONGO_COLLECTION"))
	
	collection := client.Database("webCrawlerArchive").Collection("webpages")
	// wipe collection on fresh start; comment out if you want resume later
	collection.DeleteMany(context.TODO(), struct{}{})
	
	return &Store{
		Client: client,
		Collection: collection,
	}, nil	
}


func (s *Store) Insert(doc any) {
	if s.Client == nil {
		fmt.Println("Skipping insert: no MongoDB client")
		return
	}
	res, err := s.Collection.InsertOne(context.TODO(), doc)
	if err != nil {
		fmt.Println("MongoDB insert error:", err)
	} else {
		fmt.Println("Inserted:", res.InsertedID)
	}
}


func (s *Store) Close() {
	if s.Client != nil {
		s.Client.Disconnect(context.TODO())
	}
}