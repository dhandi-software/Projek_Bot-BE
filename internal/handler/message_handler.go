package handler

import (
	"log"

	"bot_be/internal/service"

	"go.mau.fi/whatsmeow/types/events"
)

type MessageHandler struct {
	Service *service.BotService
}

func NewMessageHandler(svc *service.BotService) *MessageHandler {
	return &MessageHandler{Service: svc}
}

// EventHandler adalah fungsi utama yang dipanggil whatsmeow ketika ada event
func (h *MessageHandler) EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// Hanya proses jika ada teks dalam pesan (conversation biasa)
		// atau dari extended text message.
		var text string
		if v.Message.GetConversation() != "" {
			text = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage().GetText() != "" {
			text = v.Message.GetExtendedTextMessage().GetText()
		}

		if text != "" {
			var sender string
			if v.Info.IsFromMe {
				sender = "Admin"
			} else {
				sender = v.Info.PushName
				if sender == "" {
					sender = "+" + v.Info.Sender.User
				}
			}
			err := h.Service.HandleIncomingMessage(sender, text)
			if err != nil {
				log.Printf("Error saat handle pesan dari %s: %v\n", sender, err)
			}
		}
	}
}
