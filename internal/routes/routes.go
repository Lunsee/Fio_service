package routes

import (
	"encoding/json"
	db "fio_service/internal/database"
	"fio_service/internal/models"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type PersonRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

// AddPerson creates a new person
// @Summary Create a new person
// @Description Add a new person with name, surname, patronymic, age, gender, and ethnicity
// @Accept  json
// @Produce  json
// @Param person body PersonRequest true "Person Details"
// @Success 201 {object} models.Persons "Created person"
// @Failure 400 {string} string "Invalid JSON or Missing Fields"
// @Failure 500 {string} string "Internal Server Error"
// @Router /persons [post]
func AddPerson(w http.ResponseWriter, r *http.Request) {

	log.Println("Info: <AddPerson> endpoint...")

	var newPerson PersonRequest
	err := json.NewDecoder(r.Body).Decode(&newPerson)
	if err != nil {
		log.Printf("Error: Failed to decode JSON body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if newPerson.Name == "" || newPerson.Surname == "" {
		log.Printf("Error: Missing required fields (name or surname)")
		http.Error(w, "Name and surname are required", http.StatusBadRequest)
		return
	}

	log.Printf("Debug: Incoming endpoint request body %+v", newPerson)

	//Возраст
	api_age_url := fmt.Sprintf("%s/?name=%s", os.Getenv("API_AGE_URL"),
		url.QueryEscape(newPerson.Name))
	ageResp, err := http.Get(api_age_url) // запрос
	if err != nil {
		log.Printf("Error: Age API request failed: %v", err)
		http.Error(w, "Age service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer ageResp.Body.Close()

	var ageData struct {
		Age int `json:"age"`
	}

	if err := json.NewDecoder(ageResp.Body).Decode(&ageData); err != nil {
		log.Printf("Error: Failed to parse age API response: %v", err)
		http.Error(w, "Invalid age data", http.StatusInternalServerError)
		return
	}
	log.Printf("Debug: ageData API request answer: %v", ageData.Age)

	//Пол
	api_gender_url := fmt.Sprintf("%s/?name=%s", os.Getenv("API_GENDER_URL"),
		url.QueryEscape(newPerson.Name))
	genderResp, err := http.Get(api_gender_url) // запрос
	if err != nil {
		log.Printf("Error: gender API request failed: %v", err)
		http.Error(w, "gender service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer genderResp.Body.Close()

	var genderData struct {
		Gender string `json:"gender"`
	}

	if err := json.NewDecoder(genderResp.Body).Decode(&genderData); err != nil {
		log.Printf("Error: Failed to parse gender API response: %v", err)
		http.Error(w, "Invalid gender data", http.StatusInternalServerError)
		return
	}
	log.Printf("Debug: genderData API request answer: %v", genderData.Gender)

	//Национальность
	api_ethnicity_url := fmt.Sprintf("%s/?name=%s", os.Getenv("API_ETHNICITY_URL"),
		url.QueryEscape(newPerson.Name))
	ethnicity_Resp, err := http.Get(api_ethnicity_url) // запрос
	if err != nil {
		log.Printf("Error: SEX API request failed: %v", err)
		http.Error(w, "SEX service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer ethnicity_Resp.Body.Close()

	var ethnicityData struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}

	if err := json.NewDecoder(ethnicity_Resp.Body).Decode(&ethnicityData); err != nil {
		log.Printf("Error: Failed to parse ethnicity API response: %v", err)
		http.Error(w, "Invalid ethnicity data", http.StatusInternalServerError)
		return
	}

	ethnicity := "unknown" // default
	maxProb := -1.0

	for _, c := range ethnicityData.Country { // Ищем национальность с наибольшим значением вероятности
		if c.Probability > maxProb {
			maxProb = c.Probability
			ethnicity = c.CountryID
		}
	}
	log.Printf("Debug: ethnicity API detected country with max probability: %s (%.3f)", ethnicity, maxProb)

	person := models.Persons{
		Name:       newPerson.Name,
		Surname:    newPerson.Surname,
		Patronymic: newPerson.Patronymic,
		Age:        ageData.Age,
		Gender:     genderData.Gender,
		Ethnicity:  ethnicity,
	}
	database := db.GetDB()

	result := database.Create(&person)
	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(person)
	log.Printf("Success: Created person with ID %d", person.ID)
}

// DeletePerson removes a person by ID
// @Summary Delete a person by ID
// @Description Delete a person using their ID
// @Param id path string true "Person ID"
// @Success 202 {object} models.Persons "Deleted person"
// @Failure 400 {string} string "Invalid ID"
// @Failure 404 {string} string "Person not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /persons/{id} [delete]
func DeletePerson(w http.ResponseWriter, r *http.Request) {

	log.Println("Info: <DeletePerson> endpoint...")
	vars := mux.Vars(r)
	personId := vars["id"]
	w.Write([]byte("Deleting person with ID: " + personId))

	// Конвертируем строку в uint
	id, err := strconv.ParseUint(personId, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	log.Printf("Debug: url id: %v", id)

	database := db.GetDB()

	var person models.Persons
	result := database.First(&person, id)
	if result.Error != nil {
		http.Error(w, "Not Found ", http.StatusNotFound)
		return
	}
	log.Printf("Debug: Found object from db %v with id: %v", person, id)

	result = database.Delete(&person)
	if result.Error != nil {
		http.Error(w, "Failed to delete person", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(person)
	log.Printf("Success: Deteled person with ID %d", person.ID)
}

// EditPerson updates a person's details
// @Summary Edit a person's details
// @Description Update the details of an existing person
// @Accept  json
// @Produce  json
// @Param person body models.Persons true "Updated Person Details"
// @Success 200 {object} models.Persons "Updated person"
// @Failure 400 {string} string "Invalid JSON"
// @Failure 404 {string} string "Person not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /persons [put]
func EditPerson(w http.ResponseWriter, r *http.Request) {

	log.Println("Info: <EditPerson> endpoint...")
	var newPerson models.Persons

	err := json.NewDecoder(r.Body).Decode(&newPerson)
	if err != nil {
		log.Printf("Error: Failed to decode JSON body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	log.Printf("Debug: Incoming endpoint request body %+v", newPerson)

	database := db.GetDB()

	var person models.Persons
	result := database.First(&person, newPerson.ID)
	if result.Error != nil {
		http.Error(w, "Person Not Found", http.StatusNotFound)
		return
	}
	log.Printf("Debug: Found person in DB %+v", person)

	person.Name = newPerson.Name
	person.Surname = newPerson.Surname
	person.Patronymic = newPerson.Patronymic
	person.Age = newPerson.Age
	person.Gender = newPerson.Gender
	person.Ethnicity = newPerson.Ethnicity
	person.UpdatedAt = time.Now()

	result = database.Save(&person)
	if result.Error != nil {
		http.Error(w, "Person was Not saved", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(person)
	log.Printf("Success: Updated person with ID %d", person.ID)
}

// GetPerson retrieves a list of people with optional filters
// @Summary Get a list of people
// @Description Get a list of people with optional query parameters such as page, limit, name, surname, age, gender, ethnicity
// @Produce  json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Limit number" default(10)
// @Param name query string false "Filter by name"
// @Param surname query string false "Filter by surname"
// @Param age query int false "Filter by age"
// @Param gender query string false "Filter by gender"
// @Param ethnicity query string false "Filter by ethnicity"
// @Success 200 {array} models.Persons "List of people"
// @Failure 400 {string} string "Invalid parameters"
// @Failure 500 {string} string "Database error"
// @Router /persons [get]
func GetPerson(w http.ResponseWriter, r *http.Request) {

	log.Println("Info: <GetPerson> endpoint...")
	queryParams := r.URL.Query()
	log.Printf("Debug: queryParams: %+v", queryParams)

	//default
	page := 1
	limit := 10

	//get page
	if pageStr := queryParams.Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			http.Error(w, "Invalid 'page' parameter", http.StatusBadRequest)
			return
		}
		page = p
	}
	log.Printf("Debug: Page value: %+v", page)

	//get limit
	if limitStr := queryParams.Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 {
			http.Error(w, "Invalid 'limit' parameter", http.StatusBadRequest)
			return
		}
		limit = l
	}
	log.Printf("Debug: limit value: %+v", limit)

	database := db.GetDB()
	query := database.Model(&models.Persons{})

	//Фильтрация
	if name := queryParams.Get("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%") //
	}
	if surname := queryParams.Get("surname"); surname != "" {
		query = query.Where("surname ILIKE ?", "%"+surname+"%")
	}
	// Фильтрация по возрасту
	if ageStr := queryParams.Get("age"); ageStr != "" {
		age, err := strconv.Atoi(ageStr)
		if err == nil {
			query = query.Where("age = ?", age)
		}
	}
	//по полу
	if gender := queryParams.Get("gender"); gender != "" {
		query = query.Where("gender = ?", gender)
	}
	// национальности
	if ethnicity := queryParams.Get("ethnicity"); ethnicity != "" {
		query = query.Where("ethnicity = ?", ethnicity)
	}

	offset := (page - 1) * limit              // кол во записей
	query = query.Offset(offset).Limit(limit) // запрос

	var persons []models.Persons
	if err := query.Find(&persons).Error; err != nil {
		log.Printf("Error: failed to fetch persons: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(persons)
}
