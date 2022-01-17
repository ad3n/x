package x

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func ValidatteHeader(c *gin.Context, mandatory RequestMandatory, logger KafkaLogger, coreEndpoint *string) bool {
	if mandatory.BeType == "" {
		mandatory.BeType = BeNonCore
	}

	if mandatory.BeType == BeCore {
		c.Header(ChannelIDHeaderKey, *coreEndpoint)
	}

	go func() {
		var payload interface{}
		jsonData, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			temp := map[string]interface{}{}
			c.Request.ParseForm()
			for key, value := range c.Request.PostForm {
				temp[key] = value
			}

			payload = temp
		} else {
			err = json.Unmarshal(jsonData, &payload)
			if err != nil {
				payload = ""
			}
		}

		err = logger.Send(
			logger.CreateMessageFromContext(
				c, c.GetHeader(ChannelIDHeaderKey),
				c.Request.URL.Path,
				RequestTrue,
				fmt.Sprintf("%s%s?%s", c.Request.Host, c.Request.URL.Path, c.Request.URL.RawQuery),
				payload, nil,
			),
		)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	c.Set(BeTypeContextKey, mandatory.BeType)
	if BeCore == mandatory.BeType {
		return true
	}

	for _, k := range mandatory.Headers {
		if c.GetHeader(k) == "" {
			return false
		}
	}

	return true
}
