package main

import (
	"crypto/tls"
	"encoding/json"

	//"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

var (
    user        = os.Getenv("ENLIGHTEN_USERNAME")
    password    = os.Getenv("ENLIGHTEN_PASSWORD")
    envoySerial = os.Getenv("ENVOY_SERIAL")
    envoySite   = os.Getenv("ENVOY_SITE")
    envoyHost   = os.Getenv("ENVOY_HOST")
    loginURL    = "https://entrez.enphaseenergy.com/login"
    tokenURL    = "https://entrez.enphaseenergy.com/entrez_tokens"
    token       string
	tr          = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},}
)

func getToken() string {
    payloadLogin := url.Values{
        "username": {user},
        "password": {password},
    }

    resp, err := http.PostForm(loginURL, payloadLogin)
    if err != nil {
        log.Fatalf("Failed to login: %v", err)
    }
    defer resp.Body.Close()

    payloadToken := url.Values{
        "Site":      {envoySite},
        "serialNum": {envoySerial},
    }

    client := &http.Client{}
    req, err := http.NewRequest("POST", tokenURL, strings.NewReader(payloadToken.Encode()))
    if err != nil {
        log.Fatalf("Failed to create token request: %v", err)
    }
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		for _, cookie := range resp.Cookies() {
			req.AddCookie(cookie)
	}

    tokenResp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Failed to get token: %v", err)
    }
    defer tokenResp.Body.Close()

    doc, err := goquery.NewDocumentFromReader(tokenResp.Body)
    if err != nil {
        log.Fatalf("Failed to parse token response: %v", err)
    }

    
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    bodyString := string(bodyBytes)

    if strings.Contains(bodyString, "Wrong username or password") {
        log.Fatalf("Error: Wrong username or password for Enlighten")
    }
    
    token = doc.Find("textarea").Text()
    return token
}

func checkToken(token string) string {
    endpoint := "https://" + envoyHost + "/auth/check_jwt"
    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        log.Fatalf("Failed to create check token request: %v", err)
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", "Bearer "+token)

    client := &http.Client{Transport: tr}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Cannot connect to Envoy: %v", err)
        return getToken()
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    bodyString := string(bodyBytes)
    
    if !strings.Contains(bodyString, "Valid token.") {
        log.Println("Refreshing token")
        return getToken()
    }

    return token
}

func getEnvoyData(c *gin.Context) {
    token = checkToken(token)
    endpoint := "https://" + envoyHost + "/ivp/pdm/energy"
    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        log.Fatalf("Failed to create data request: %v", err)
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", "Bearer "+token)

    client := &http.Client{Transport: tr}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Cannot connect to Envoy: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot connect to Envoy"})
        return
    }
    defer resp.Body.Close()

    var data map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&data)

    c.JSON(http.StatusOK, data["production"].(map[string]interface{})["pcu"])
}

func main() {
    r := gin.Default()
    r.GET("/production/", getEnvoyData)
    r.Run()
}
