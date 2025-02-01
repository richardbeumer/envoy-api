package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

var (
	user       = os.Getenv("ENLIGHTEN_USERNAME")
	password   = os.Getenv("ENLIGHTEN_PASSWORD")
	envoySerial = os.Getenv("ENVOY_SERIAL")
	envoySite  = os.Getenv("ENVOY_SITE")
	envoyHost  = os.Getenv("ENVOY_HOST")
	token      string
)

const (
	loginURL = "https://entrez.enphaseenergy.com/login"
	tokenURL = "https://entrez.enphaseenergy.com/entrez_tokens"
)

func getToken() string {
	payloadLogin := map[string]string{
		"username": user,
		"password": password,
	}

	loginResponse, err := http.PostForm(loginURL, payloadLogin)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	defer loginResponse.Body.Close()

	payloadToken := map[string]string{
		"Site":      envoySite,
		"serialNum": envoySerial,
	}

	tokenResponse, err := http.PostForm(tokenURL, payloadToken)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}
	defer tokenResponse.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tokenResponse.Body)
	if err != nil {
		log.Fatalf("Failed to parse token response: %v", err)
	}

	token = doc.Find("textarea").Text()
	return token
}

func checkToken(token string) string {
	endpoint := fmt.Sprintf("https://%s/auth/check_jwt", envoyHost)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Cannot connect to Envoy: %v", err)
		return getToken()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Valid token." {
		log.Println("Refreshing token")
		return getToken()
	}

	return token
}

func getEnvoyData(c *gin.Context) {
	token = checkToken(token)
	endpoint := fmt.Sprintf("https://%s/ivp/pdm/energy", envoyHost)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Cannot connect to Envoy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot connect to Envoy"})
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	c.JSON(http.StatusOK, data["production"].(map[string]interface{})["pcu"])
}

func main() {
	r := gin.Default()
	r.GET("/production/", getEnvoyData)
	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}