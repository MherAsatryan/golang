package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Promotion struct {
	ID             string
	Price          float64
	ExpirationDate time.Time
}

var (
	db *sql.DB
)

func main() {

	// Open a database connection
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(localhost:3306)/promotionsDB?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create promotions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS promotions (
			id TEXT(200) PRIMARY KEY,
			price REAL NOT NULL,
			expiration_date DATE NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// _, err = db.Exec("DELETE FROM promotions")
	// if err != nil {
	// 	panic(err)
	// }

	// Read CSV file
	err1 := readCSVFile("promotions.csv")
	if err1 != nil {
		log.Fatal(err1)
	}

	// Start a background goroutine to read the CSV file every 30 minutes
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			err := readCSVFile("promotions.csv")
			if err != nil {
				log.Println(err)
			}
		}
	}()

	// Set up the HTTP endpoint
	router := mux.NewRouter()
	router.HandleFunc("/promotions/{id}", getPromotionByID).Methods("GET")
	log.Fatal(http.ListenAndServe(":1321", router))

}

// HTTP handler function to retrieve promotion by ID
func getPromotionByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var promotion Promotion

	// Query the database
	res, err := db.Query("SELECT * FROM promotions WHERE id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Close()

	if res.Next() {
		err2 := res.Scan(&promotion.ID, &promotion.Price, &promotion.ExpirationDate)
		if err2 != nil {
			log.Fatal(err2)
			return
		}
	} else {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode(promotion)
}

func worker(db *sql.DB, jobs <-chan []string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Prepare the SQL statement
	stmt, err := db.Prepare("INSERT INTO `promotions`(`id`, `price`, `expiration_date`) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	// Insert data into database
	for record := range jobs {
		if len(record) < 3 {
			continue
		}

		price, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Println("Failed to parse price:", err)
			continue
		}

		// skip the part "+0200 CEST" from expirationDate
		expirationDate, err := time.Parse("2006-01-02 15:04:05", record[2][:19])
		if err != nil {
			log.Println("Failed to parse expiration date:", err)
			continue
		}

		promotion := Promotion{
			ID:             record[0],
			Price:          price,
			ExpirationDate: expirationDate,
		}

		_, err = stmt.Exec(promotion.ID, promotion.Price, promotion.ExpirationDate)
		if err != nil {
			log.Println(err)
		}
	}
}

// Function to read records from CSV file
func readCSVFile(filename string) error {
	log.Println("Start storing records into db")

	// Create a CSV reader
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Create a worker pool
	numWorkers := 10
	wg := sync.WaitGroup{}
	wg.Add(numWorkers)
	jobs := make(chan []string, 100)
	for i := 0; i < numWorkers; i++ {
		go worker(db, jobs, &wg)
	}

	// Parse CSV data
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				log.Fatal(err)
			}
		}
		jobs <- record
	}
	close(jobs)

	// Wait for workers to complete
	wg.Wait()
	log.Println("Finish storing records into db")
	return nil
}
