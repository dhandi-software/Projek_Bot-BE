package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bot_be/internal/provider"
	"bot_be/internal/service"
	"bot_be/internal/wshub"

	"github.com/gofiber/fiber/v2"
)

type MessageHandler struct {
	Service *service.BotService
}

func NewMessageHandler(svc *service.BotService) *MessageHandler {
	return &MessageHandler{Service: svc}
}

// HandleWebhook is called by the Fiber router when WAHA sends a webhook
func (h *MessageHandler) HandleWebhook(c *fiber.Ctx) error {
	var body map[string]interface{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	event, _ := body["event"].(string)

	if event == "session.status" {
		payload, _ := body["payload"].(map[string]interface{})
		status, _ := payload["status"].(string)
		
		if status == "SCAN_QR_CODE" {
			// WAHA returns QR code in base64 sometimes, or we can fetch it
			// For simplicity we will set status, frontend will hit /api/wa/qr
			// We can fetch QR image from /api/screenshot later if needed
		} else if status == "WORKING" {
			log.Println("WAHA Session connected!")
		}
		return c.SendStatus(fiber.StatusOK)
	}

	if event == "message" || event == "message.any" {
		payload, ok := body["payload"].(map[string]interface{})
		if !ok {
			return c.SendStatus(fiber.StatusOK)
		}

		text, _ := payload["body"].(string)
		from, _ := payload["from"].(string)
		id, _ := payload["id"].(string)

		// Check if 'from' is an internal LID (e.g. 144998163022079@lid) and convert to real phone JID
		if strings.HasSuffix(from, "@lid") {
			if _data, ok := payload["_data"].(map[string]interface{}); ok {
				if key, ok := _data["key"].(map[string]interface{}); ok {
					if alt, ok := key["remoteJidAlt"].(string); ok && alt != "" {
						from = alt
						payload["from"] = alt
					}
				}
			}
		}
		
		isFromMe, _ := payload["fromMe"].(bool)

		to, _ := payload["to"].(string)
		if to == "" {
			to, _ = payload["chatId"].(string)
		}

		senderId := from
		roomId := from
		senderRole := "user"

		if isFromMe {
			senderId = "Admin"
			senderRole = "admin"
			if to != "" {
				roomId = to
			}
		}

		if text != "" {
			sender := from
			if isFromMe {
				sender = "Admin"
			}
			
			// Process with BotService
			err := h.Service.HandleIncomingMessage(sender, text)
			if err != nil {
				log.Printf("Error saat handle pesan dari %s: %v\n", sender, err)
			}

			// Broadcast to WebSockets
			wsPayload := map[string]interface{}{
				"event": "receive_message",
				"data": map[string]interface{}{
					"id":        id,
					"content":   text,
					"senderId":  senderId,
					"roomId":    roomId,
					"isGroup":   false,
					"isPublic":  false,
					"sender": map[string]interface{}{
						"username": sender,
						"role":     senderRole,
					},
					"createdAt": time.Now().Format(time.RFC3339),
				},
			}
			
			wshub.BroadcastMessage(wsPayload)

			// Forward to n8n Webhook if configured and message is not from self
			n8nUrl := os.Getenv("N8N_WEBHOOK_URL")
			if n8nUrl != "" && !isFromMe {
				go forwardToN8n(n8nUrl, body)
			}
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func forwardToN8n(targetURL string, body map[string]interface{}) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}

// FetchQR helper function to be called by /api/wa/qr
func FetchQR() string {
	// If session doesn't exist or stopped, auto start session
	status, _ := provider.GetSessionStatus()
	if status == "STOPPED" || status == "" {
		_ = provider.StartSession()
		time.Sleep(1 * time.Second)
	}

	resp, err := provider.DoWahaRequest("GET", "/api/default/auth/qr", nil)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(body) == 0 {
			return ""
		}

		// Check if JSON response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err == nil {
			if data, ok := result["data"].(string); ok && data != "" {
				if strings.HasPrefix(data, "data:image") {
					return data
				}
				return "data:image/png;base64," + data
			}
			if raw, ok := result["raw"].(string); ok && raw != "" {
				return raw
			}
		}

		// If raw binary PNG from WAHA, base64 encode it
		encoded := base64.StdEncoding.EncodeToString(body)
		return "data:image/png;base64," + encoded
	}
	return ""
}

