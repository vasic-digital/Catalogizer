package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func testAuth() {
	// Test registration
	registerData := map[string]interface{}{
		"username":   "testuser",
		"email":      "test@example.com",
		"password":   "testpassword123",
		"first_name": "Test",
		"last_name":  "User",
	}

	jsonData, _ := json.Marshal(registerData)

	resp, err := http.Post("http://localhost:8080/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Register Response: %s\n", string(body))

	// Test login
	loginData := map[string]interface{}{
		"username": "testuser",
		"password": "testpassword123",
	}

	loginJson, _ := json.Marshal(loginData)

	resp2, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJson))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	fmt.Printf("Login Response: %s\n", string(body2))

	// Test protected endpoint
	resp3, err := http.Get("http://localhost:8080/api/v1/catalog")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp3.Body.Close()

	body3, _ := io.ReadAll(resp3.Body)
	fmt.Printf("Protected Endpoint (no auth): %s\n", string(body3))
}
