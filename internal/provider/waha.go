package provider

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

// WAHA API URL comes from env or defaults to http://waha:3000
func GetWahaURL() string {
	url := os.Getenv("WAHA_API_URL")
	if url == "" {
		return "http://localhost:3000"
	}
	return url
}

// GetWahaAPIKey gets the API key from env
func GetWahaAPIKey() string {
	// We'll hardcode it to match docker-compose for now, or read from env
	return "dhandi_waha_secret" 
}

func DoWahaRequest(method, endpoint string, payload []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if payload != nil {
		req, err = http.NewRequest(method, GetWahaURL()+endpoint, bytes.NewBuffer(payload))
	} else {
		req, err = http.NewRequest(method, GetWahaURL()+endpoint, nil)
	}
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", GetWahaAPIKey()) // WAHA API Key auth
	
	client := &http.Client{}
	return client.Do(req)
}

// StartSession initiates the 'default' session in WAHA
func StartSession() error {
	webhookUrl := "http://backend:8080/api/wa/webhook"
	n8nWebhookUrl := os.Getenv("N8N_WEBHOOK_URL")

	webhooks := []map[string]interface{}{
		{
			"url": webhookUrl,
			"events": []string{
				"message",
				"message.any",
				"session.status",
			},
		},
	}

	if n8nWebhookUrl != "" {
		webhooks = append(webhooks, map[string]interface{}{
			"url": n8nWebhookUrl,
			"events": []string{
				"message",
				"message.any",
			},
		})
	}

	payload := map[string]interface{}{
		"name": "default",
		"config": map[string]interface{}{
			"noweb": map[string]interface{}{
				"store": map[string]interface{}{
					"enabled":  true,
					"fullSync": true,
				},
			},
			"webhooks": webhooks,
		},
	}
	jsonPayload, _ := json.Marshal(payload)
	
	resp, err := DoWahaRequest("POST", "/api/sessions/start", jsonPayload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetSessionStatus returns the status of the 'default' session
func GetSessionStatus() (string, error) {
	resp, err := DoWahaRequest("GET", "/api/sessions?all=true", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var sessions []map[string]interface{}
	if err := json.Unmarshal(body, &sessions); err != nil {
		return "", err
	}
	
	for _, s := range sessions {
		if s["name"] == "default" {
			if status, ok := s["status"].(string); ok {
				return status, nil
			}
		}
	}
	return "STOPPED", nil
}

// LogoutSession stops and logs out the session
func LogoutSession() error {
	payload := map[string]string{"name": "default"}
	jsonPayload, _ := json.Marshal(payload)
	
	resp, err := DoWahaRequest("POST", "/api/sessions/logout", jsonPayload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Latest QR code is kept here to serve the frontend
var WAQRString string

