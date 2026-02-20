package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

var _ ports.Notifier = (*TelegramAdapter)(nil)

type TelegramAdapter struct {
	Token   string
	ChatID  string
	TopicID int
	ApiURL  string
}

func NewTelegram(token, chatID, topicID string) *TelegramAdapter {
	tid, _ := strconv.Atoi(topicID)

	return &TelegramAdapter{
		Token:   token,
		ChatID:  chatID,
		TopicID: tid,
		ApiURL:  fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
	}
}

func (t *TelegramAdapter) SendText(message string) error {
	payload := map[string]any{
		"chat_id":    t.ChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	if t.TopicID > 0 {
		payload["message_thread_id"] = t.TopicID
	}

	return t.doRequest(payload)
}

func (t *TelegramAdapter) Send(offer domain.Offer) error {
	msg := fmt.Sprintf(
		"🚨 <b>NOVO ACHADO NO %s!</b>\n\n🎮 <b>%s</b>\n💰 Preço: <b>R$ %.2f</b>\n\n🔗 <a href=\"%s\">Ver Anúncio</a>",
		offer.Source, offer.Title, offer.Price, offer.Link,
	)

	return t.SendText(msg)
}

func (t *TelegramAdapter) doRequest(payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(t.ApiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		slog.Error("erro na api do telegram", "status", resp.StatusCode, "motivo", string(bodyBytes))
		return fmt.Errorf("erro na api do telegram: status %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
