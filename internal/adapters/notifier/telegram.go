package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/LXSCA7/gorimpo/internal/core/domain"
	"github.com/LXSCA7/gorimpo/internal/core/ports"
)

var _ ports.Notifier = (*TelegramAdapter)(nil)

type TelegramAdapter struct {
	Token  string
	ChatID string
	ApiURL string
	Routes map[string]string
}

func NewTelegram(token, chatID string) *TelegramAdapter {
	return &TelegramAdapter{
		Token:  token,
		ChatID: chatID,
		ApiURL: fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
	}
}

func (t *TelegramAdapter) SetRoutes(routes map[string]string) {
	t.Routes = routes
}

func (t *TelegramAdapter) SendText(message, category string) error {
	payload := map[string]any{
		"chat_id":    t.ChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	destID := t.Routes[category]
	if destID != "" && destID != "0" {
		topicID, _ := strconv.Atoi(destID)
		payload["message_thread_id"] = topicID
	}

	return t.doRequest(payload)
}

func (t *TelegramAdapter) Send(offer domain.Offer, category, searchTerm string, showSearchTerm bool) error {
	header := "🚨 <b>NOVO ACHADO NO " + offer.Source + "!</b>"
	if showSearchTerm {
		header = fmt.Sprintf("🔍 <i>Busca: %s</i>\n%s", searchTerm, header)
	}

	msg := fmt.Sprintf(
		"%s\n\n🎮 <b>%s</b>\n💰 Preço: <b>R$ %.2f</b>\n\n🔗 <a href=\"%s\">Ver Anúncio</a>",
		header, offer.Title, offer.Price, offer.Link,
	)

	return t.SendText(msg, category)
}

func (t *TelegramAdapter) SendPhoto(data []byte, caption string, category string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	destID := t.ChatID
	threadID := t.Routes[category]

	_ = w.WriteField("chat_id", destID)
	if threadID != "" && threadID != "0" {
		_ = w.WriteField("message_thread_id", threadID)
	}
	_ = w.WriteField("caption", caption)
	_ = w.WriteField("parse_mode", "HTML")

	fw, err := w.CreateFormFile("photo", "screenshot.jpg")
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
		return err
	}
	w.Close()

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto", t.Token)
	req, err := http.NewRequest("POST", apiURL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram error: %d", resp.StatusCode)
	}

	return nil
}

func (t *TelegramAdapter) CreateCategory(name string) (string, error) {
	type createTopicResponse struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageThreadID int `json:"message_thread_id"`
		} `json:"result"`
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/createForumTopic", t.Token)

	payload := map[string]any{
		"chat_id": t.ChatID,
		"name":    name,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erro ao montar json do novo tópico: %v", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erro de rede ao bater no telegram: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		slog.Error("Erro ao criar tópico no Telegram", "status", resp.StatusCode, "motivo", string(bodyBytes))
		return "", fmt.Errorf("telegram recusou criar tópico. Status: %d", resp.StatusCode)
	}

	var result createTopicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("erro ao decodificar a resposta do telegram: %v", err)
	}

	if !result.Ok {
		return "", fmt.Errorf("telegram retornou ok=false ao tentar criar o tópico")
	}

	return strconv.Itoa(result.Result.MessageThreadID), nil
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

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		var telegramErr struct {
			Parameters struct {
				RetryAfter int `json:"retry_after"`
			} `json:"parameters"`
		}

		if err := json.Unmarshal(bodyBytes, &telegramErr); err == nil && telegramErr.Parameters.RetryAfter > 0 {
			sleepTime := telegramErr.Parameters.RetryAfter

			slog.Warn("🚨 Telegram mandou a gente segurar (Erro 429).", "sleeping_seconds", sleepTime)
			time.Sleep(time.Duration(sleepTime) * time.Second)

			return t.doRequest(payload)
		}
	}

	slog.Error("Erro na API do Telegram", "status", resp.StatusCode, "motivo", string(bodyBytes))
	return fmt.Errorf("erro na api do telegram: status %d - %s", resp.StatusCode, string(bodyBytes))
}
