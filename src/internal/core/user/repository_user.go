package user

import (
	"context"
	"errors"

	"github.com/ctfloyd/hazelmere-api/src/internal/database"
	"github.com/ctfloyd/hazelmere-api/src/internal/foundation/monitor"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id string) (UserData, error)
	GetUserByRunescapeName(ctx context.Context, runescapeName string) (UserData, error)
	GetAllUsers(ctx context.Context) ([]UserData, error)
	GetUsersWithTrackingEnabled(ctx context.Context) ([]UserData, error)
	CreateUser(ctx context.Context, user UserData) (UserData, error)
	UpdateUser(ctx context.Context, user UserData) (UserData, error)
}

type mongoUserRepository struct {
	monitor    *monitor.Monitor
	collection *mongo.Collection
}

func NewUserRepository(userCollection *mongo.Collection, mon *monitor.Monitor) UserRepository {
	return &mongoUserRepository{
		collection: userCollection,
		monitor:    mon,
	}
}

func (ur *mongoUserRepository) GetUserById(ctx context.Context, id string) (UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.GetUserById")
	defer span.End()

	filter := bson.M{"_id": id}

	result := ur.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return UserData{}, database.ErrNotFound
		}

		return UserData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var user UserData
	err := result.Decode(&user)
	if err != nil {
		return UserData{}, errors.Join(database.ErrGeneric, err)
	}

	return user, nil
}

func (ur *mongoUserRepository) GetUserByRunescapeName(ctx context.Context, runescapeName string) (UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.GetUserByRunescapeName")
	defer span.End()

	filter := bson.M{"runescapeName": runescapeName}

	result := ur.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return UserData{}, database.ErrNotFound
		}

		return UserData{}, errors.Join(database.ErrGeneric, result.Err())
	}

	var user UserData
	err := result.Decode(&user)
	if err != nil {
		return UserData{}, errors.Join(database.ErrGeneric, err)
	}

	return user, nil
}

func (ur *mongoUserRepository) GetAllUsers(ctx context.Context) ([]UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.GetAllUsers")
	defer span.End()

	cursor, err := ur.collection.Find(ctx, bson.D{})
	if err != nil {
		return []UserData{}, errors.Join(database.ErrGeneric, err)
	}

	var results []UserData
	if err = cursor.All(ctx, &results); err != nil {
		return []UserData{}, errors.Join(database.ErrGeneric, err)
	}

	return results, nil
}

func (ur *mongoUserRepository) GetUsersWithTrackingEnabled(ctx context.Context) ([]UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.GetUsersWithTrackingEnabled")
	defer span.End()

	filter := bson.M{"trackingStatus": string(TrackingStatusEnabled)}

	cursor, err := ur.collection.Find(ctx, filter)
	if err != nil {
		return []UserData{}, errors.Join(database.ErrGeneric, err)
	}

	var results []UserData
	if err = cursor.All(ctx, &results); err != nil {
		return []UserData{}, errors.Join(database.ErrGeneric, err)
	}

	return results, nil
}

func (ur *mongoUserRepository) CreateUser(ctx context.Context, user UserData) (UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.CreateUser")
	defer span.End()

	_, err := ur.collection.InsertOne(ctx, user)
	if err != nil {
		return UserData{}, errors.Join(database.ErrGeneric, err)
	}
	return user, nil
}

func (ur *mongoUserRepository) UpdateUser(ctx context.Context, user UserData) (UserData, error) {
	ctx, span := ur.monitor.StartSpan(ctx, "mongoUserRepository.UpdateUser")
	defer span.End()

	filter := bson.M{"_id": user.Id}
	update := bson.M{"$set": user}

	_, err := ur.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return UserData{}, errors.Join(database.ErrGeneric, err)
	}

	return user, nil
}
