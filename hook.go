package logrus2telegram

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

// TelegramBotHook is a hook for Logrus logging library to send logs directly to Telegram.
type TelegramBotHook struct {
	levels   []logrus.Level
	notifyOn map[logrus.Level]struct{}
	client   *http.Client
	url      string
	chatIDs  []int64
	format   func(e *logrus.Entry) (string, error)
}

// config for hook.
type config struct {
	levels         []logrus.Level
	notifyOn       map[logrus.Level]struct{}
	requestTimeout time.Duration
	format         func(e *logrus.Entry) (string, error)
}

// Option configures the hook instance.
type Option func(*config) error

// NotifyOn enables notification in messages for specified log levels.
func NotifyOn(levels []logrus.Level) Option {
	return func(h *config) error {
		if len(levels) < 1 {
			return errors.New("at least one level for notification is required")
		}

		for _, level := range levels {
			h.notifyOn[level] = struct{}{}
		}

		return nil
	}
}

// Levels allows to specify levels for the hook.
func Levels(levels []logrus.Level) Option {
	return func(h *config) error {
		if len(levels) < 1 {
			return errors.New("at least one level is required")
		}

		h.levels = levels

		return nil
	}
}

// Format specifies the format function for the log entry.
func Format(format func(e *logrus.Entry) (string, error)) Option {
	return func(h *config) error {
		if format == nil {
			return errors.New("the format function is nil")
		}

		h.format = format

		return nil
	}
}

// RequestTimeout specifies HTTP request timeout to Telegram API.
func RequestTimeout(requestTimeout time.Duration) Option {
	return func(h *config) error {
		if requestTimeout < 0 {
			return errors.New("the request timeout must be positive")
		}

		h.requestTimeout = requestTimeout

		return nil
	}
}

func defaultFormat(entry *logrus.Entry) (string, error) {
	m, err := entry.String()
	if err != nil {
		return "", fmt.Errorf("failed to serialize log entry: %w", err)
	}

	return m, nil
}

// NewHook creates new hook instance.
func NewHook(token string, chatIDs []int64, options ...Option) (*TelegramBotHook, error) {
	if len(chatIDs) < 1 {
		return nil, errors.New("at least one chatID is required")
	}

	cfg := &config{
		levels:         []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel},
		format:         defaultFormat,
		notifyOn:       make(map[logrus.Level]struct{}),
		requestTimeout: 3 * time.Second,
	}

	for _, option := range options {
		err := option(cfg)
		if err != nil {
			return nil, err
		}
	}

	return &TelegramBotHook{
		client:   &http.Client{Timeout: cfg.requestTimeout},
		url:      fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
		chatIDs:  chatIDs,
		format:   cfg.format,
		notifyOn: cfg.notifyOn,
		levels:   cfg.levels,
	}, nil
}

// message is JSON payload representation sent to Telegram API.
type message struct {
	ChatID              int64  `json:"chat_id"`
	Text                string `json:"text"`
	DisableNotification bool   `json:"disable_notification"`
}

// Fire sends the log entry to Telegram.
func (h *TelegramBotHook) Fire(entry *logrus.Entry) error {
	text, err := h.format(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}

	disableNotification := !h.notify(entry.Level)
	for _, chatID := range h.chatIDs {
		encoded, err := json.Marshal(message{chatID, text, disableNotification})
		if err != nil {
			return err
		}

		request, err := http.NewRequest(http.MethodPost, h.url, bytes.NewBuffer(encoded))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := h.client.Do(request)
		if err != nil {
			return fmt.Errorf("failed to send HTTP request to Telegram API: %w", err)
		} else if response.StatusCode != http.StatusOK {
			return fmt.Errorf("")
		}

		if err := response.Body.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (h *TelegramBotHook) notify(l logrus.Level) bool {
	if len(h.notifyOn) == 0 {
		return true
	}

	if _, notify := h.notifyOn[l]; notify {
		return true
	}

	return false
}

// Levels define on which log levels this hook would trigger.
func (h *TelegramBotHook) Levels() []logrus.Level {
	return h.levels
}
