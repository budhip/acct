package middleware

import (
	"bytes"
	"io"
	"strings"
	"time"

	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	xlog "bitbucket.org/Amartha/go-x/log"
)

const (
	httpLog = "[HTTP-REQUEST]"
)

func (m *AppMiddleware) logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			ctx := c.Request().Context()
			req := c.Request()
			res := c.Response()
			reqBody := m.parseRequestBody(c)
			resBody := m.getResponseBody(c)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			p := req.URL.Path
			if p == "" {
				p = "/"
			}

			headers := &headerBag{
				vals: make(map[string]string),
			}

			for name, values := range req.Header {
				for _, value := range values {
					if strings.ToLower(name) == "x-secret-key" ||
						strings.ToLower(name) == "authorization" {
						continue
					}
					headers.vals[strings.ToLower(name)] = value
				}
			}

			var requestBody, responseBody map[string]interface{}
			json.Unmarshal(reqBody, &requestBody)
			json.Unmarshal(resBody.Bytes(), &responseBody)

			latency := time.Since(start)

			fields := []xlog.Field{
				xlog.String("fullPath", req.RequestURI),
				xlog.String("path", req.URL.Path),
				xlog.String("route", c.Path()),
				xlog.String("method", req.Method),
				xlog.Int("status", res.Status),
				xlog.Object("headers", headers),
				xlog.Any("request", requestBody),
				xlog.Any("response", responseBody),
				xlog.String("latency", latency.String()),
			}

			requestId := req.Header.Get("X-Request-Id")
			if requestId != "" {
				fields = append(fields, xlog.String("request-id", requestId))
			}

			n := res.Status
			switch {
			case n >= 500:
				xlog.Error(ctx, httpLog, fields...)
			case n >= 400:
				xlog.Warn(ctx, httpLog, fields...)
			case n >= 300:
				xlog.Info(ctx, httpLog, fields...)
			default:
				xlog.Info(ctx, httpLog, fields...)
			}

			return nil
		}
	}
}

type headerBag struct {
	vals map[string]string
}

func (h *headerBag) MarshalLogObject(enc xlog.ObjectEncoder) error {
	for k, v := range h.vals {
		enc.AddString(k, v)
	}
	return nil
}

func (m *AppMiddleware) parseRequestBody(c echo.Context) []byte {
	var body []byte
	if c.Request().Body != nil {
		body, _ = io.ReadAll(c.Request().Body)
	}
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}

func (m *AppMiddleware) getResponseBody(c echo.Context) *bytes.Buffer {
	resBody := new(bytes.Buffer)
	mw := io.MultiWriter(c.Response().Writer, resBody)
	writer := &bodyDumpResponseWriter{writer: mw, responseWriter: c.Response().Writer}
	c.Response().Writer = writer
	return resBody
}

type bodyDumpResponseWriter struct {
	writer         io.Writer
	responseWriter http.ResponseWriter
}

func (w *bodyDumpResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.responseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}
