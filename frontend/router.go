package frontend

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/veraison/verifier"
)

// Create a new gin engine with a custom JSON logger and 5xx on panic()
func initGin() *gin.Engine {
	r := gin.New()

	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		type _JsonLogFormatterParams struct {
			TimeStamp    time.Time     `json:"timestamp"`
			StatusCode   int           `json:"status_code"`
			Latency      time.Duration `json:"latency"`
			ClientIP     string        `json:"client_ip,omitempty"`
			Method       string        `json:"method"`
			Path         string        `json:"path"`
			ErrorMessage string        `json:"error_message,omitempty"`
			BodySize     int           `json:"body_size"`
		}

		b, _ := json.Marshal(&_JsonLogFormatterParams{
			TimeStamp:    param.TimeStamp,
			StatusCode:   param.StatusCode,
			Latency:      param.Latency,
			ClientIP:     param.ClientIP,
			Method:       param.Method,
			Path:         param.Path,
			ErrorMessage: param.ErrorMessage,
			BodySize:     param.BodySize,
		})

		return string(b) + "\n"
	}))

	r.Use(gin.Recovery())

	return r
}

func NewRouter(logger *zap.Logger, v *verifier.Verifier) *gin.Engine {
	ctrl := NewController(logger, v)

	r := initGin()

	rg := r.Group("/challenge-response/v1")
	{
		rg.POST("/newSession", ctrl.NewSession)
		rg.POST("/session/:sessionId", ctrl.Verify)
	}

	return r
}
