package main

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/DramaFever/go-logging"
	"github.com/DramaFever/raven-go"
)

var GitHash string

func main() {
	log, err := logging.LogToStdout(logging.DebugLvl, "https://33d48b23c05b43a6b076f4e17c9de0f1:c225ead12d0c4fdb949fac687d4ea75a@sentry.drama9.com/6", map[string]string{
		"version": "charlie",
		"client":  "example",
	})
	if err != nil {
		panic(err)
	}
	log.SetPackagePrefixes([]string{"github.com/DramaFever/go-logging"})
	log.SetRelease(GitHash)
	log.Debug("Using release", GitHash)
	log.Debug("This", "is", "a", "debug message")
	log.Infof("This is an %s message", "info")
	log.Warn("This is a warn message")
	log.Errorf("I've decided that the time %s is an error!", time.Now())
	err = errors.New("Something went wrong!")
	log.Errorf("There was an error: %+v\n", err)
	stack := raven.NewStacktrace(0, -1, nil)
	log.AddMeta(raven.NewException(err, stack)).Error("Manually included the error this time:", err)
	req, _ := http.NewRequest("POST", "https://www.dramafever.com/logs", bytes.NewBufferString(`{"a": 1, "b": 2}`))
	log.AddMeta(raven.NewHttp(req)).Warn("There was a warning on this request!")
}
