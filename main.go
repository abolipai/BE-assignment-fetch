package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)


// <--------- Structs for API calls ---------->
// Request payload for /add
type Transaction struct {
	Payer     string    `json:"payer"`
	Points    int       `json:"points"`
	Timestamp time.Time `json:"timestamp"`
}

// Request payload for /spend
type SpendRequest struct {
	Points int `json:"points"`
}

// Response paylaod for /spend
type SpentPoints struct {
	Payer  string `json:"payer"`
	Points int    `json:"points"`
}

var (
	transactions []Transaction
	balanceMap    map[string]int
)

// Function for the add points endpoint
func addPointsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST call is supported", http.StatusMethodNotAllowed)
		return
	}
	// Parse the JSON request body into a Transaction struct
	var transaction Transaction
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the transaction to the transactions slice
	transactions = append(transactions, transaction)

	// Sort the transactions by timestamp
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Timestamp.Before(transactions[j].Timestamp)
	})

	log.Printf("In ADD api, current transactions %v", transactions)

	balanceMap[transaction.Payer] += transaction.Points

	// Respond with a success message
	w.WriteHeader(http.StatusOK)
}

// Function for the balance endpoint
func balanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET call is supported", http.StatusMethodNotAllowed)
		return
	}
	// Respond with a map of points for each payer in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balanceMap)
}

// Function for the spend points endpoint
func spendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST call is supported", http.StatusMethodNotAllowed)
		return
	}
	var spendRequest SpendRequest
	err := json.NewDecoder(r.Body).Decode(&spendRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user has enough points to spend
	totalPoints := 0
	for _, points := range balanceMap {
		totalPoints += points
	}
	if spendRequest.Points > totalPoints {
		http.Error(w, "User doesn't have enough points", http.StatusBadRequest)
		return
	}

	// Map of points spent for each payer
	var spentPoints []SpentPoints
	spentMap := make(map[string]int)

	for transactionIndex, transaction := range transactions {

		if spendRequest.Points <= 0 {
			break
		}

        // Check total points of this payer and compare with the points in that transaction
		// This also takes care of the cases where the payer has dedcuted points (added negative points)
		minPoints := min(balanceMap[transaction.Payer], transaction.Points, spendRequest.Points)

		// Remove the possible points from the transaction and the balance map
		transactions[transactionIndex].Points = transactions[transactionIndex].Points - minPoints
		balanceMap[transaction.Payer] = balanceMap[transaction.Payer] - minPoints

		// Update the spend request points
		spendRequest.Points = spendRequest.Points - minPoints
		
		// Update map of points spent for each payer
		spentMap[transaction.Payer] = spentMap[transaction.Payer] - minPoints
    }

	for key, value := range spentMap {
		spentPoints = append(spentPoints, SpentPoints{Payer: key, Points: value})
	}

	log.Printf("In SPEND api, current spent map %v", spentMap)
	// Respond with the list of payer names and points spent
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spentPoints)
}

// Function to find the minimum of three integers
func min(x, y, z int) int {
    if x <= y && x <= z {
        return x
    } else if y <= x && y <= z {
        return y
    } else {
        return z
    }
}


func main() {
	// Add logs to a file
	logFile, err := os.OpenFile("transactions.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	// Initialize the transactions slice
	transactions = make([]Transaction, 0)
	balanceMap = make(map[string]int)

	http.HandleFunc("/add", addPointsHandler)
	http.HandleFunc("/balance", balanceHandler)
	http.HandleFunc("/spend", spendHandler)

	// Start the server on port 8080
	fmt.Println("Server is running on :8000")
	http.ListenAndServe(":8000", nil)
}
