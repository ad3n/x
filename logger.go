package x

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
)

const (
	RequestTrue  = "true"
	RequestFalse = "false"

	CorrelationIDHeaderKey = "x-correlation-id"
	StanIDHeaderKey        = "x-stan-id"
	ChannelIDHeaderKey     = "x-channel-id"
)

type (
	Message struct {
		Date            string      `json:"date"`
		CorrelationID   string      `json:"X-Correlation-Id"`
		StanID          string      `json:"X-Stan-Id"`
		TransactionType string      `json:"trxType"`
		Channel         string      `json:"channel"`
		IsRequest       string      `json:"isRequest"`
		URL             string      `json:"url"`
		Headers         string      `json:"headers"`
		Payload         interface{} `json:"payload"`
	}

	KafkaLogger struct {
		producer sarama.SyncProducer
		topic    string
	}
)

func NewKafkaLogger(servers []string, topic string) KafkaLogger {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(servers, config)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return KafkaLogger{
		producer: producer,
		topic:    topic,
	}
}

func (l KafkaLogger) CreateMessageFromContext(
	c *gin.Context,
	channel string,
	trxType string,
	isRequest string,
	url string,
	payload interface{},
	exceptionMessage *string,
) Message {
	headers := map[string][]string{}
	for k, v := range c.Request.Header {
		nK := strings.ToLower(k)
		if nK == APIKeyHeaderKey || nK == AuthorizationHeaderKey {
			if len(v[0]) > 4 {
				v[0] = fmt.Sprintf("%s***********%s", v[0][:2], v[0][len(v[0])-3:])
			} else {
				v[0] = fmt.Sprintf("%s***********%s", v[0], v[0])
			}
		}

		headers[k] = v
	}

	blank := ""
	if exceptionMessage == nil {
		exceptionMessage = &blank
	}

	if *exceptionMessage != blank {
		var file string
		var line int
		var caller string

		pc, file, line, ok := runtime.Caller(1)
		detail := runtime.FuncForPC(pc)
		if ok || detail != nil {
			caller = detail.Name()
		}

		payload = map[string]interface{}{
			"payload":     payload,
			"stack_trace": fmt.Sprintf("%s#%d: %s", file, line, caller),
			"exception":   *exceptionMessage,
		}
	}

	dumped, err := json.Marshal(headers)
	if err != nil {
		dumped = []byte("")
	}

	return Message{
		Date:            time.Now().Format("2006-02-01 15:04:05.999"),
		CorrelationID:   c.GetHeader(CorrelationIDHeaderKey),
		StanID:          c.GetHeader(StanIDHeaderKey),
		TransactionType: trxType,
		Channel:         channel,
		IsRequest:       isRequest,
		URL:             url,
		Headers:         string(dumped),
		Payload:         payload,
	}
}

func (l KafkaLogger) Send(mesage Message) error {
	msg, err := json.Marshal(mesage)
	if err != nil {
		return err
	}

	_, _, err = l.producer.SendMessage(&sarama.ProducerMessage{
		Key:   sarama.StringEncoder(mesage.CorrelationID),
		Topic: l.topic,
		Value: sarama.ByteEncoder(msg),
	})

	if err != nil {
		return err
	}

	return nil
}
