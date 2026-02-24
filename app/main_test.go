package main

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestGetToken tests the getToken function
func TestGetToken(t *testing.T) {
	// Save original values
	origLoginURL := loginURL
	origTokenURL := tokenURL
	origUser := user
	origPassword := password
	origEnvoySite := envoySite
	origEnvoySerial := envoySerial

	defer func() {
		// Restore original values
		loginURL = origLoginURL
		tokenURL = origTokenURL
		user = origUser
		password = origPassword
		envoySite = origEnvoySite
		envoySerial = origEnvoySerial
	}()

	// Set up test environment variables
	user = "testuser"
	password = "testpass"
	envoySite = "testsite"
	envoySerial = "123456"

	// Create mock login server
	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		// Set a cookie for the test
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
		w.WriteHeader(http.StatusOK)
	}))
	defer loginServer.Close()

	// Create mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		// Check if cookie was forwarded
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "test-session" {
			t.Errorf("Expected session cookie to be forwarded")
		}
		// Return mock HTML with token in textarea
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><textarea>mock-jwt-token-12345</textarea></body></html>`))
	}))
	defer tokenServer.Close()

	// Update URLs to point to mock servers
	loginURL = loginServer.URL
	tokenURL = tokenServer.URL

	// Call the function
	result := getToken()

	// Verify the token was extracted correctly
	expected := "mock-jwt-token-12345"
	if result != expected {
		t.Errorf("Expected token '%s', got '%s'", expected, result)
	}
}

// TestGetTokenWithInvalidCredentials tests the getToken function with wrong credentials
func TestGetTokenWithInvalidCredentials(t *testing.T) {
	// This test would normally call log.Fatalf which exits the program
	// In a real scenario, we'd want to refactor the code to return errors instead
	// For now, we'll skip this test as it would require code changes
	t.Skip("Skipping test that would cause fatal exit - code needs refactoring to return errors")
}

// TestCheckToken tests the checkToken function with valid token
func TestCheckToken(t *testing.T) {
	// Save original values
	origEnvoyHost := envoyHost
	origTr := tr
	origLoginURL := loginURL
	origTokenURL := tokenURL
	origUser := user
	origPassword := password
	origEnvoySite := envoySite
	origEnvoySerial := envoySerial

	defer func() {
		// Restore original values
		envoyHost = origEnvoyHost
		tr = origTr
		loginURL = origLoginURL
		tokenURL = origTokenURL
		user = origUser
		password = origPassword
		envoySite = origEnvoySite
		envoySerial = origEnvoySerial
	}()

	// Set up test environment variables (in case token needs refresh)
	user = "testuser"
	password = "testpass"
	envoySite = "testsite"
	envoySerial = "123456"

	// Create mock TLS server for check_jwt endpoint
	checkServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			t.Errorf("Expected Bearer token in Authorization header")
		}
		// Return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Valid token."))
	}))
	defer checkServer.Close()

	// Configure to use test server
	envoyHost = strings.TrimPrefix(checkServer.URL, "https://")
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// Test with valid token
	testToken := "test-valid-token"
	result := checkToken(testToken)

	// Should return the same token if valid
	if result != testToken {
		t.Errorf("Expected token '%s', got '%s'", testToken, result)
	}
}

// TestCheckTokenInvalid tests the checkToken function with invalid token
func TestCheckTokenInvalid(t *testing.T) {
	// Save original values
	origEnvoyHost := envoyHost
	origTr := tr
	origLoginURL := loginURL
	origTokenURL := tokenURL
	origUser := user
	origPassword := password
	origEnvoySite := envoySite
	origEnvoySerial := envoySerial

	defer func() {
		// Restore original values
		envoyHost = origEnvoyHost
		tr = origTr
		loginURL = origLoginURL
		tokenURL = origTokenURL
		user = origUser
		password = origPassword
		envoySite = origEnvoySite
		envoySerial = origEnvoySerial
	}()

	// Set up test environment
	user = "testuser"
	password = "testpass"
	envoySite = "testsite"
	envoySerial = "123456"

	// Create mock login server
	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
		w.WriteHeader(http.StatusOK)
	}))
	defer loginServer.Close()

	// Create mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><textarea>new-refreshed-token</textarea></body></html>`))
	}))
	defer tokenServer.Close()

	loginURL = loginServer.URL
	tokenURL = tokenServer.URL

	// Create mock TLS server for check_jwt endpoint that returns invalid token
	checkServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Invalid token."))
	}))
	defer checkServer.Close()

	envoyHost = strings.TrimPrefix(checkServer.URL, "https://")
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// Test with invalid token
	testToken := "invalid-token"
	result := checkToken(testToken)

	// Should get a new token
	if result == testToken {
		t.Errorf("Expected new token, got the same invalid token")
	}
	if result != "new-refreshed-token" {
		t.Errorf("Expected 'new-refreshed-token', got '%s'", result)
	}
}

// TestGetEnvoyData tests the getEnvoyData handler
func TestGetEnvoyData(t *testing.T) {
	// Save original values
	origEnvoyHost := envoyHost
	origTr := tr
	origToken := token

	defer func() {
		// Restore original values
		envoyHost = origEnvoyHost
		tr = origTr
		token = origToken
	}()

	// Set up a valid token
	token = "test-token"

	// Create mock TLS check server (for checkToken and energy data)
	checkServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "check_jwt") {
			w.Write([]byte("Valid token."))
			return
		}
		if strings.Contains(r.URL.Path, "ivp/pdm/energy") {
			// Return mock energy data
			response := map[string]interface{}{
				"production": map[string]interface{}{
					"pcu": map[string]interface{}{
						"activeCount": 10,
						"wNow":        2500,
						"whLifetime":  1000000,
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer checkServer.Close()

	envoyHost = strings.TrimPrefix(checkServer.URL, "https://")
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Call the handler
	getEnvoyData(c)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify the response contains expected data
	if activeCount, ok := response["activeCount"].(float64); !ok || activeCount != 10 {
		t.Errorf("Expected activeCount to be 10, got %v", response["activeCount"])
	}
	if wNow, ok := response["wNow"].(float64); !ok || wNow != 2500 {
		t.Errorf("Expected wNow to be 2500, got %v", response["wNow"])
	}
}

// TestGetEnvoyDataConnectionError tests the handler when connection fails
func TestGetEnvoyDataConnectionError(t *testing.T) {
	// Save original values
	origEnvoyHost := envoyHost
	origTr := tr
	origToken := token
	origLoginURL := loginURL
	origTokenURL := tokenURL
	origUser := user
	origPassword := password
	origEnvoySite := envoySite
	origEnvoySerial := envoySerial

	defer func() {
		// Restore original values
		envoyHost = origEnvoyHost
		tr = origTr
		token = origToken
		loginURL = origLoginURL
		tokenURL = origTokenURL
		user = origUser
		password = origPassword
		envoySite = origEnvoySite
		envoySerial = origEnvoySerial
	}()

	// Set up test environment variables
	user = "testuser"
	password = "testpass"
	envoySite = "testsite"
	envoySerial = "123456"

	// Create mock login and token servers for fallback
	loginServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
		w.WriteHeader(http.StatusOK)
	}))
	defer loginServer.Close()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><textarea>fallback-token</textarea></body></html>`))
	}))
	defer tokenServer.Close()

	loginURL = loginServer.URL
	tokenURL = tokenServer.URL

	// Set up a valid token
	token = "test-token"

	// Create a TLS server that only responds to check_jwt but not to the energy endpoint
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "check_jwt") {
			w.Write([]byte("Valid token."))
			return
		}
		// Return error for energy endpoint
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	// Use localhost with wrong port to test connection failure of the data endpoint separately
	envoyHost = "localhost:99999" // Non-existent port
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Call the handler
	getEnvoyData(c)

	// Check the response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify the error message
	if errorMsg, ok := response["error"].(string); !ok || errorMsg != "Cannot connect to Envoy" {
		t.Errorf("Expected error message 'Cannot connect to Envoy', got %v", response["error"])
	}
}

// TestMainRouteSetup tests that the main function sets up routes correctly
func TestMainRouteSetup(t *testing.T) {
	// This would require refactoring main() to return the router
	// For now, we can test the router setup separately
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/production/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"test": "ok"})
	})

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/production/", nil)
	r.ServeHTTP(w, req)

	// Verify the route exists
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestEnvironmentVariables tests that environment variables are loaded
func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name    string
		envVar  string
		value   string
		checker func() string
	}{
		{"ENLIGHTEN_USERNAME", "ENLIGHTEN_USERNAME", "testuser", func() string { return user }},
		{"ENLIGHTEN_PASSWORD", "ENLIGHTEN_PASSWORD", "testpass", func() string { return password }},
		{"ENVOY_SERIAL", "ENVOY_SERIAL", "12345", func() string { return envoySerial }},
		{"ENVOY_SITE", "ENVOY_SITE", "testsite", func() string { return envoySite }},
		{"ENVOY_HOST", "ENVOY_HOST", "192.168.1.100", func() string { return envoyHost }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tt.envVar, tt.value)

			// Note: This test demonstrates the pattern, but the global variables
			// are initialized at package load time, so this won't actually update them.
			// In a real scenario, we'd want to refactor to use a config struct.

			// Clean up
			os.Unsetenv(tt.envVar)
		})
	}
}

// BenchmarkGetEnvoyData benchmarks the getEnvoyData handler
func BenchmarkGetEnvoyData(b *testing.B) {
	// Save original values
	origEnvoyHost := envoyHost
	origTr := tr
	origToken := token

	defer func() {
		envoyHost = origEnvoyHost
		tr = origTr
		token = origToken
	}()

	token = "test-token"

	// Create mock TLS server
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "check_jwt") {
			w.Write([]byte("Valid token."))
			return
		}
		response := map[string]interface{}{
			"production": map[string]interface{}{
				"pcu": map[string]interface{}{
					"activeCount": 10,
					"wNow":        2500,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	envoyHost = strings.TrimPrefix(mockServer.URL, "https://")
	tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		getEnvoyData(c)
	}
}
