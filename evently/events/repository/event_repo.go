package repository

import (
	"context"
	"errors"
	"events/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventRepository interface {
	Create(event *models.Event) (*models.Event, error)
	FindByID(id string) (*models.Event, error)
	FindAll(page int64, limit int64) ([]models.Event, error)
	FindAllUpcomingEvents(page, limit int64) ([]models.UpcomingEvent, error)
	FindAvailableSeatsForIds(ids []string) (map[string]int64, error)
	UpdateFields(id string, updates map[string]interface{}) error
	GetCapacityUtilization(ctx context.Context, eventID string, page, limit int64) ([]models.CapacityUtilization, error)
	GetMostBookedEvents(ctx context.Context, limit int64) ([]models.MostBookedEvent, error)
	GetMostPopularEvents(ctx context.Context, limit int64) ([]models.MostPopularEvent, error)
	Delete(id string) error
}

type eventRepo struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewEventRepository(db *mongo.Database) EventRepository {
	return &eventRepo{
		collection: db.Collection("events"),
		ctx:        context.Background(),
	}
}

func (r *eventRepo) Create(event *models.Event) (*models.Event, error) {
	event.ID = primitive.NewObjectID()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(r.ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (r *eventRepo) FindByID(id string) (*models.Event, error) {
	eventId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var event models.Event
	err = r.collection.FindOne(r.ctx, bson.M{"_id": eventId}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("event not found")
		}

		return nil, err
	}

	return &event, nil
}

func (r *eventRepo) FindAvailableSeatsForIds(ids []string) (map[string]int64, error) {
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}
		objectIDs = append(objectIDs, oid)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	projection := options.Find().SetProjection(bson.M{"available_seats": 1})

	cursor, err := r.collection.Find(r.ctx, filter, projection)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	result := make(map[string]int64)
	for cursor.Next(r.ctx) {
		var ev struct {
			ID             primitive.ObjectID `bson:"_id"`
			AvailableSeats int64              `bson:"available_seats"`
		}
		if err := cursor.Decode(&ev); err != nil {
			return nil, err
		}
		result[ev.ID.Hex()] = ev.AvailableSeats
	}

	return result, nil
}

func (r *eventRepo) FindAll(page int64, limit int64) ([]models.Event, error) {
	skip := (page - 1) * limit

	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)
	findOptions.SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.collection.Find(r.ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var events []models.Event
	for cursor.Next(r.ctx) {
		var event models.Event
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *eventRepo) FindAllUpcomingEvents(page, limit int64) ([]models.UpcomingEvent, error) {
	skip := (page - 1) * limit
	now := time.Now()

	findOptions := options.Find()
	findOptions.SetLimit(limit)
	findOptions.SetSkip(skip)
	findOptions.SetSort(bson.D{{Key: "date", Value: 1}})

	findOptions.SetProjection(bson.M{
		"title":           1,
		"venue":           1,
		"date":            1,
		"available_seats": 1,
		"total_seats":     1,
	})

	cursor, err := r.collection.Find(r.ctx, bson.M{"date": bson.M{"$gt": now}}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var events []models.UpcomingEvent
	for cursor.Next(r.ctx) {
		var event models.UpcomingEvent
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err1 := cursor.Err(); err1 != nil {
		return nil, err1
	}

	return events, nil
}

func (r *eventRepo) UpdateFields(id string, updates map[string]interface{}) error {
	eventId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updates["updated_at"] = time.Now()

	filter := bson.M{"_id": eventId}
	update := bson.M{"$set": updates}

	res, err1 := r.collection.UpdateOne(r.ctx, filter, update)

	if err1 != nil {
		return err1
	}

	if res.MatchedCount == 0 {
		return errors.New("event not found")
	}

	return nil
}

func (r *eventRepo) Delete(id string) error {
	eventId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	res, err1 := r.collection.DeleteOne(r.ctx, bson.M{"_id": eventId})
	if err1 != nil {
		return err1
	}

	if res.DeletedCount == 0 {
		return errors.New("event not found")
	}

	return nil
}

func (r *eventRepo) GetMostBookedEvents(ctx context.Context, limit int64) ([]models.MostBookedEvent, error) {

	pipeline := mongo.Pipeline{
		{{
			Key: "$project",
			Value: bson.D{
				{Key: "event_id", Value: "$_id"},
				{Key: "title", Value: "$title"},
				{Key: "booked_seats", Value: bson.D{
					{Key: "$subtract", Value: bson.A{"$total_seats", "$available_seats"}},
				}},
			},
		}},
		{{
			Key:   "$sort",
			Value: bson.D{{Key: "booked_seats", Value: -1}},
		}},
		{{
			Key:   "$limit",
			Value: limit,
		}},
	}

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []models.MostBookedEvent
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *eventRepo) GetMostPopularEvents(ctx context.Context, limit int64) ([]models.MostPopularEvent, error) {

	pipeline := mongo.Pipeline{
		{{
			Key: "$project",
			Value: bson.D{
				{Key: "event_id", Value: "$_id"},
				{Key: "title", Value: "$title"},
				{Key: "total_seats", Value: 1},
				{Key: "booked_seats", Value: bson.D{
					{Key: "$subtract", Value: bson.A{"$total_seats", "$available_seats"}},
				}},
				{Key: "occupancy_rate", Value: bson.D{
					{Key: "$round", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{
							bson.D{{Key: "$subtract", Value: bson.A{"$total_seats", "$available_seats"}}},
							"$total_seats",
						}}},
						2,
					}},
				}},
			},
		}},
		{{
			Key:   "$sort",
			Value: bson.D{{Key: "occupancy_rate", Value: -1}},
		}},
		{{
			Key:   "$limit",
			Value: limit,
		}},
	}

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []models.MostPopularEvent
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *eventRepo) GetCapacityUtilization(ctx context.Context, eventID string, page, limit int64) ([]models.CapacityUtilization, error) {

	matchStage := bson.D{}
	if eventID != "" {
		matchStage = bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: eventID}}}}
	}

	pipeline := mongo.Pipeline{}

	if len(matchStage) > 0 {
		pipeline = append(pipeline, matchStage)
	}

	projectStage := bson.D{{
		Key: "$project",
		Value: bson.D{
			{Key: "event_id", Value: "$_id"},
			{Key: "title", Value: "$title"},
			{Key: "capacity_utilisation", Value: bson.D{

				{Key: "$round", Value: bson.A{
					bson.D{{Key: "$multiply", Value: bson.A{
						bson.D{{Key: "$divide", Value: bson.A{
							bson.D{{Key: "$subtract", Value: bson.A{"$total_seats", "$available_seats"}}},
							"$total_seats",
						}}},
						100,
					}}},
					2,
				}},
			}},
		},
	}}

	pipeline = append(pipeline, projectStage)
	skip := (page - 1) * limit

	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.D{{Key: "capacity_utilisation", Value: -1}}}},

		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []models.CapacityUtilization
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
