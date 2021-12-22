# logrus2telegram

[![Build Status](https://app.travis-ci.com/krasun/logrus2telegram.svg?branch=main)](https://app.travis-ci.com/krasun/logrus2telegram)
[![codecov](https://codecov.io/gh/krasun/logrus2telegram/branch/main/graph/badge.svg?token=8NU6LR4FQD)](https://codecov.io/gh/krasun/logrus2telegram)
[![Go Report Card](https://goreportcard.com/badge/github.com/krasun/logrus2telegram)](https://goreportcard.com/report/github.com/krasun/logrus2telegram)
[![GoDoc](https://godoc.org/https://godoc.org/github.com/krasun/logrus2telegram?status.svg)](https://godoc.org/github.com/krasun/logrus2telegram)

`logrus2telegram` is a Telegram bot hook for [Logrus](https://github.com/sirupsen/logrus) logging library in Go.

## Installation

As simple as:

```
go get github.com/krasun/logrus2telegram
```

## Usage 

```go
package main

import (
	log "github.com/sirupsen/logrus"
	
	"github.com/krasun/logrus2telegram"
)

func main() {	
	hook, err := logrus2telegram.NewHook(
		<token>, 
		[]int64{<chat identifier>},
		// the levels of messages sent to Telegram
		// default: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel}
		logrus2telegram.Levels(log.AllLevels),
		// default: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel, logrus.WarnLevel, logrus.InfoLevel}
		// the levels of messages sent to Telegram with notifications
		logrus2telegram.NotifyOn([]log.Level{log.PanicLevel, log.FatalLevel, log.ErrorLevel, log.InfoLevel}),
		// default: 3 * time.second
		logrus2telegram.RequestTimeout(10*time.Second),
		// default: entry.String() time="2021-12-22T14:48:56+02:00" level=debug msg="example"
		logrus2telegram.Format(func(e *log.Entry) (string, error) {
			return fmt.Sprintf("%s %s", strings.ToUpper(e.Level.String()), e.Message), nil
		}),
	)
	if err != nil {
		// ...
	}
	log.AddHook(hook)
```

## Tests 

To make sure that the code is fully tested and covered:

```
$ go test .
ok  	github.com/krasun/logrus2telegram	0.470s
```

## Known Usages 

...

## License 

**logrus2telegram** is released under [the MIT license](LICENSE).