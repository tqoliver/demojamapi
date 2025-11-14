package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// resetGlobalItems is a helper function to reset our in-memory DB before each test.
// This is crucial for making tests independent and repeatable.
func resetGlobalItems() {
	itemsLock.Lock()
	defer itemsLock.Unlock()
	// Clear and re-populate the slice with known mock data
	items = []Item{
		{ID: "1", Name: "Mock Item 1", Description: "First mock item"},
		{ID: "2", Name: "Mock Item 2", Description: "Second mock item"},
	}
}

// TestGetItems (GET /items)
func TestGetItems(t *testing.T) {
	resetGlobalItems()

	// Create a new HTTP request
	req := httptest.NewRequest("GET", "/items", nil)
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler directly
	getItems(rr, req)

	// --- Check results ---

	// 1. Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// 2. Check the response body
	var returnedItems []Item
	if err := json.NewDecoder(rr.Body).Decode(&returnedItems); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// 3. Check the data
	if len(returnedItems) != 2 {
		t.Errorf("handler returned unexpected number of items: got %d want %d",
			len(returnedItems), 2)
	}
}

// TestGetItem (GET /items/{id})
func TestGetItem(t *testing.T) {
	// Sub-test for "Item Found"
	t.Run("Item Found", func(t *testing.T) {
		resetGlobalItems()

		req := httptest.NewRequest("GET", "/items/1", nil)
		rr := httptest.NewRecorder()

		// **Key Step**: We must manually add the URL variables
		// that gorilla/mux would normally parse.
		vars := map[string]string{
			"id": "1",
		}
		req = mux.SetURLVars(req, vars)

		// Call the handler
		getItem(rr, req)

		// 1. Check status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// 2. Check body
		var item Item
		if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// 3. Check data
		if item.ID != "1" || item.Name != "Mock Item 1" {
			t.Errorf("handler returned wrong item: got %+v want item with ID 1", item)
		}
	})

	// Sub-test for "Item Not Found"
	t.Run("Item Not Found", func(t *testing.T) {
		resetGlobalItems()

		req := httptest.NewRequest("GET", "/items/999", nil)
		rr := httptest.NewRecorder()

		vars := map[string]string{
			"id": "999",
		}
		req = mux.SetURLVars(req, vars)

		getItem(rr, req)

		// 1. Check status code
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusNotFound)
		}
	})
}

// TestCreateItem (POST /items)
func TestCreateItem(t *testing.T) {
	// Sub-test for "Valid Payload"
	t.Run("Valid Payload", func(t *testing.T) {
		resetGlobalItems()
		initialLength := len(items)

		// Create our request body (JSON)
		payload := []byte(`{"name":"New Item", "description":"A new test item"}`)
		req := httptest.NewRequest("POST", "/items", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		createItem(rr, req)

		// 1. Check status code
		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusCreated)
		}

		// 2. Check response body
		var item Item
		if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if item.Name != "New Item" {
			t.Errorf("handler returned wrong item name: got %s want %s",
				item.Name, "New Item")
		}
		if item.ID == "" {
			t.Error("handler returned item with no ID")
		}

		// 3. Check global state (was it actually added?)
		itemsLock.Lock()
		if len(items) != initialLength+1 {
			t.Errorf("item was not added to the slice: got len %d want %d",
				len(items), initialLength+1)
		}
		itemsLock.Unlock()
	})

	// Sub-test for "Invalid Payload"
	t.Run("Invalid Payload", func(t *testing.T) {
		resetGlobalItems()
		initialLength := len(items)

		// Malformed JSON
		payload := []byte(`{"name":"Bad JSON", "description":}`)
		req := httptest.NewRequest("POST", "/items", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		createItem(rr, req)

		// 1. Check status code
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code for bad payload: got %v want %v",
				status, http.StatusBadRequest)
		}

		// 2. Check global state (should not have changed)
		itemsLock.Lock()
		if len(items) != initialLength {
			t.Errorf("slice length changed on bad request: got %d want %d",
				len(items), initialLength)
		}
		itemsLock.Unlock()
	})
}

// TestUpdateItem (PUT /items/{id})
func TestUpdateItem(t *testing.T) {
	// Sub-test for "Item Found"
	t.Run("Item Found", func(t *testing.T) {
		resetGlobalItems()

		payload := []byte(`{"name":"Updated Name", "description":"Updated Description"}`)
		req := httptest.NewRequest("PUT", "/items/1", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		vars := map[string]string{"id": "1"}
		req = mux.SetURLVars(req, vars)

		updateItem(rr, req)

		// 1. Check status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// 2. Check body
		var item Item
		json.NewDecoder(rr.Body).Decode(&item)
		if item.Name != "Updated Name" {
			t.Errorf("handler returned wrong updated name: got %s want %s",
				item.Name, "Updated Name")
		}
		// The ID should *not* change
		if item.ID != "1" {
			t.Error("handler changed the item ID during update")
		}

		// 3. Check global state
		itemsLock.Lock()
		if items[0].Name != "Updated Name" {
			t.Error("global state was not updated correctly")
		}
		itemsLock.Unlock()
	})

	// Sub-test for "Item Not Found"
	t.Run("Item Not Found", func(t *testing.T) {
		resetGlobalItems()

		payload := []byte(`{"name":"Updated Name", "description":"Updated Description"}`)
		req := httptest.NewRequest("PUT", "/items/999", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		vars := map[string]string{"id": "999"}
		req = mux.SetURLVars(req, vars)

		updateItem(rr, req)

		// 1. Check status
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusNotFound)
		}
	})
}

// TestDeleteItem (DELETE /items/{id})
func TestDeleteItem(t *testing.T) {
	// Sub-test for "Item Found"
	t.Run("Item Found", func(t *testing.T) {
		resetGlobalItems() // Starts with 2 items
		initialLength := len(items)

		req := httptest.NewRequest("DELETE", "/items/1", nil)
		rr := httptest.NewRecorder()

		vars := map[string]string{"id": "1"}
		req = mux.SetURLVars(req, vars)

		deleteItem(rr, req)

		// 1. Check status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// 2. Check global state
		itemsLock.Lock()
		if len(items) != initialLength-1 {
			t.Errorf("item was not removed from slice: got len %d want %d",
				len(items), initialLength-1)
		}
		// Check that the *correct* item was deleted
		if items[0].ID == "1" {
			t.Error("wrong item was deleted or item was not deleted")
		}
		itemsLock.Unlock()
	})

	// Sub-test for "Item Not Found"
	t.Run("Item Not Found", func(t *testing.T) {
		resetGlobalItems() // Reset state (2 items)
		initialLength := len(items)

		req := httptest.NewRequest("DELETE", "/items/999", nil)
		rr := httptest.NewRecorder()

		vars := map[string]string{"id": "999"}
		req = mux.SetURLVars(req, vars)

		deleteItem(rr, req)

		// 1. Check status
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusNotFound)
		}

		// 2. Check global state (should be unchanged)
		itemsLock.Lock()
		if len(items) != initialLength {
			t.Errorf("slice length changed on bad request: got %d want %d",
				len(items), initialLength)
		}
		itemsLock.Unlock()
	})
}
