//This is the Demo Jam API that includes an in memory database.
//This API has all of the CRUD functionality


package main


import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Item struct (Model)
// This represents the data we're working with.
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// In-memory "database"
var (
	items     []Item
	itemsLock sync.Mutex // Mutex to make our slice-based DB thread-safe
)

// respondWithError is a helper function for sending JSON error messages
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON is a helper function for sending JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to marshal JSON response")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// --- Handler Functions ---

// getItems (GET /items)
// This retrieves the full list of items.
func getItems(w http.ResponseWriter, r *http.Request) {
	itemsLock.Lock()
	defer itemsLock.Unlock()

	respondWithJSON(w, http.StatusOK, items)
}

// getItem (GET /items/{id})
// This retrieves a single item by its ID.
func getItem(w http.ResponseWriter, r *http.Request) {
	itemsLock.Lock()
	defer itemsLock.Unlock()

	params := mux.Vars(r) // Get URL parameters
	id := params["id"]

	for _, item := range items {
		if item.ID == id {
			respondWithJSON(w, http.StatusOK, item)
			return
		}
	}
	respondWithError(w, http.StatusNotFound, "Item not found")
}

// createItem (POST /items)
// This covers your "add" and "post" request. It creates a new item.
func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	itemsLock.Lock()
	defer itemsLock.Unlock()

	// Simple ID generation (in a real app, use UUIDs or database serials)
	item.ID = strconv.Itoa(rand.Intn(1000000))
	items = append(items, item)

	respondWithJSON(w, http.StatusCreated, item)
}

// updateItem (PUT /items/{id})
// This covers your "update" request. It modifies an existing item.
func updateItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var updatedItem Item
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedItem); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	itemsLock.Lock()
	defer itemsLock.Unlock()

	for index, item := range items {
		if item.ID == id {
			// Found the item, now update it
			items[index].Name = updatedItem.Name
			items[index].Description = updatedItem.Description
			// Note: We keep the original ID
			respondWithJSON(w, http.StatusOK, items[index])
			return
		}
	}

	respondWithError(w, http.StatusNotFound, "Item not found")
}

// deleteItem (DELETE /items/{id})
// This covers your "delete" request.
func deleteItem(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	itemsLock.Lock()
	defer itemsLock.Unlock()

	for index, item := range items {
		if item.ID == id {
			// Remove the item from the slice
			// This syntax means "append everything before this index...
			// with everything after this index."
			items = append(items[:index], items[index+1:]...)
			respondWithJSON(w, http.StatusOK, map[string]string{"result": "success", "id_deleted": id})
			return
		}
	}

	respondWithError(w, http.StatusNotFound, "Item not found")
}

// --- Main Function ---

func main() {
	// Initialize the router
	r := mux.NewRouter()

	// Add some mock data
	items = append(items, Item{ID: "1", Name: "Default Item 1", Description: "This is the first item"})
	items = append(items, Item{ID: "2", Name: "Default Item 2", Description: "This is the second item"})
	items = append(items, Item{ID: "3", Name: "Default Item 3", Description: "This is the third item"})
	items = append(items, Item{ID: "4", Name: "Default Item 4", Description: "This is the fourth item"})
	items = append(items, Item{ID: "5", Name: "Default Item 5", Description: "This is the fifth item"})

	// Define API endpoints and map them to handler functions
	// Your "get" functions
	r.HandleFunc("/items", getItems).Methods("GET")
	r.HandleFunc("/items/{id}", getItem).Methods("GET")

	// Your "add" / "post" function
	r.HandleFunc("/items", createItem).Methods("POST")

	// Your "update" function
	r.HandleFunc("/items/{id}", updateItem).Methods("PUT")

	// Your "delete" function
	r.HandleFunc("/items/{id}", deleteItem).Methods("DELETE")

	// Start the server
	log.Println("🚀 Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
