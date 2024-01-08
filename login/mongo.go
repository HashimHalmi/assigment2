package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	// MongoDB connection
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/LoginFormPractice")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB connected")

	collection = client.Database("LoginFormPractice").Collection("users")
	fmt.Println("Users collection created")

	// HTTP server with Gorilla Mux
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	r.HandleFunc("/signup", signupHandler).Methods("GET", "POST")
	r.HandleFunc("/", loginHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/delete-account", deleteAccountHandler).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":3000", nil)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		signupPostHandler(w, r)
		return
	}
	http.ServeFile(w, r, "signup.html")
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "login.html")
}

func signupPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	lname := r.FormValue("lname")
	email := r.FormValue("email")
	password := r.FormValue("password")
	cpassword := r.FormValue("cpassword")

	// Check if user already exists in MongoDB collection
	filter := bson.D{{Key: "name", Value: name}}
	var existingUser struct{ Name string }
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)

	if err == nil {
		// User already exists, return error message
		fmt.Fprintf(w, "User '%s' already exists", name)
		return
	}

	// Check if the password and confirm password match
	if password != cpassword {
		fmt.Fprintf(w, "Password and Confirm Password do not match")
		return
	}

	// User does not exist, create new user
	newUser := bson.D{
		{Key: "name", Value: name},
		{Key: "lname", Value: lname},
		{Key: "email", Value: email},
		{Key: "password", Value: password},
	}
	_, err = collection.InsertOne(context.Background(), newUser)

	if err != nil {
		fmt.Fprintf(w, "Error creating user: %v", err)
		return
	}

	fmt.Fprintf(w, "User '%s' created successfully", name)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Check if user exists in MongoDB collection and if the password is correct
	filter := bson.D{{Key: "email", Value: email}, {Key: "password", Value: password}}
	var existingUser struct{ email string }
	err := collection.FindOne(context.Background(), filter).Decode(&existingUser)

	if err != nil {
		// User not found or incorrect password
		fmt.Fprintf(w, "Incorrect password or user not found in MongoDB")
		return
	}

	// User found, password correct
	// Render home page or perform other actions
	fmt.Fprintf(w, "Welcome, %s!", email)
}
func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Additional security measures such as password validation should be added here

		// Delete user from MongoDB collection
		filter := bson.D{{Key: "email", Value: email}, {Key: "password", Value: password}}
		result, err := collection.DeleteOne(context.Background(), filter)

		if err != nil {
			// Handle error
			fmt.Fprintf(w, "Error deleting account: %v", err)
			return
		}

		if result.DeletedCount == 0 {
			// User not found
			fmt.Fprintf(w, "User not found for deletion")
			return
		}

		// Account deleted successfully
		fmt.Fprintf(w, "Account deleted successfully")
	} else {
		// Handle unsupported HTTP methods
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
