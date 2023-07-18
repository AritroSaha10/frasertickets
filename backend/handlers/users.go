package handlers

import (
	"net/http"

	"github.com/aritrosaha10/frasertickets/lib"
	"github.com/aritrosaha10/frasertickets/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Admin         bool                 `json:"admin" bson:"admin"`
	StudentNumber string               `json:"student_number" bson:"student_number"`
	FirstName     string               `json:"first_name" bson:"first_name"`
	LastName      string               `json:"last_name" bson:"last_name"`
	ProfilePicURL string               `json:"pfp_url" bson:"pfp_url"`
	TicketsOwned  []primitive.ObjectID `json:"tickets_owned" bson:"tickets_owned"`
}

func (user *User) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type UsersResource struct{}

func (rs UsersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)    // GET /users - returns list of users, only available to admins
	r.Post("/", rs.Create) // POST /users - add new user to database, only run during sign up process

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", rs.Get)      // GET /users/{id} - returns user data, only available to admins and user
		r.Patch("/", rs.Update) // PATCH /users/{id} - updates user data, only available to admins and user
	})

	return r
}

func (rs UsersResource) List(w http.ResponseWriter, r *http.Request) {
	cursor, err := lib.Datastore.Db.Collection("users").Find(r.Context(), bson.D{})
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Attempt to convert MongoDB data into Go data
	var results []User
	if err := cursor.All(r.Context(), &results); err != nil {
		render.Render(w, r, util.ErrServer(err))
		return
	}

	// Convert into list of renderers to turn into JSON
	list := []render.Renderer{}
	for _, user := range results {
		list = append(list, &user)
	}

	// Return as JSON array, fallback if it fails
	if err := render.RenderList(w, r, list); err != nil {
		render.Render(w, r, util.ErrRender(err))
		return
	}
}

func (rs UsersResource) Create(w http.ResponseWriter, r *http.Request) {
	// Create fake user for now
	// TODO: Fill in using info from token given
	tmpUser := User{
		Admin:         false,
		StudentNumber: "709121",
		FirstName:     "Arnab",
		LastName:      "Saha",
		ProfilePicURL: "https://www.pngall.com/wp-content/uploads/5/Linux-PNG-Photo.png",
		TicketsOwned:  []primitive.ObjectID{},
	}

	// Create the user doc in MongoDB
	res, err := lib.Datastore.Db.Collection("users").InsertOne(r.Context(), tmpUser)
	if err != nil {
		render.Render(w, r, util.ErrServer(err))
		return
	}
	tmpUser.ID = res.InsertedID.(primitive.ObjectID)

	// Return as JSON, fallback if it fails
	if err := render.Render(w, r, &tmpUser); err != nil {
		render.Render(w, r, util.ErrRender(err))
		return
	}
}

func (rs UsersResource) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("..."))
}

func (rs UsersResource) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("..."))
}
