package models

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aritrosaha10/frasertickets/lib"
	"github.com/aritrosaha10/frasertickets/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Ticket struct {
	ID                primitive.ObjectID     `json:"id"        bson:"_id,omitempty"`
	Owner             string                 `json:"ownerID"   bson:"owner"` // owner ID
	OwnerData         User                   `json:"ownerData" bson:"ownerData"`
	Event             primitive.ObjectID     `json:"eventID"   bson:"event"`
	EventData         Event                  `json:"eventData" bson:"eventData"`
	Timestamp         time.Time              `json:"timestamp" bson:"timestamp"`
	ScanCount         int                    `json:"scanCount" bson:"scanCount"`
	LastScanTimestamp time.Time              `json:"lastScanTime" bson:"lastScanTime"`
	MaxScanCount      int                    `json:"maxScanCount" bson:"maxScanCount"`
	CustomFields      map[string]interface{} `json:"customFields" bson:"customFields"`
}

func (ticket *Ticket) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type TicketScan struct {
	Index           int       `json:"index"`
	Timestamp       time.Time `json:"timestamp"`
	TicketData      Ticket    `json:"ticketData"`
	UserData        User      `json:"userData"`
	Processed       bool      `json:"processed"`
	NoProcessReason string    `json:"noProcessReason"`
}

func (scan *TicketScan) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CreateTicketIndices(ctx context.Context) error {
	// Create appropriate indices
	eventOwnerPairIdxModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "event", Value: 1},
			{Key: "owner", Value: 1},
		},
	}
	eventIdxModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "event", Value: 1},
		},
	}
	ownerIdxModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "owner", Value: 1},
		},
	}

	// Try creating the indices
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := lib.Datastore.Db.Collection(ticketsColName).
		Indexes().
		CreateMany(
			ctx,
			[]mongo.IndexModel{
				eventOwnerPairIdxModel,
				eventIdxModel,
				ownerIdxModel,
			},
			opts,
		)

	return err
}

func GetTickets(ctx context.Context, filter bson.M) ([]Ticket, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: filter},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "events"},
				{Key: "localField", Value: "event"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "eventData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$eventData"},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "owner"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "ownerData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$ownerData"},
		},
	}

	// Try to get data from MongoDB
	cursor, err := lib.Datastore.Db.Collection(ticketsColName).Aggregate(ctx, pipeline)
	if err != nil {
		return []Ticket{}, err
	}
	defer cursor.Close(ctx)

	// Attempt to convert BSON data into Ticket structs
	var tickets []Ticket
	if err := cursor.All(ctx, &tickets); err != nil {
		return []Ticket{}, err
	}

	return tickets, nil
}

func GetTicketCount(ctx context.Context, filter bson.M) (int64, error) {
	// Try to get data from MongoDB
	count, err := lib.Datastore.Db.Collection(ticketsColName).CountDocuments(ctx, filter)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func SearchForTicket(
	ctx context.Context,
	eventID primitive.ObjectID,
	userID string,
) (Ticket, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.M{"event": eventID, "owner": userID}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "events"},
				{Key: "localField", Value: "event"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "eventData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$eventData"},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "owner"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "ownerData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$ownerData"},
		},
	}

	// Try to get data from DB
	cursor, err := lib.Datastore.Db.Collection(ticketsColName).Aggregate(ctx, pipeline)
	if err != nil {
		return Ticket{}, err
	}
	defer cursor.Close(ctx)

	// Attempt to convert BSON data into Ticket structs
	var ticket Ticket
	if nextExists := cursor.Next(ctx); !nextExists {
		return Ticket{}, mongo.ErrNoDocuments
	}

	if err := cursor.Decode(&ticket); err != nil {
		return Ticket{}, err
	}

	// No error handling needed (ticket & err default to empty struct / nil)
	return ticket, err
}

func GetTicket(ctx context.Context, id primitive.ObjectID) (Ticket, error) {
	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.M{"_id": id}},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "events"},
				{Key: "localField", Value: "event"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "eventData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$eventData"},
		},
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "users"},
				{Key: "localField", Value: "owner"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "ownerData"},
			},
			},
		},
		{
			{Key: "$unwind", Value: "$ownerData"},
		},
	}

	// Try to get data from DB
	cursor, err := lib.Datastore.Db.Collection(ticketsColName).Aggregate(ctx, pipeline)
	if err != nil {
		return Ticket{}, err
	}
	defer cursor.Close(ctx)

	// Attempt to convert BSON data into Ticket structs
	var ticket Ticket
	if nextExists := cursor.Next(ctx); !nextExists {
		return Ticket{}, mongo.ErrNoDocuments
	}

	if err := cursor.Decode(&ticket); err != nil {
		return Ticket{}, err
	}

	// No error handling needed (ticket & err default to empty struct / nil)
	return ticket, err
}

func CheckIfTicketExists(ctx context.Context, filter bson.M) (bool, error) {
	// Directly return results from DB
	count, err := lib.Datastore.Db.Collection(ticketsColName).CountDocuments(ctx, filter)
	return count == 1, err
}

func CreateNewTicket(ctx context.Context, ticket Ticket) (primitive.ObjectID, error) {
	// Set timestamp to now
	ticket.Timestamp = time.Now()

	// Check if ticket already exists
	_, err := SearchForTicket(ctx, ticket.Event, ticket.Owner)
	if err == nil {
		return primitive.NilObjectID, ErrAlreadyExists
	}

	// Get event if it exists
	event, err := GetEvent(ctx, bson.M{"_id": ticket.Event})
	if err == mongo.ErrNoDocuments {
		return primitive.NilObjectID, ErrNotFound
	} else if err != nil {
		return primitive.NilObjectID, err
	}

	// Check if user exists
	userExists, err := CheckIfUserExists(ctx, ticket.Owner)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if !userExists {
		return primitive.NilObjectID, ErrNotFound
	}

	// Check if ticket's custom data matches schema
	valid, schemaErrs, err := ValidateCustomEventFields(ctx, event, ticket.CustomFields)
	if err != nil {
		return primitive.NilObjectID, err
	}
	if !valid {
		errStr := ""
		for _, schemaErr := range schemaErrs {
			errStr += schemaErr.String() + "\n"
		}
		return primitive.NilObjectID, fmt.Errorf(errStr)
	}

	// Try to add ticket
	res, err := lib.Datastore.Db.Collection(ticketsColName).InsertOne(ctx, ticket)
	if err != nil {
		return primitive.NilObjectID, err
	}

	// Return object ID
	return res.InsertedID.(primitive.ObjectID), err
}

func UpdateExistingTicketByKeys(
	ctx context.Context,
	id primitive.ObjectID,
	updates map[string]interface{},
) error {
	// TODO: Somehow enforce schema during these updates
	UPDATABLE_KEYS := map[string]bool{
		"scanCount":    true,
		"lastScanTime": true,
		"maxScanCount": true,
	}
	CUSTOM_UPDATABLE_KEYS := map[string]bool{}

	// TODO: Enforce optional data schema
	ticket, err := GetTicket(ctx, id)
	if err != nil {
		return err
	}

	// Convert schema to struct format
	customFieldSchema, err := util.ConvertRawCustomFieldsSchema(ticket.EventData.RawCustomFieldsSchema)
	if err != nil {
		return err
	}
	// Add all the new properties
	for key, property := range customFieldSchema.Properties {
		if property.Editable {
			CUSTOM_UPDATABLE_KEYS[key] = true
		}
	}

	// Convert the string/interface map to BSON updates
	bsonUpdates := bson.D{}
	for key, val := range updates {
		tmpKey := key // To allow for manipulation for custom fields
		// Don't allow other keys to be updated
		if !UPDATABLE_KEYS[key] {
			if CUSTOM_UPDATABLE_KEYS[key] {
				// TODO: Validate this value against the schema, already done in frontend but best to keep backup
				// Adjust key so that it updates under customFields object
				tmpKey = "customFields." + key
			} else {
				return ErrEditNotAllowed
			}
		}

		// Add the key/val pair in BSON
		bsonUpdates = append(bsonUpdates, bson.E{Key: tmpKey, Value: val})
	}

	// Try to update document in DB
	res, err := lib.Datastore.Db.Collection(ticketsColName).
		UpdateByID(ctx, id, bson.D{{Key: "$set", Value: bsonUpdates}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoDocumentModified
	}
	return nil
}

func DeleteTicket(ctx context.Context, id primitive.ObjectID) error {
	// Delete ticket
	res, err := lib.Datastore.Db.Collection(ticketsColName).DeleteOne(ctx, bson.M{"_id": id})

	// Handle no document found
	if err == nil {
		if res.DeletedCount == 0 {
			err = ErrNoDocumentModified
		}
	}
	return err
}

func DeleteAllTicketsForEvent(ctx context.Context, eventID primitive.ObjectID) error {
	// Delete all tickets to event
	_, err := lib.Datastore.Db.Collection(ticketsColName).DeleteMany(ctx, bson.M{"event": eventID})
	return err
}
