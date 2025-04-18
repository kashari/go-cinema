package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var LOGFILEBASE = "/home/mkashari/go-cinema/logs/"
var _log *log.Logger
var _f *os.File
var _today time.Time = time.Now()

func init() {
	var err error
	var infoLogFile = LOGFILEBASE + time.Now().Format("2006-01-02") + ".log"

	_f, err = os.OpenFile(infoLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panicf("error opening file: %v", err)
	}
	wr := io.MultiWriter(_f, os.Stdout)
	_log = log.New(wr, "INFO ", log.LstdFlags|log.Lmicroseconds)
	_log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !dateEqual(_today, time.Now()) {
			_today = time.Now()

			dailyLogFile := LOGFILEBASE + time.Now().Format("2006-01-02") + ".log"
			newF, err := os.OpenFile(dailyLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Panicf("error opening file: %v", err)
			}
			wr := io.MultiWriter(newF, os.Stdout)
			_log.SetOutput(wr)
		}

		var bs string
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			body, _ := io.ReadAll(c.Request.Body)
			bs = string(body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		// better if you have a user in the context

		go _log.Println(c.ClientIP(), c.Request.Method, c.Request.RequestURI, "REQUEST", bs)

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()
		mdl := ResponseModel{}

		if blw.Status() > 201 {
			resp, err := io.ReadAll(blw.body)
			if err == nil {
				json.Unmarshal(resp, &mdl)
			}
		}

		go _log.Println(c.Request.RemoteAddr, c.Request.Method, c.Request.RequestURI, "RESPONSE", blw.Status(), mdl.Message)
	}
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type ResponseModel struct {
	Message string `json:"message"`
}
