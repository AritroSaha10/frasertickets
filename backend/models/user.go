package models

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/aritrosaha10/frasertickets/lib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID            string `json:"id"             bson:"_id,omitempty"` // This is also the UUID in Firebase Auth
	Admin         bool   `json:"admin"          bson:"admin"`
	SuperAdmin    bool   `json:"superadmin"     bson:"superadmin"`
	StudentNumber string `json:"student_number" bson:"student_number"`
	FullName      string `json:"full_name"      bson:"full_name"`
	ProfilePicURL string `json:"pfp_url"        bson:"pfp_url"`
}

func (user *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func CreateUserIndices(ctx context.Context) error {
	// Create appropriate indices
	studentNumberIdxModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "student_number", Value: 1},
		},
	}

	// Try creating the indices
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := lib.Datastore.Db.Collection(usersColName).
		Indexes().
		CreateMany(
			ctx,
			[]mongo.IndexModel{
				studentNumberIdxModel,
			},
			opts,
		)

	return err
}

func GetAllUsers(ctx context.Context) ([]User, error) {
	// Try to get data from MongoDB
	cursor, err := lib.Datastore.Db.Collection(usersColName).Find(ctx, bson.D{})
	if err != nil {
		return []User{}, err
	}

	// Attempt to convert BSON data into User structs
	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return []User{}, err
	}

	return users, nil
}

func GetUserByKey(ctx context.Context, key string, value string) (User, error) {
	// Try to fetch data from DB
	var user User
	err := lib.Datastore.Db.Collection(usersColName).FindOne(ctx, bson.M{key: value}).Decode(&user)

	// No error handling needed (user & err will default to empty struct / nil)
	return user, err
}

func CheckIfUserExists(ctx context.Context, uid string) (bool, error) {
	// Directly return results from DB
	count, err := lib.Datastore.Db.Collection(usersColName).CountDocuments(ctx, bson.M{"_id": uid})
	return count == 1, err
}

func CheckIfUserWithStudentNumberExists(ctx context.Context, studentNumber string) (bool, error) {
	// Directly return results from DB
	count, err := lib.Datastore.Db.Collection(usersColName).CountDocuments(ctx, bson.M{"student_number": studentNumber})
	return count == 1, err
}

func CreateNewUser(ctx context.Context, user User) (string, error) {
	// Try to add document
	res, err := lib.Datastore.Db.Collection(usersColName).InsertOne(ctx, user)

	// Return object ID
	return res.InsertedID.(string), err
}

func UpdateExistingUserByStruct(ctx context.Context, user User, fieldsToUpdate User) (User, error) {
	// Confirm that user object exists in database
	_, err := GetUserByKey(ctx, "_id", user.ID)
	if err != nil {
		return User{}, err
	}

	// Check whether fields to update includes Object ID (which cannot be changed)
	if fieldsToUpdate.ID != "" {
		return User{}, fmt.Errorf("cannot update object id of user")
	}

	// Update user object with new values from fieldsToUpdate using reflection
	fieldsToUpdateValue := reflect.ValueOf(fieldsToUpdate)
	userValue := reflect.ValueOf(user)
	for i := 0; i < fieldsToUpdateValue.NumField(); i++ {
		userValue.Field(i).Set(fieldsToUpdateValue.Field(i))
	}
	newUser := userValue.Interface().(User)

	// Run replace operation
	_, err = lib.Datastore.Db.Collection(usersColName).
		ReplaceOne(ctx, bson.D{{Key: "_id", Value: user.ID}}, newUser)

	// Error checking
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func UpdateExistingUserByKeys(
	ctx context.Context,
	id string,
	updates map[string]interface{},
) error {
	// TODO: Somehow enforce schema during these updates, especially since users can run this function
	// to limited extent
	UPDATABLE_KEYS := map[string]bool{
		"admin":          true,
		"superadmin":     true,
		"student_number": true,
		"full_name":      true,
		"pfp_url":        true,
	}

	// Convert the string/interface map to BSON updates
	bsonUpdates := bson.D{}
	for key, val := range updates {
		// Don't allow other keys to be updated
		if !UPDATABLE_KEYS[key] {
			return ErrEditNotAllowed
		}

		// Add the key/val pair in BSON
		bsonUpdates = append(bsonUpdates, bson.E{Key: key, Value: val})
	}

	// Try to update document in DB
	res, err := lib.Datastore.Db.Collection(usersColName).
		UpdateByID(ctx, id, bson.D{{Key: "$set", Value: bsonUpdates}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoDocumentModified
	}
	return nil
}
