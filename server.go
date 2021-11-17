package main

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"strconv"
)
func getAllCustomer(context echo.Context) error {
	if len(customersDB) == 0 {
		return context.JSON(http.StatusNotFound, "No customers available!")
	}
	return context.JSON(http.StatusOK, customersDB)
}

func getCustomer(context echo.Context) error {
	id := context.Param("id")
	pop, ok := customersDB[id]
	if ok {
		return context.JSON(http.StatusOK, pop)
	}
	return context.JSON(http.StatusNotFound, "Customer not found!")
}

func updateCustomer(context echo.Context) error {
	type body struct {
		Name string			`json:"Name" validate:"required"`
		Address string		`json:"Address" validate:"required"`
	}
	var reqBody body
	if err := context.Bind(&reqBody); err != nil {
		return err
	}

	if err := validator.New().Struct(reqBody); err != nil {
		return context.JSON(http.StatusNotFound, "invalid format!")
	}

	id := context.Param("id")
	_, ok := customersDB[id]
	if ok {
		customersDB[id] = Customer{
			Id: id,
			Name: reqBody.Name,
			Address: reqBody.Address,
		}
		return context.JSON(http.StatusOK, customersDB[id])
	}
	return context.JSON(http.StatusNotFound, "Customer not found!")
}

func deleteCustomer(context echo.Context) error {
	id := context.Param("id")
	pop, ok := customersDB[id]
	if ok {
		delete(customersDB, id)
		return context.JSON(http.StatusOK, fmt.Sprintf("%v Deleted!", pop.Name))
	}
	return context.JSON(http.StatusNotFound, "Customer Not Found!")
}

func validateId(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		if _, err := strconv.Atoi(context.Param("id")); err != nil {
			return context.JSON(http.StatusNotFound, "id must be neumeric")
		}
		return next(context)
	}
}

func addCustomer(context echo.Context) error {
	type body struct {
		Name string			`json:"Name" validate:"required"`
		Address string		`json:"Address" validate:"required"`
	}
	var reqBody body
	if err := context.Bind(&reqBody); err != nil {
		return err
	}

	if err := validator.New().Struct(reqBody); err != nil {
		return context.JSON(http.StatusNotFound, "invalid format!")
	}

	id := strconv.Itoa(idCounter)
	idCounter++
	customersDB[id] = Customer{
		Id: id,
		Name: reqBody.Name,
		Address: reqBody.Address,
	}

	return context.JSON(http.StatusOK, customersDB[id])
}

type Customer struct {
	Id string
	Name string
	Address string
}

var customersDB map[string]Customer
var idCounter int

func main() {
	// Initializing customer database with two entries
	customersDB = map[string]Customer{
		"1" : {
			Id: "1",
			Name: "Niloy Deb Roy",
			Address: "Ja, 115/1, Mohakhali, Dhaka",
		},
		"2" : {
			Id: "2",
			Name: "Moeen Khan",
			Address: "66, Mohakhali, Dhaka",
		},
	}

	idCounter = len(customersDB) + 1

	// Getting port number from the envrionment variable
	port := os.Getenv("CUSTOMERS_PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/api/v1/customers", getAllCustomer)
	e.GET("/api/v1/customers/:id", getCustomer, validateId)
	e.PUT("/api/v1/customers/:id", updateCustomer, validateId)
	e.DELETE("/api/v1/customers/:id", deleteCustomer, validateId)
	e.POST("/api/v1/customers", addCustomer)

	e.Logger.Print(fmt.Sprintf("Listening to the port: %s", port))
	e.Logger.Fatal(e.Start(fmt.Sprintf("localhost:%s", port)))
}