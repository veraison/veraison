package frontend

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moogar0880/problems"

	"github.com/veraison/common"
	"github.com/veraison/tokenprocessor"
	"github.com/veraison/verifier"
)

var defaultMaxLifetime = 120
var defaultNonceSize = 32

func reportProblem(g *gin.Context, status int, details ...string) {
	prob := problems.NewStatusProblem(status)

	if len(details) > 0 {
		prob.Detail = strings.Join(details, ", ")
	}

	g.Header("Content-Type", "application/problem+json")
	g.AbortWithStatusJSON(status, prob)
}

type Controller struct {
	logger         ILogger
	sessionManager *SessionManager
	tokenProcessor *tokenprocessor.TokenProcessor
	verifier       *verifier.Verifier
}

func NewController(logger ILogger, tp *tokenprocessor.TokenProcessor, v *verifier.Verifier) *Controller {
	c := &Controller{}
	c.Init(logger, tp, v)
	return c
}

func (c *Controller) Init(logger ILogger, tp *tokenprocessor.TokenProcessor, v *verifier.Verifier) {
	c.logger = logger
	c.tokenProcessor = tp
	c.verifier = v
	c.sessionManager = NewSessionManager(defaultMaxLifetime)
}

func (c Controller) NewSession(g *gin.Context) {
	nonceSizeString := g.Query("nonceSize")
	nonceString := g.Query("nonce")

	var nonce []byte
	var err error

	if nonceSizeString != "" && nonceString != "" {
		reportProblem(g, http.StatusBadRequest, "only one of \"nonce\" or \"nonceSize\" must be specified; found both")
		return
	}

	if nonceString != "" {
		nonce, err = base64.StdEncoding.DecodeString(nonceString)
		if err != nil {
			reportProblem(g, http.StatusBadRequest, "could not decode nonce", err.Error())
			return
		}
	} else { // generate nonce of specific size
		var nonceSize int

		if nonceSizeString != "" {
			nonceSize, err = strconv.Atoi(nonceSizeString)
			if err != nil {
				reportProblem(g, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			nonceSize = defaultNonceSize
		}

		nonce = make([]byte, nonceSize)
		_, err := rand.Read(nonce)
		if err != nil {
			reportProblem(g, http.StatusInternalServerError, err.Error())
			return
		}
	}

	session, err := c.sessionManager.StartSession(nonce)
	if err != nil {
		reportProblem(g, http.StatusInternalServerError, err.Error())
		return
	}

	reqURL := g.Request.URL
	basePath := path.Dir(reqURL.Path)
	sessionPath := path.Join(basePath, "session", strconv.FormatInt(session.GetID(), 10))

	g.Header("Content-Type", "application/vnd.veraison.challenge-response-session+json")
	g.Header("Location", sessionPath)
	g.Header("Cache-Control", "no-cache")
	g.JSON(http.StatusCreated, session.SessionInfo)
}

func (c Controller) Verify(g *gin.Context) {
	// TODO: implement multi-tenancy
	tenantID := 1

	// TODO: parameterise verification mode.
	simpleVerif := false

	sessionID, err := strconv.Atoi(g.Param("sessionId"))
	if err != nil {
		g.AbortWithError(http.StatusBadRequest, err)
		return
	}

	session := c.sessionManager.GetSession(int64(sessionID))
	if session == nil {
		reportProblem(g, http.StatusBadRequest, fmt.Sprintf("no session with id %d", sessionID))
		return
	}

	contentTypes := g.Request.Header["Content-Type"]
	if len(contentTypes) != 1 {
		reportProblem(g, http.StatusBadRequest, "must specify exactly one content type")
		return
	}
	tokenFormat := contentTypeToTokenFormat(contentTypes[0])

	tokenData, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		reportProblem(g, http.StatusBadRequest, err.Error())
		return
	}

	evidenceContext, err := c.tokenProcessor.Process(tenantID, tokenFormat, tokenData)
	if err != nil {
		reportProblem(g, http.StatusBadRequest, err.Error())
		return
	}

	attestationResult, err := c.verifier.Verify(evidenceContext, simpleVerif)
	if err != nil {
		reportProblem(g, http.StatusInternalServerError, err.Error())
		return
	}

	session.SetState("complete")

	responseBody := ResponseBody{
		SessionInfo: session.SessionInfo,
		Evidence: ResponseEvidence{
			Type:  contentTypes[0],
			Value: tokenData,
		},
		Result: *attestationResult,
	}

	if err = c.sessionManager.EndSession(session.GetID()); err != nil {
		reportProblem(g, http.StatusInternalServerError, err.Error())
		return
	}

	g.JSON(http.StatusOK, responseBody)
}

func (c Controller) Close() {
	c.verifier.Close()
}

func contentTypeToTokenFormat(contentType string) common.TokenFormat {
	switch contentType {
	case "application/psa-attestation-token":
		return common.PsaIatToken
	case "applitation/riot-attestation-token":
		return common.DiceToken
	default:
		return common.UnknownToken
	}
}
