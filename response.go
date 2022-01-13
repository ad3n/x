package x

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ResponseCodeHeaderKey    = "X-ResponseCode"
	ResponseMessageHeaderKey = "X-ResponseDesc"

	DomainInternal = "fuse"
	DomainProspera = "pros"
	DomainTemenos  = "tmn"
	DomainWow      = "wow"
)

type (
	responseWriter struct {
		gin.ResponseWriter
		body *bytes.Buffer
	}
)

func (r *responseWriter) WriteString(s string) (n int, err error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func (r *responseWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func CreateResponse(c *gin.Context, domain string, code string, content interface{}) {
	message, ok := ResponseMap[domain][code]
	if !ok {
		code = ErrorUndefinedCode
		message = ErrorUndefinedMessage
	}

	c.Header(ResponseCodeHeaderKey, code)
	c.Header(ResponseMessageHeaderKey, message)
	c.JSON(http.StatusOK, content)
	c.Abort()
}
