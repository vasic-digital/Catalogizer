package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LoginRequest struct {
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	DeviceInfo DeviceInfo `json:"device_info,omitempty"`
	RememberMe bool      `json:"remember_me"`
}

type DeviceInfo struct {
	DeviceType      *string `json:"device_type,omitempty"`
	Platform        *string `json:"platform,omitempty"`
	PlatformVersion *string `json:"platform_version,omitempty"`
	AppVersion      *string `json:"app_version,omitempty"`
	DeviceModel     *string `json:"device_model,omitempty"`
	DeviceName      *string `json:"device_name,omitempty"`
	ScreenSize      *string `json:"screen_size,omitempty"`
	IsEmulator      *bool   `json:"is_emulator,omitempty"`
}

func makeRequest(method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func main() {
	fmt.Println("=== Testing Health ===")
	resp, err := makeRequest("GET", "http://localhost:8080/health", nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Health Status: %d\n", resp.StatusCode)
	resp.Body.Close()
	
	fmt.Println("\n=== Testing Login (without device_info) ===")
	loginReq := LoginRequest{
		Username:   "testuser",
		Password:   "password123",
		RememberMe: false,
	}
	
	resp, err = makeRequest("POST", "http://localhost:8080/api/v1/auth/login", loginReq, 
		map[string]string{"Content-Type": "application/json"})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	
	fmt.Printf("Login Status: %d\n", resp.StatusCode)
	fmt.Printf("Login Response: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		fmt.Println("Login successful!")
	} else {
		fmt.Println("Login failed")
	}
}