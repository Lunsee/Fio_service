package main

import (
	_ "fio_service/docs"
	db "fio_service/internal/database"
	"fio_service/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title FIO API
// @version 1.0
// @description Test task : This is a simple FIO API to manage Persons
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

func main() {

	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // default port - 8080
	}

	//db
	db.ConnectToPostgres()
	//routes
	r := mux.NewRouter()

	r.HandleFunc("/AddPerson", routes.AddPerson).Methods("POST")
	log.Println("Route registered: /AddPerson (POST)")

	r.HandleFunc("/DeletePerson/{id}", routes.DeletePerson).Methods("DELETE")
	log.Println("Route registered: /DeletePerson/{id} (DELETE)")

	r.HandleFunc("/EditPerson", routes.EditPerson).Methods("PUT")
	log.Println("Route registered: /EditPerson (PUT)")

	r.HandleFunc("/GetPerson", routes.GetPerson).Methods("GET")
	log.Println("Route registered: /GetPerson (GET)")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	log.Println("Swagger route registered: /swagger/")

	log.Printf("Server started on :%s", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
