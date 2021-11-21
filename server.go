package main

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func (customers *Customers) getAllCustomer(context echo.Context) error {
	client, ctx, customerCollection := getDB()
	defer client.Disconnect(ctx)

	cursor, err := customerCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &customers.customers); err != nil {
		log.Fatal(err)
	}
	if len(customers.customers) == 0 {
		return context.JSON(http.StatusNotFound, "No customers available!")
	}
	return context.JSON(http.StatusOK, customers.customers)
}

func (customer *Customer) getCustomer(context echo.Context) error {
	id := context.Param("id")

	client, ctx, customerCollection := getDB()
	defer client.Disconnect(ctx)

	if err := customerCollection.FindOne(ctx, bson.M{"id": id}).Decode(&customer); err != nil {
		return context.JSON(http.StatusNotFound, "Customer not found!")
	}
	return context.JSON(http.StatusOK, customer)
}

func (customer *Customer) updateCustomer(context echo.Context) error {
	type body struct {
		Name    string `json:"Name" validate:"required"`
		Address string `json:"Address" validate:"required"`
	}
	var reqBody body
	if err := context.Bind(&reqBody); err != nil {
		return err
	}
	if err := validator.New().Struct(reqBody); err != nil {
		return context.JSON(http.StatusNotFound, "invalid format!")
	}

	id := context.Param("id")
	client, ctx, customerCollection := getDB()
	defer client.Disconnect(ctx)

	filter := bson.M{"id": id}
	update := bson.D{
		{"$set", bson.D{
			{"name", reqBody.Name},
			{"address", reqBody.Address},
		},
		},
	}
	result, err := customerCollection.UpdateOne(
		ctx,
		filter,
		update,
	)
	if err != nil {
		log.Fatal(err)
	}

	if result.MatchedCount == 0 {
		return context.JSON(http.StatusNotFound, "Customer not found!")
	}
	fmt.Printf("Updated %v Document!\n", result.ModifiedCount)
	return context.JSON(http.StatusOK, "Customer updated successfully")
}

func (customer *Customer) deleteCustomer(context echo.Context) error {
	id := context.Param("id")

	client, ctx, customerCollection := getDB()
	defer client.Disconnect(ctx)

	filter := bson.M{"id": id}
	result, err := customerCollection.DeleteOne(
		ctx,
		filter,
	)
	if err != nil {
		log.Fatal(err)
	}

	if result.DeletedCount == 0 {
		return context.JSON(http.StatusNotFound, "Customer Not Found!")
	}
	fmt.Printf("Deleted %v Document!\n", result.DeletedCount)
	return context.JSON(http.StatusOK, "Customer Deleted successfully")
}

func validateId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		if _, err := strconv.Atoi(context.Param("id")); err != nil {
			return context.JSON(http.StatusNotFound, "id must be neumeric")
		}
		return next(context)
	}
}

func (customer Customer) addCustomer(context echo.Context) error {
	type body struct {
		Name    string `json:"Name" validate:"required"`
		Address string `json:"Address" validate:"required"`
	}
	var reqBody body
	if err := context.Bind(&reqBody); err != nil {
		return err
	}

	if err := validator.New().Struct(reqBody); err != nil {
		return context.JSON(http.StatusNotFound, "invalid format!")
	}

	client, ctx, customerCollection := getDB()
	defer client.Disconnect(ctx)

	cursor, err := customerCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	max := 0
	for cursor.Next(ctx) {
		if err = cursor.Decode(&customer); err != nil {
			log.Fatal(err)
		}
		if id, _ := strconv.Atoi(customer.Id); id > max {
			max = id
		}
	}
	max++
	id := strconv.Itoa(max)
	customer = Customer{
		Id:      id,
		Name:    reqBody.Name,
		Address: reqBody.Address,
	}

	insertedResult, err := customerCollection.InsertOne(ctx, customer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %v\n", insertedResult.InsertedID)
	return context.JSON(http.StatusOK, customer)
}

func getDB() (*mongo.Client, context.Context, *mongo.Collection) {
	uri := os.Getenv("URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	customerDatabase := client.Database("customerDB")
	customerCollection := customerDatabase.Collection("customer")

	return client, ctx, customerCollection
}

type Customer struct {
	object_id primitive.ObjectID `bson:"_id,omitempty"`
	Id        string             `json:"id,omitempty" bson:"id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name,omitempty"`
	Address   string             `json:"address,omitempty" bson:"address,omitempty"`
}

type Customers struct {
	customers []Customer
}

func main() {

	godotenv.Load()
	// Getting port number from the envrionment variable
	port := os.Getenv("CUSTOMERS_PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/api/v1/customers", new(Customers).getAllCustomer)
	e.GET("/api/v1/customers/:id", new(Customer).getCustomer, validateId)
	e.PUT("/api/v1/customers/:id", new(Customer).updateCustomer, validateId)
	e.DELETE("/api/v1/customers/:id", new(Customer).deleteCustomer, validateId)
	e.POST("/api/v1/customers", new(Customer).addCustomer)

	e.Logger.Print(fmt.Sprintf("Listening to the port: %s", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf("localhost:%s", port)))
}
