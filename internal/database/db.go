package db

import (
	"fio_service/internal/models"
	"fmt"
	"log"
	"os"

	testData "fio_service/test"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Load .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("loading .env file to database successfully")

}

var db *gorm.DB

func ConnectToPostgres() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error

	// Загружаем временную зону
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Printf("Warning: couldn't load Moscow time zone: %v, falling back to UTC", err)
		location = time.UTC
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(location)
		},
	})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	log.Println("✅ Successfully connected to the database!")

	Migrate()
	LoadTestData()
	ResetPersonIDSequence()
}

func GetDB() *gorm.DB {
	return db
}

func Migrate() {
	if !db.Migrator().HasTable(&models.Persons{}) {
		err := db.AutoMigrate(&models.Persons{})
		if err != nil {
			log.Fatal(" Migration failed:", err)
		}
		log.Println("Database migrated successfully!")
	} else {
		log.Println("Info: Table already exists, skipping migration.")
	}
}

func LoadTestData() {
	//check
	var count int64
	err := db.Model(&models.Persons{}).Count(&count).Error
	if err != nil {
		log.Fatal("Failed to check existing data:", err)
	}

	if count > 0 {
		log.Println("Info: Test data already exists, skipping loading test data...")
		return
	}

	for _, song := range testData.TestPersons {
		err := db.Create(&song).Error
		if err != nil {
			log.Fatal("Failed to insert test data:", err)
		}
	}
	log.Println("Test data inserted successfully!")

}

func ResetPersonIDSequence() {
	var maxID int64
	err := db.Table("persons").Select("MAX(id)").Scan(&maxID).Error
	if err != nil {
		log.Printf("Warning: failed to get max ID: %v", err)
		return
	}

	err = db.Exec("SELECT setval('persons_id_seq', ?, false)", maxID+1).Error
	if err != nil {
		log.Printf("Warning: failed to reset sequence: %v", err)
	} else {
		log.Printf("Info: sequence persons_id_seq reset to %d", maxID+1)
	}
}
