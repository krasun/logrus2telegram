package logrus2telegram_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/krasun/logrus2telegram"
)

func TestSimple(t *testing.T) {
	log := logrus.New()

	client := newClient(func(req *http.Request) *http.Response {
		expectedURL, _ := url.Parse("https://api.telegram.org/bottest_token/sendMessage")
		equals(t, expectedURL, req.URL)

		// TODO: test body

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			Header:     make(http.Header),
		}
	})

	hook, err := logrus2telegram.NewHook(
		"test_token",
		[]int64{42},
		logrus2telegram.UseClient(client),
		logrus2telegram.Format(func(e *logrus.Entry) (string, error) {
			return fmt.Sprintf("%s:%s", e.Level, e.Message), nil
		}),
	)
	ok(t, err)
	log.AddHook(hook)

	log.Infof("some_log_message")
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

type roundTripper func(req *http.Request) *http.Response

func (rt roundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return rt(request), nil
}

func newClient(roundTripper roundTripper) *http.Client {
	return &http.Client{
		Transport: roundTripper,
	}
}
