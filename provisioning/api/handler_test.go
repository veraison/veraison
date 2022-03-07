// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"

	"github.com/veraison/endorsement"
	mock_deps "github.com/veraison/veraison/provisioning/api/mocks"
	"github.com/veraison/veraison/provisioning/decoder"
)

var (
	testGoodDecoderResponse = decoder.EndorsementDecoderResponse{
		TrustAnchors: []*endorsement.TrustAnchor{
			&endorsement.TrustAnchor{},
		},
		SwComponents: []*endorsement.SwComponent{
			&endorsement.SwComponent{},
		},
	}
	testFailedTaRes = endorsement.AddTrustAnchorResponse{
		Status: &endorsement.Status{Result: false},
	}
	testGoodTaRes = endorsement.AddTrustAnchorResponse{
		Status: &endorsement.Status{Result: true},
	}
	testFailedSwCompRes = endorsement.AddSwComponentsResponse{
		Status: &endorsement.Status{Result: false},
	}
	testGoodSwCompRes = endorsement.AddSwComponentsResponse{
		Status: &endorsement.Status{Result: true},
	}
)

func TestHandler_Submit_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ProvisioningSessionMediaType),
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	g.Request.Header.Add("Accept", "application/unsupported+ber")

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_UnsupportedMediaType(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/unsupported+json"
	supportedMediaTypes := "application/type-1, application/type-2"

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(false)
	dm.EXPECT().
		SupportedMediaTypes().
		Return(supportedMediaTypes)

	sc := mock_deps.NewMockIStoreClient(ctrl)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusUnsupportedMediaType
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Unsupported Media Type",
		Status: http.StatusUnsupportedMediaType,
		Detail: fmt.Sprintf("no active plugin found for %s", mediaType),
	}
	expectedAcceptHeader := supportedMediaTypes

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedAcceptHeader, w.Result().Header.Get("Accept"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_NoBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)

	sc := mock_deps.NewMockIStoreClient(ctrl)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "empty body",
	}

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	emptyBody := []byte("")

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(emptyBody))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_Submit_DecodeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	decoderError := "decoder manager says: doh!"

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(nil, errors.New(decoderError))

	sc := mock_deps.NewMockIStoreClient(ctrl)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf("decoder manager returned error: %s", decoderError)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_store_AddTrustAnchor_failure1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(&testGoodDecoderResponse, nil)

	sc := mock_deps.NewMockIStoreClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(nil, errors.New(storeError))

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for trust anchor: %s",
		storeError,
	)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_store_AddTrustAnchor_failure2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"
	testFailedTaRes.Status.ErrorDetail = storeError

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(&testGoodDecoderResponse, nil)

	sc := mock_deps.NewMockIStoreClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testFailedTaRes, nil)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for trust anchor: %s",
		storeError,
	)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_store_AddSwComponents_failure1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(&testGoodDecoderResponse, nil)

	sc := mock_deps.NewMockIStoreClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddSwComponents(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddSwComponentsRequest{
					Info: []*endorsement.SwComponent{
						&endorsement.SwComponent{},
					},
				},
			),
		).
		Return(nil, errors.New(storeError))

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for software components: %s",
		storeError,
	)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_store_AddSwComponents_failure2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")
	storeError := "store says doh!"
	testFailedSwCompRes.Status.ErrorDetail = storeError

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(&testGoodDecoderResponse, nil)

	sc := mock_deps.NewMockIStoreClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddSwComponents(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddSwComponentsRequest{
					Info: []*endorsement.SwComponent{
						&endorsement.SwComponent{},
					},
				},
			),
		).
		Return(&testFailedSwCompRes, nil)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedFailureReason := fmt.Sprintf(
		"endorsement store returned error: store operation failed for software components: %s",
		storeError,
	)
	expectedStatus := "failed"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.NotNil(t, body.FailureReason)
	assert.Equal(t, expectedFailureReason, *body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}

func TestHandler_Submit_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mediaType := "application/good+json"
	endo := []byte("some data")

	dm := mock_deps.NewMockIDecoderManager(ctrl)
	dm.EXPECT().
		IsSupportedMediaType(
			gomock.Eq(mediaType),
		).
		Return(true)
	dm.EXPECT().
		Dispatch(
			gomock.Eq(mediaType),
			gomock.Eq(endo),
		).
		Return(&testGoodDecoderResponse, nil)

	sc := mock_deps.NewMockIStoreClient(ctrl)
	sc.EXPECT().
		AddTrustAnchor(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddTrustAnchorRequest{
					TrustAnchor: testGoodDecoderResponse.TrustAnchors[0],
				},
			),
		).
		Return(&testGoodTaRes, nil)
	sc.EXPECT().
		AddSwComponents(
			gomock.Eq(context.TODO()),
			gomock.Eq(
				&endorsement.AddSwComponentsRequest{
					Info: []*endorsement.SwComponent{
						&endorsement.SwComponent{},
					},
				},
			),
		).
		Return(&testGoodSwCompRes, nil)

	h := NewHandler(dm, sc)

	expectedCode := http.StatusOK
	expectedType := ProvisioningSessionMediaType
	expectedStatus := "success"

	w := httptest.NewRecorder()
	g, _ := gin.CreateTestContext(w)

	g.Request, _ = http.NewRequest(http.MethodPost, "/", bytes.NewReader(endo))
	g.Request.Header.Add("Content-Type", mediaType)
	g.Request.Header.Add("Accept", ProvisioningSessionMediaType)

	h.Submit(g)

	var body ProvisioningSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Nil(t, body.FailureReason)
	assert.Equal(t, expectedStatus, body.Status)
}
