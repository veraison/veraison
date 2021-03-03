package frontend

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/veraison/common"
	"github.com/veraison/tokenprocessor"
	"github.com/veraison/verifier"
)

var defaultMaxLifetime = 120

type Controller struct {
	logger         ILogger
	sessionManager *SessionManager
	tokenProcessor *tokenprocessor.TokenProcessor
	verifier       *verifier.Verifier
}

func NewController(logger ILogger, tp *tokenprocessor.TokenProcessor, v *verifier.Verifier) *Controller {
	c := new(Controller)
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
	nonceSize, err := strconv.Atoi(g.Query("nonceSize"))
	if err != nil {
		g.AbortWithError(http.StatusBadRequest, err) //nolint:errcheck
		return
	}

	session, err := c.sessionManager.StartSession(nonceSize)
	if err != nil {
		g.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
		return
	}

	reqURL := g.Request.URL
	basePath := path.Dir(reqURL.Path)
	sessionPath := path.Join(basePath, "session", strconv.FormatInt(session.GetID(), 10))

	g.Header("Content-Type", "application/rats-challenge-response-session+json")
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
		g.AbortWithError(http.StatusBadRequest, err) //nolint:errcheck
		return
	}

	session := c.sessionManager.GetSession(int64(sessionID))
	if session == nil {
		g.AbortWithError(http.StatusBadRequest, fmt.Errorf("no session with id %d", sessionID)) //nolint:errcheck
		return
	}

	contentTypes := g.Request.Header["Content-Type"]
	if len(contentTypes) != 1 {
		g.AbortWithError(http.StatusBadRequest, errors.New("must specify exactly one content type")) //nolint:errcheck
		return
	}
	tokenFormat := contentTypeToTokenFormat(contentTypes[0])

	tokenData, err := ioutil.ReadAll(g.Request.Body)
	if err != nil {
		g.AbortWithError(http.StatusBadRequest, err) //nolint:errcheck
		return
	}

	evidenceContext, err := c.tokenProcessor.Process(tenantID, tokenFormat, tokenData)
	if err != nil {
		g.AbortWithError(http.StatusBadRequest, err) //nolint:errcheck
		return
	}

	attestationResult, err := c.verifier.Verify(evidenceContext, simpleVerif)
	if err != nil {
		g.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
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
		g.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
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
