package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aritrosaha10/frasertickets/lib"
	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Event struct {
	ID                    primitive.ObjectID     `json:"id"              bson:"_id,omitempty"`
	Name                  string                 `json:"name"            bson:"name"`
	Description           string                 `json:"description"     bson:"description"`
	ImageURLs             []string               `json:"img_urls"        bson:"img_urls"`
	Location              string                 `json:"location"        bson:"location"` // Ex. name of venue
	Address               string                 `json:"address"         bson:"address"`
	StartTimestamp        time.Time              `json:"start_timestamp" bson:"start_timestamp"`
	EndTimestamp          time.Time              `json:"end_timestamp"   bson:"end_timestamp"`
	RawCustomFieldsSchema map[string]interface{} `json:"custom_fields_schema" bson:"custom_fields_schema"` // Schema for extra data in JSON Schema format
}

func (event *Event) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func GetAllEvents(ctx context.Context) ([]Event, error) {
	// Try to get data from MongoDB
	cursor, err := lib.Datastore.Db.Collection(eventsColName).Find(ctx, bson.D{})
	if err != nil {
		return []Event{}, err
	}

	// Attempt to convert BSON data into Event structs
	var events []Event
	if err := cursor.All(ctx, &events); err != nil {
		return []Event{}, err
	}

	return events, nil
}

func GetEvent(ctx context.Context, filter bson.M) (Event, error) {
	// Try to fetch data from DB
	var event Event
	err := lib.Datastore.Db.Collection(eventsColName).
		FindOne(ctx, filter).
		Decode(&event)

	// No error handling needed (user & err will default to empty struct / nil)
	return event, err
}

func CheckIfEventExists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	// Directly return results from DB
	count, err := lib.Datastore.Db.Collection(eventsColName).CountDocuments(ctx, bson.M{"_id": id})
	return count == 1, err
}

func CreateNewEvent(ctx context.Context, event Event) (primitive.ObjectID, error) {
	// Validate custom fields schema
	schemaLoader := gojsonschema.NewGoLoader(event.RawCustomFieldsSchema)
	_, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Try to add document
	res, err := lib.Datastore.Db.Collection(eventsColName).InsertOne(ctx, event)

	// Return object ID
	return res.InsertedID.(primitive.ObjectID), err
}

func ValidateCustomEventFields(ctx context.Context, event Event, customFields map[string]interface{}) (bool, []gojsonschema.ResultError, error) {
	schemaLoader := gojsonschema.NewGoLoader(event.RawCustomFieldsSchema)
	customFieldsLoader := gojsonschema.NewGoLoader(customFields)

	result, err := gojsonschema.Validate(schemaLoader, customFieldsLoader)
	if err != nil {
		return false, []gojsonschema.ResultError{}, err
	}

	if result.Valid() {
		return true, result.Errors(), nil
	} else {
		return false, result.Errors(), nil
	}
}

func UpdateExistingEvent(ctx context.Context, id string, updates map[string]interface{}) error {
	UPDATABLE_KEYS := map[string]bool{
		"name":                 true,
		"description":          true,
		"img_urls":             true,
		"location":             true,
		"address":              true,
		"start_timestamp":      true,
		"end_timestamp":        true,
		"custom_fields_schema": false, // Not allowed because since a ticket might exist with only old attributes
	}

	// Get event to get the custom field schema
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Convert the string/interface map to BSON updates
	bsonUpdates := bson.D{}
	for key, val := range updates {
		// Don't allow other keys to be updated
		if !UPDATABLE_KEYS[key] {
			return ErrEditNotAllowed
		}

		// Convert timestamps to time.Time objects
		if key == "start_timestamp" || key == "end_timestamp" {
			// TODO: Validate if the start_timestamp is before end_timestamp, probably not necessary but may be helpful
			if timestampStr, ok := val.(string); ok {
				timestamp, err := time.Parse(time.RFC3339, timestampStr)
				if err == nil {
					bsonUpdates = append(bsonUpdates, bson.E{Key: key, Value: timestamp})
				} else {
					log.Warn().Err(err).Str("key", key).Msg("could not parse timestamp as RFC3339")
					return errors.Join(fmt.Errorf("could not parse timestamp as RFC3339"), err)
				}
			} else {
				log.Warn().Err(err).Str("key", key).Msg("could not parse timestamp as string")
				return errors.Join(fmt.Errorf("could not parse timestamp as string"), err)
			}
		} else {
			// Add the key/val pair in BSON
			bsonUpdates = append(bsonUpdates, bson.E{Key: key, Value: val})
		}
	}

	// Try to update document in DB
	res, err := lib.Datastore.Db.Collection(eventsColName).
		UpdateByID(ctx, objectID, bson.D{{Key: "$set", Value: bsonUpdates}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoDocumentModified
	}
	return nil
}

func DeleteEvent(ctx context.Context, id primitive.ObjectID) error {
	// Check if event exists
	exists, err := CheckIfEventExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	// Delete all tickets to event
	err = DeleteAllTicketsForEvent(ctx, id)
	if err != nil {
		return err
	}

	// Delete event
	res, err := lib.Datastore.Db.Collection(eventsColName).DeleteOne(ctx, bson.M{"_id": id})

	// Handle no document found
	if err == nil {
		if res.DeletedCount == 0 {
			err = ErrNotFound
		}
	}
	return err
}
