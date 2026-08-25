package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"bot_be/internal/provider"

	"github.com/gofiber/fiber/v2"
)

type ChatHandler struct {
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{}
}

// formatPhoneNumber formats raw number like 6281319240256 to +62 813-1924-0256
func formatPhoneNumber(raw string) string {
	re := regexp.MustCompile(`\D`)
	digits := re.ReplaceAllString(raw, "")

	if strings.HasPrefix(digits, "62") && len(digits) >= 10 {
		return fmt.Sprintf("+62 %s-%s-%s", digits[2:5], digits[5:9], digits[9:])
	}
	if len(digits) > 0 {
		return "+" + digits
	}
	return raw
}

// GetContacts fetches the list of chats/contacts from WhatsApp via WAHA with resolved contact names
func (h *ChatHandler) GetContacts(c *fiber.Ctx) error {
	// 1. Fetch contacts from WAHA to get saved contact names & pushnames
	contactNameMap := make(map[string]string)
	if respContacts, err := provider.DoWahaRequest("GET", "/api/contacts/all?session=default", nil); err == nil {
		if bodyContacts, err := ioutil.ReadAll(respContacts.Body); err == nil {
			var wahaContacts []map[string]interface{}
			if json.Unmarshal(bodyContacts, &wahaContacts) == nil {
				for _, ct := range wahaContacts {
					id, _ := ct["id"].(string)
					lid, _ := ct["lid"].(string)
					phone, _ := ct["phoneNumber"].(string)
					name, _ := ct["name"].(string)
					pushname, _ := ct["pushname"].(string)

					finalName := name
					if finalName == "" {
						finalName = pushname
					}

					if finalName != "" && !strings.Contains(finalName, "@") {
						if id != "" {
							contactNameMap[id] = finalName
							contactNameMap[strings.Split(id, "@")[0]] = finalName
						}
						if lid != "" {
							contactNameMap[lid] = finalName
							contactNameMap[strings.Split(lid, "@")[0]] = finalName
						}
						if phone != "" {
							contactNameMap[phone] = finalName
							contactNameMap[strings.Split(phone, "@")[0]] = finalName
						}
					}
				}
			}
		}
		respContacts.Body.Close()
	}

	// 2. Fetch active chats from WAHA
	resp, err := provider.DoWahaRequest("GET", "/api/default/chats", nil)
	if err != nil || resp.StatusCode >= 400 {
		return c.JSON([]interface{}{})
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var chats []map[string]interface{}
	if err := json.Unmarshal(body, &chats); err != nil {
		return c.JSON([]interface{}{})
	}

	type ContactResponse struct {
		ID       string `json:"id"`
		IsGroup  bool   `json:"isGroup"`
		Username string `json:"username"`
	}

	var response []ContactResponse

	for _, chat := range chats {
		id, _ := chat["id"].(string)
		if id == "" {
			continue
		}

		isGroup := strings.HasSuffix(id, "@g.us") || strings.HasSuffix(id, "@group.us")
		name, _ := chat["name"].(string)
		pushname, _ := chat["pushname"].(string)
		accountLid, _ := chat["accountLid"].(string)
		pnJid, _ := chat["pnJid"].(string)

		cleanNum := strings.Split(id, "@")[0]

		displayName := ""
		if name != "" && !strings.Contains(name, "@") {
			displayName = name
		} else if pushname != "" && !strings.Contains(pushname, "@") {
			displayName = pushname
		} else if mapped, ok := contactNameMap[id]; ok {
			displayName = mapped
		} else if mapped, ok := contactNameMap[pnJid]; ok && pnJid != "" {
			displayName = mapped
		} else if mapped, ok := contactNameMap[accountLid]; ok && accountLid != "" {
			displayName = mapped
		} else if mapped, ok := contactNameMap[cleanNum]; ok {
			displayName = mapped
		}

		// Direct individual contact fetch fallback if still no name
		if displayName == "" && !isGroup {
			if respCt, err := provider.DoWahaRequest("GET", "/api/default/contacts/"+id, nil); err == nil {
				if bodyCt, err := ioutil.ReadAll(respCt.Body); err == nil {
					var singleCt map[string]interface{}
					if json.Unmarshal(bodyCt, &singleCt) == nil {
						if n, ok := singleCt["name"].(string); ok && n != "" && !strings.Contains(n, "@") {
							displayName = n
						} else if p, ok := singleCt["pushname"].(string); ok && p != "" && !strings.Contains(p, "@") {
							displayName = p
						}
					}
				}
				respCt.Body.Close()
			}
		}

		if displayName == "" {
			if isGroup {
				displayName = name
				if displayName == "" {
					displayName = "Grup WhatsApp"
				}
			} else {
				displayName = formatPhoneNumber(cleanNum)
			}
		}

		response = append(response, ContactResponse{
			ID:       id,
			IsGroup:  isGroup,
			Username: displayName,
		})
	}

	return c.JSON(response)
}

// GetChatHistory fetches recent messages for a specific JID and formats them for the frontend
func (h *ChatHandler) GetChatHistory(c *fiber.Ctx) error {
	jidStr := c.Params("jid")
	
	resp, err := provider.DoWahaRequest("GET", fmt.Sprintf("/api/default/chats/%s/messages?limit=30", jidStr), nil)
	if err != nil || resp == nil || resp.StatusCode >= 400 {
		return c.JSON([]interface{}{})
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var rawMessages []map[string]interface{}
	if err := json.Unmarshal(body, &rawMessages); err != nil {
		return c.JSON([]interface{}{})
	}

	// Reverse rawMessages so messages are ordered chronologically ascending (oldest first)
	for i, j := 0, len(rawMessages)-1; i < j; i, j = i+1, j-1 {
		rawMessages[i], rawMessages[j] = rawMessages[j], rawMessages[i]
	}

	type FormattedMessage struct {
		ID        string            `json:"id"`
		Content   string            `json:"content"`
		SenderID  interface{}       `json:"senderId"`
		CreatedAt string            `json:"createdAt"`
		IsRead    bool              `json:"isRead"`
		IsDeleted bool              `json:"isDeleted"`
		Sender    map[string]string `json:"sender,omitempty"`
	}

	var formatted []FormattedMessage
	for _, msg := range rawMessages {
		id, _ := msg["id"].(string)
		content, _ := msg["body"].(string)
		fromMe, _ := msg["fromMe"].(bool)
		
		var ts int64
		if tsFloat, ok := msg["timestamp"].(float64); ok {
			ts = int64(tsFloat)
		}
		
		createdAt := time.Now().Format(time.RFC3339)
		if ts > 0 {
			createdAt = time.Unix(ts, 0).Format(time.RFC3339)
		}

		senderId := msg["from"]
		if fromMe {
			senderId = "Admin"
		}

		pushName := ""
		role := "user"
		if fromMe {
			pushName = "Admin"
			role = "admin"
		} else if data, ok := msg["_data"].(map[string]interface{}); ok {
			pushName, _ = data["pushName"].(string)
		}

		formatted = append(formatted, FormattedMessage{
			ID:        id,
			Content:   content,
			SenderID:  senderId,
			CreatedAt: createdAt,
			IsRead:    true,
			IsDeleted: false,
			Sender: map[string]string{
				"username": pushName,
				"role":     role,
			},
		})
	}
	
	return c.JSON(formatted)
}

// SendMessage sends a text message to a specific JID via WAHA
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	chatId := req.To
	if strings.HasSuffix(chatId, "@s.whatsapp.net") {
		chatId = strings.Replace(chatId, "@s.whatsapp.net", "@c.us", 1)
	}

	payload := map[string]interface{}{
		"session": "default",
		"chatId":  chatId,
		"text":    req.Message,
	}
	jsonPayload, _ := json.Marshal(payload)

	resp, err := provider.DoWahaRequest("POST", "/api/sendText", jsonPayload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return c.Status(resp.StatusCode).JSON(fiber.Map{"error": "Failed to send message via WAHA"})
	}

	return c.JSON(fiber.Map{"status": "success"})
}
