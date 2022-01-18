package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/veraison/endorsement"
	"github.com/veraison/veraison/provisioning/decoder"
	"github.com/veraison/veraison/provisioning/storeclient"
)

type IHandler interface {
	Submit(c *gin.Context)
}

type Handler struct {
	DecoderManager decoder.IDecoderManager
	StoreClient    storeclient.IStoreClient
}

func NewHandler(
	dm decoder.IDecoderManager,
	sc storeclient.IStoreClient,
) IHandler {
	return &Handler{
		DecoderManager: dm,
		StoreClient:    sc,
	}
}

type ProvisioningSession struct {
	Status        string  `json:"status"`
	Expiry        string  `json:"expiry"`
	FailureReason *string `json:"failure-reason,omitempty"`
}

const (
	ProvisioningSessionMediaType = "application/vnd.veraison.provisioning-session+json"
)

func (o *Handler) Submit(c *gin.Context) {
	// read the accept header and make sure that it's compatible with what we
	// support
	offered := c.NegotiateFormat(ProvisioningSessionMediaType)
	if offered != ProvisioningSessionMediaType {
		ReportProblem(c,
			http.StatusNotAcceptable,
			fmt.Sprintf("the only supported output format is %s", ProvisioningSessionMediaType),
		)
		return
	}

	// read media type
	mediaType := c.Request.Header.Get("Content-Type")

	if !o.DecoderManager.IsSupportedMediaType(mediaType) {
		c.Header("Accept", o.DecoderManager.SupportedMediaTypes())
		ReportProblem(c,
			http.StatusUnsupportedMediaType,
			fmt.Sprintf("no active plugin found for %s", mediaType),
		)
		return
	}

	// read body
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		ReportProblem(c,
			http.StatusBadRequest,
			fmt.Sprintf("error reading body: %s", err),
		)
		return
	}

	if len(payload) == 0 {
		ReportProblem(c,
			http.StatusBadRequest,
			"empty body",
		)
		return
	}

	// From here onwards we assume that a provisioning session exists and that
	// every further communication (apart from panics) will be through that
	// object instead of using RFC7807 Problem Details.  We can add support for
	// stashing session state later on when we will implement the asynchronous
	// API model.  For now, the object is created opportunistically.

	// pass data to the identified plugin for normalisation
	rsp, err := o.DecoderManager.Dispatch(mediaType, payload)
	if err != nil {
		sendFailedProvisioningSession(
			c,
			fmt.Sprintf("decoder manager returned error: %s", err),
		)
		return
	}

	// forward normalised data to the endorsement store
	if err := o.store(rsp); err != nil {
		sendFailedProvisioningSession(
			c,
			fmt.Sprintf("endorsement store returned error: %s", err),
		)
		return
	}

	sendSuccessfulProvisioningSession(c)
}

func (o *Handler) store(rsp *decoder.EndorsementDecoderResponse) error {
	for _, ta := range rsp.TrustAnchors {
		taReq := &endorsement.AddTrustAnchorRequest{TrustAnchor: ta}

		taRes, err := o.StoreClient.AddTrustAnchor(context.TODO(), taReq)
		if err != nil {
			return fmt.Errorf("store operation failed for trust anchor: %w", err)
		}

		if !taRes.GetStatus().Result {
			return fmt.Errorf(
				"store operation failed for trust anchor: %s",
				taRes.Status.GetErrorDetail(),
			)
		}
	}

	for _, swComp := range rsp.SwComponents {
		swCompReq := &endorsement.AddSwComponentsRequest{
			Info: []*endorsement.SwComponent{
				swComp,
			},
		}

		swCompRes, err := o.StoreClient.AddSwComponents(context.TODO(), swCompReq)
		if err != nil {
			return fmt.Errorf("store operation failed for software components: %w", err)
		}

		if !swCompRes.GetStatus().Result {
			return fmt.Errorf(
				"store operation failed for software components: %s",
				swCompRes.Status.GetErrorDetail(),
			)
		}
	}

	return nil
}

func sendFailedProvisioningSession(c *gin.Context, failureReason string) {
	c.Header("Content-Type", ProvisioningSessionMediaType)
	c.JSON(
		http.StatusOK,
		&ProvisioningSession{
			Status:        "failed",
			Expiry:        time.Now().Format(time.RFC3339),
			FailureReason: &failureReason,
		},
	)
}

func sendSuccessfulProvisioningSession(c *gin.Context) {
	c.Header("Content-Type", ProvisioningSessionMediaType)
	c.JSON(
		http.StatusOK,
		&ProvisioningSession{
			Status: "success",
			Expiry: time.Now().Format(time.RFC3339),
		},
	)
}
