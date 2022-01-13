package x

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
)

const (
	BeCore    = "core"
	BeNonCore = "nonCore"

	BeTypeContextKey       = "be-type"
	ThreeScaleHeaderKey    = "x-3scale-proxy-secret-token"
	APIKeyHeaderKey        = "x-api-key"
	AuthorizationHeaderKey = "authorization"
)

type (
	RequestMandatory struct {
		Headers []string
		BeType  string
	}
)

func ThreeScaleValidation(storedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(ThreeScaleHeaderKey)
		if token != storedToken {
			CreateResponse(c, DomainInternal, ErrorFormatCode, gin.H{})

			return
		}

		c.Next()
	}
}

func ResponseLogger(logger KafkaLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &responseWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: c.Writer,
		}

		c.Writer = w
		c.Header(APIKeyHeaderKey, "")

		c.Next()

		go func() {
			var payload interface{}
			var data interface{}
			err := json.Unmarshal(w.body.Bytes(), &data)
			if err == nil {
				payload = data
			} else {
				payload = w.body.String()
			}

			err = logger.Send(
				logger.CreateMessageFromContext(
					c, c.GetHeader(ChannelIDHeaderKey),
					c.Request.URL.Path,
					RequestFalse,
					fmt.Sprintf("%s%s?%s", c.Request.Host, c.Request.URL.Path, c.Request.URL.RawQuery),
					payload, nil,
				),
			)
			if err != nil {
				fmt.Println(err.Error())
			}
		}()
	}
}

func ValidateHeader(mandatory RequestMandatory, logger KafkaLogger, coreEndpoint *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ValidatteHeader(c, mandatory, logger, coreEndpoint) {
			CreateResponse(c, DomainInternal, ErrorFormatCode, gin.H{})
		}

		c.Next()
	}
}
