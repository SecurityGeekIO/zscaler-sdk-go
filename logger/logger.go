package logger

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey        ContextKey = "rid"
	RequestStartTimeKey ContextKey = "rstart"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type nopLogger struct{}

func (l *nopLogger) Printf(format string, v ...interface{}) {}

func NewNopLogger() Logger {
	return &nopLogger{}
}

type defaultLogger struct {
	logger  *log.Logger
	Verbose bool
}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	trimedF := strings.TrimSpace(format)
	if (strings.HasPrefix(trimedF, "[DEBUG]") || strings.HasPrefix(trimedF, "[TRACE]")) && !l.Verbose {
		return
	}

	l.logger.Printf(format, v...)
}

func GetDefaultLogger(loggerPrefix string) Logger {
	loggingEnabled, _ := strconv.ParseBool(os.Getenv("ZSCALER_SDK_LOG"))
	if !loggingEnabled {
		return &nopLogger{}
	}
	verbose, _ := strconv.ParseBool(os.Getenv("ZSCALER_SDK_VERBOSE"))
	return &defaultLogger{
		logger:  log.New(os.Stdout, loggerPrefix, log.LstdFlags|log.Lshortfile),
		Verbose: verbose,
	}
}

const (
	logReqMsg = `[DEBUG] Request "%s %s" details:
---[ ZSCALER SDK REQUEST | ID:%s ]-------------------------------
%s
---------------------------------------------------------`

	logRespMsg = `[DEBUG] Response "%s %s" details:
---[ ZSCALER SDK RESPONSE | ID:%s | Duration:%s ]--------------------------------
%s
-------------------------------------------------------`
)

func WriteLog(logger Logger, format string, args ...interface{}) {
	if logger != nil {
		logger.Printf(format, args...)
	}
}

func LogRequestSensitive(logger Logger, req *http.Request, reqID string, sensitiveContent []string) {
	if logger != nil && req != nil {
		out, err := httputil.DumpRequestOut(req, true)
		for _, s := range sensitiveContent {
			out = []byte(strings.ReplaceAll(string(out), s, "********"))
		}
		if err == nil {
			WriteLog(logger, logReqMsg, req.Method, req.URL, reqID, string(out))
		}
	}
}

func SetRequestDetails(req *http.Request) *http.Request {
	reqID := uuid.NewString()
	start := time.Now()
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, reqID)
	ctx = context.WithValue(ctx, RequestStartTimeKey, start)
	req = req.WithContext(ctx)

	return req
}

func GetRequestDetails(req *http.Request) (string, time.Time) {
	if req == nil {
		return "", time.Now()
	}
	ridAny := req.Context().Value(RequestIDKey)
	startAny := req.Context().Value(RequestStartTimeKey)

	if ridAny == nil || startAny == nil {
		return "", time.Now()
	}

	rid, ok := ridAny.(string)
	if !ok {
		return "", time.Now()
	}

	start, ok := startAny.(time.Time)
	if !ok {
		return "", time.Now()
	}

	return rid, start
}

func LogRequest(logger Logger, req *http.Request, otherHeaderParams map[string]string, body bool) {
	if logger != nil && req != nil {
		l, ok := logger.(*defaultLogger)
		if ok && l.Verbose {
			for k, v := range otherHeaderParams {
				req.Header.Add(k, v)
			}
		}
		reqID, _ := GetRequestDetails(req)
		out, err := httputil.DumpRequestOut(req, body)
		if err == nil {
			WriteLog(logger, logReqMsg, req.Method, req.URL, reqID, string(out))
		}
	}
}

func LogResponse(logger Logger, resp *http.Response) {
	if logger != nil && resp != nil {
		reqID, start := GetRequestDetails(resp.Request)
		// Dump the entire response
		out, err := httputil.DumpResponse(resp, true)
		if err == nil {
			WriteLog(logger, logRespMsg, resp.Request.Method, resp.Request.URL, reqID, time.Since(start).String(), string(out))
		} else {
			WriteLog(logger, logRespMsg, resp.Request.Method, resp.Request.URL, reqID, time.Since(start).String(), "Got error:"+err.Error())
		}
	}
}
