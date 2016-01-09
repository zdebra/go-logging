package logging

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DramaFever/raven-go"
)

func getFilePath() string {
	var name []string
	_, filename, _, _ := runtime.Caller(1)
	filename = path.Dir(filename)
	if testing.Coverage() > 0 {
		gopath := strings.Split(os.Getenv("GOPATH"), ":")
		for _, p := range gopath {
			newFilename := strings.TrimPrefix(filename, path.Join(p, "src")+"/")
			if newFilename != filename {
				filename = newFilename
				break
			}
		}
		name = append(name, filename, "_test", "_obj_test")
	} else {
		name = append(name, filename)
	}
	name = append(name, "log.go")
	return path.Join(name...)
}

func TestLevelIncludes(t *testing.T) {
	type levelTest struct {
		// logLevel is the level configured on the log
		// stmtLevel is the level the statement is logged with
		// includes is whether or not the statment should be logged
		logLevel, stmtLevel Level
		includes            bool
	}
	levelTests := []levelTest{
		{logLevel: DebugLvl, stmtLevel: DebugLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: InfoLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: InfoLvl, stmtLevel: InfoLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: WarnLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: WarnLvl, stmtLevel: InfoLvl, includes: false},
		{logLevel: WarnLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: WarnLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: ErrorLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: InfoLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: WarnLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: ErrorLvl, includes: true},
	}
	for _, test := range levelTests {
		includes := test.logLevel.includes(test.stmtLevel)
		if includes != test.includes {
			t.Errorf("Expected %s.Includes(%s) to be %t, got %t", test.logLevel, test.stmtLevel, test.includes, includes)
		}
	}
}

func TestLevelAsSentryLevel(t *testing.T) {
	conversionTests := map[Level]raven.Severity{
		DebugLvl: raven.DEBUG,
		InfoLvl:  raven.INFO,
		WarnLvl:  raven.WARNING,
		ErrorLvl: raven.ERROR,
	}
	for lvl, sev := range conversionTests {
		result := lvl.asSentryLevel()
		if result != sev {
			t.Errorf("Expected %s to be raven severity %s, got %s instead", lvl, sev, result)
		}
	}
}

func TestItoa(t *testing.T) {
	testInts := map[int]string{
		0:       "0",
		1:       "1",
		2:       "2",
		3:       "3",
		4:       "4",
		5:       "5",
		6:       "6",
		7:       "7",
		8:       "8",
		9:       "9",
		10:      "10",
		20:      "20",
		30:      "30",
		40:      "40",
		50:      "50",
		60:      "60",
		70:      "70",
		80:      "80",
		90:      "90",
		100:     "100",
		200:     "200",
		300:     "300",
		400:     "400",
		500:     "500",
		1000:    "1000",
		2000:    "2000",
		3000:    "3000",
		10000:   "10000",
		100000:  "100000",
		1000000: "1000000",
	}
	for i, a := range testInts {
		var b []byte
		var wid int
		switch {
		case i < 10:
			wid = 1
		case i >= 10 && i < 100:
			wid = 2
		case i >= 100 && i < 1000:
			wid = 3
		case i >= 1000 && i < 10000:
			wid = 4
		case i >= 10000 && i < 100000:
			wid = 5
		case i >= 100000 && i < 1000000:
			wid = 6
		case i >= 1000000:
			wid = 7
		}
		itoa(&b, i, wid)
		if string(b) != a {
			t.Errorf("Expected %d to be %s, got %s instead\n", i, a, string(b))
		}
	}
}

func TestFormatHeader(t *testing.T) {
	type header struct {
		now   time.Time
		file  string
		line  int
		level Level
	}
	headers := map[string]header{
		"2015-07-02T13:28:42 [WARN] /my/test/file.go:145: ": {
			now:   time.Date(2015, time.July, 2, 13, 28, 42, 0, time.UTC),
			file:  "/my/test/file.go",
			line:  145,
			level: WarnLvl,
		},
	}
	for out, in := range headers {
		var buf []byte
		formatHeader(&buf, in.now, in.file, in.line, in.level)
		if string(buf) != out {
			t.Errorf("Expected output to be '%s', got '%s' from %+v\n", out, string(buf), in)
		}
	}
}

func TestOutput(t *testing.T) {
	var buf bytes.Buffer
	log, err := New(DebugLvl, &buf, "", nil)
	if err != nil {
		t.Fatalf("Unexpected error: %+v\n", err)
	}
	err = log.output(0, "My test output", InfoLvl)
	if err != nil {
		t.Errorf("Unexpected error: %+v\n", err)
	}
	year, month, day := time.Now().Date()
	hour, minute, second := time.Now().Clock()
	file := getFilePath()
	line := 472
	if testing.Coverage() > 0 {
		line = 577
	}
	expected := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d [%s] %s:%d: %s\n", year, month, day, hour, minute, second, InfoLvl, file, line, "My test output")
	if buf.String() != expected {
		t.Errorf("Expected output to be '%s', got '%s' instead\n", expected, buf.String())
	}
}

func TestHelpers(t *testing.T) {
	var buf bytes.Buffer
	type levelTest struct {
		// logLevel is the level configured on the log
		// stmtLevel is the level the statement is logged with
		// includes is whether or not the statment should be logged
		logLevel, stmtLevel Level
		includes            bool
	}
	log, err := New(DebugLvl, &buf, "", nil)
	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
	log = log.WithCallDepth(-2)

	levelTests := []levelTest{
		{logLevel: DebugLvl, stmtLevel: DebugLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: InfoLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: DebugLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: InfoLvl, stmtLevel: InfoLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: InfoLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: WarnLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: WarnLvl, stmtLevel: InfoLvl, includes: false},
		{logLevel: WarnLvl, stmtLevel: WarnLvl, includes: true},
		{logLevel: WarnLvl, stmtLevel: ErrorLvl, includes: true},
		{logLevel: ErrorLvl, stmtLevel: DebugLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: InfoLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: WarnLvl, includes: false},
		{logLevel: ErrorLvl, stmtLevel: ErrorLvl, includes: true},
	}

	year, month, day := time.Now().Date()
	hour, minute, second := time.Now().Clock()
	file := getFilePath()
	line := 405
	if testing.Coverage() > 0 {
		line = 500
	}
	for pos, test := range levelTests {
		buf.Reset()
		log = log.WithLevel(test.logLevel)
		var f func(...interface{})
		var ff func(string, ...interface{})
		switch test.stmtLevel {
		case DebugLvl:
			f = log.Debug
			ff = log.Debugf
		case InfoLvl:
			f = log.Info
			ff = log.Infof
		case WarnLvl:
			f = log.Warn
			ff = log.Warnf
		case ErrorLvl:
			f = log.Error
			ff = log.Errorf
		default:
			t.Errorf("Unexpected level: %s\n", test.stmtLevel)
		}
		f("Test number", pos)
		line = 406
		if testing.Coverage() > 0 {
			line = 501
		}
		var expectation string
		if test.includes {
			expectation = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d [%s] %s:%d: %s %d\n", year, month, day, hour, minute, second, test.stmtLevel, file, line, "Test number", pos)
		} else {
			expectation = ""
		}
		if buf.String() != expectation {
			t.Errorf("Expected `%s`, got `%s` from %#+v\n", expectation, buf.String(), test)
		}

		buf.Reset()
		ff("Test number %d", pos)
		line = 413
		if testing.Coverage() > 0 {
			line = 510
		}
		if test.includes {
			expectation = fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d [%s] %s:%d: %s %d\n", year, month, day, hour, minute, second, test.stmtLevel, file, line, "Test number", pos)
		} else {
			expectation = ""
		}
		if buf.String() != expectation {
			t.Errorf("Expected `%s`, got `%s` from %#+v\n", expectation, buf.String(), test)
		}
	}
}
