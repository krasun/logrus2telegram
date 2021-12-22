package logrus2telegram_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/krasun/logrus2telegram"
)

func TestUseClientErrorOnNil(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(nil),
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestLevelsErrorOnEmptyLevels(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.Levels([]logrus.Level{}),
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestNotifyOnErrorOnEmpty(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.NotifyOn([]logrus.Level{}),
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestFormatErrorOnNil(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.Format(nil),
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestRequestTimeoutErrorOnNegativeTimeout(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.RequestTimeout(-1*time.Second),
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestErrorOnEmptyChatIDs(t *testing.T) {
	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{},
	)

	if hook != nil || err == nil {
		t.Errorf("expected error, but got nil")
	}
}

func TestSendRequestWithDefaultFormat(t *testing.T) {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})

	client := newClient(func(req *http.Request) (*http.Response, error) {
		expectedURL, _ := url.Parse("https://api.telegram.org/bottest_token/sendMessage")
		equals(t, expectedURL, req.URL)

		expected := struct {
			ChatID              int64  `json:"chat_id"`
			Text                string `json:"text"`
			DisableNotification bool   `json:"disable_notification"`
		}{
			42,
			"level=info msg=some_log_message\n",
			false,
		}
		actual := expected

		reader, err := req.GetBody()
		ok(t, err)
		json.NewDecoder(reader).Decode(&actual)

		equals(t, expected, actual)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			Header:     make(http.Header),
		}, nil
	})

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(client),
		logrus2telegram.Levels(logrus.AllLevels),
		logrus2telegram.RequestTimeout(3*time.Second),
	)
	ok(t, err)
	log.AddHook(hook)

	log.Infof("some_log_message")
}

func TestSendRequestWithCustomFormat(t *testing.T) {
	log := logrus.New()

	client := newClient(func(req *http.Request) (*http.Response, error) {
		expectedURL, _ := url.Parse("https://api.telegram.org/bottest_token/sendMessage")
		equals(t, expectedURL, req.URL)

		expected := struct {
			ChatID              int64  `json:"chat_id"`
			Text                string `json:"text"`
			DisableNotification bool   `json:"disable_notification"`
		}{
			42,
			"info:some_log_message",
			false,
		}
		actual := expected

		reader, err := req.GetBody()
		ok(t, err)
		json.NewDecoder(reader).Decode(&actual)

		equals(t, expected, actual)

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			Header:     make(http.Header),
		}, nil
	})

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(client),
		logrus2telegram.Format(func(e *logrus.Entry) (string, error) {
			return fmt.Sprintf("%s:%s", e.Level, e.Message), nil
		}),
		logrus2telegram.Levels(logrus.AllLevels),
		logrus2telegram.NotifyOn(logrus.AllLevels),
		logrus2telegram.RequestTimeout(3*time.Second),
	)
	ok(t, err)
	log.AddHook(hook)

	log.Infof("some_log_message")
}

func TestErrorsInHTTP(t *testing.T) {
	log := logrus.New()

	client := newClient(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("some HTTP error")
	})

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(client),
	)
	ok(t, err)
	log.AddHook(hook)

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("failed to create OS pipe: %s", err)
	}
	defer w.Close()

	stderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = stderr }()

	log.Infof("some_log_message")

	w.Close()

	output, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("failed to read all from stderr: %s", err)
	}

	fmt.Println(string(output))

	if !strings.Contains(string(output), "Failed to fire hook: failed to send HTTP request to Telegram API") {
		t.Errorf("failed to fail hook")
	}
}

func TestErrorsInHTTPStatusCode(t *testing.T) {
	log := logrus.New()

	client := newClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			Header:     make(http.Header),
		}, nil
	})

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(client),
	)
	ok(t, err)
	log.AddHook(hook)

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("failed to create OS pipe: %s", err)
	}
	defer w.Close()

	stderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = stderr }()

	log.Infof("some_log_message")

	w.Close()

	output, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("failed to read all from stderr: %s", err)
	}

	fmt.Println(string(output))

	if !strings.Contains(string(output), "Failed to fire hook: response status code is not 200, it is 400") {
		t.Errorf("failed to fail hook")
	}
}

func TestErrorsInFormat(t *testing.T) {
	log := logrus.New()

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.Format(func(e *logrus.Entry) (string, error) {
			return "", errors.New("some err")
		}),
	)
	ok(t, err)
	log.AddHook(hook)

	r, w, err := os.Pipe()
	if err != nil {
		t.Errorf("failed to create OS pipe: %s", err)
	}
	defer w.Close()

	stderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = stderr }()

	log.Infof("some_log_message")
	w.Close()

	output, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("failed to read all from stderr: %s", err)
	}

	if !strings.Contains(string(output), "Failed to fire hook: failed to format log entry: some err") {
		t.Errorf("failed to fail hook")
	}
}

// ok fails the test if an err is not nil.
func ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

type roundTripper func(req *http.Request) (*http.Response, error)

func (rt roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return rt(request)
}

func newClient(roundTripper roundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripper,
	}
}
