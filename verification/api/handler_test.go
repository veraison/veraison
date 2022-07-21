// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"
	mock_deps "github.com/veraison/services/verification/api/mocks"
)

const (
	sessionURIRegexp = `^session/[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`
)

var (
	testSupportedMediaTypeA = `application/eat_cwt; profile=http://arm.com/psa/2.0.0`
	testSupportedMediaTypeB = `application/eat_cwt; profile=PSA_IOT_PROFILE_1`
	testSupportedMediaTypeC = `application/psa-attestation-token`
	testSupportedMediaTypes = []string{
		testSupportedMediaTypeA,
		testSupportedMediaTypeB,
		testSupportedMediaTypeC,
	}
	testSupportedMediaTypesString = strings.Join(testSupportedMediaTypes, ", ")
	testUnsupportedMediaType      = "application/unknown-evidence-format+json"
	testJSONBody                  = `{ "k": "v" }`
	testSession                   = `{
	"status": "waiting",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=http://arm.com/psa/2.0.0",
		"application/eat_cwt;profile=PSA_IOT_PROFILE_1",
		"application/psa-attestation-token"
	]
}`
	testFailedSession = `{
	"status": "failed",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=http://arm.com/psa/2.0.0",
		"application/eat_cwt;profile=PSA_IOT_PROFILE_1",
		"application/psa-attestation-token"
	],
	"evidence": {
		"type":"application/eat_cwt; profile=http://arm.com/psa/2.0.0",
		"value":"eyAiayI6ICJ2IiB9"
	}
}`
	testProcessingSession = `{
	"status": "processing",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=http://arm.com/psa/2.0.0",
		"application/eat_cwt;profile=PSA_IOT_PROFILE_1",
		"application/psa-attestation-token"
	],
	"evidence": {
		"type":"application/eat_cwt; profile=http://arm.com/psa/2.0.0",
		"value":"eyAiayI6ICJ2IiB9"
	}
}`
	testCompleteSession = `{
	"status": "complete",
	"nonce": "mVubqtg3Wa5GSrx3L/2B99cQU2bMQFVYUI9aTmDYi64=",
	"expiry": "2022-07-13T13:50:24.520525+01:00",
	"accept": [
		"application/eat_cwt;profile=http://arm.com/psa/2.0.0",
		"application/eat_cwt;profile=PSA_IOT_PROFILE_1",
		"application/psa-attestation-token"
	],
	"evidence": {
		"type":"application/eat_cwt; profile=http://arm.com/psa/2.0.0",
		"value":"eyAiayI6ICJ2IiB9"
	},
	"result": "e30="
}`
	testUUIDString     = "5c5bd88b-c922-482b-ad9f-097e187b42a1"
	testUUID           = uuid.MustParse(testUUIDString)
	testResult         = `{}`
	testNewSessionURL  = "/challenge-response/v1/newSession"
	testSessionBaseURL = "/challenge-response/v1/session"
)

func TestHandler_NewChallengeResponse_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func testHandler_NewChallengeResponse_BadNonce(t *testing.T, queryParams url.Values, expectedErr string) {
	h := &Handler{}

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: expectedErr,
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = queryParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_NewChallengeResponse_AmbiguousQueryParameters(t *testing.T) {
	q := url.Values{}
	q.Add("nonce", "n")
	q.Add("nonceSize", "1")

	expectedErr := "failed handling nonce request: nonce and nonceSize are mutually exclusive"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceSizeTooBig(t *testing.T) {
	q := url.Values{}
	q.Add("nonceSize", "257")

	expectedErr := "failed handling nonce request: nonceSize must be in range 1..256"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceSizeIsZero(t *testing.T) {
	q := url.Values{}
	q.Add("nonceSize", "0")

	expectedErr := "failed handling nonce request: nonce size cannot be 0"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NonceInvalidB64(t *testing.T) {
	q := url.Values{}
	q.Add("nonce", "^^^^")

	expectedErr := "failed handling nonce request: nonce must be valid base64"

	testHandler_NewChallengeResponse_BadNonce(t, q, expectedErr)
}

func TestHandler_NewChallengeResponse_NoNonceParameters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, "/challenge-response/v1/newSession", http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Len(t, body.Nonce, int(ConfigNonceSize))
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_NonceParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting
	expectedNonce := []byte("nonce")

	qParams := url.Values{}
	// b64("nonce") => "bm9uY2U="
	qParams.Add("nonce", "bm9uY2U=")

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Equal(t, expectedNonce, body.Nonce)
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_NonceSizeParameter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes)

	h := NewHandler(sm, v)

	expectedCode := http.StatusCreated
	expectedType := ChallengeResponseSessionMediaType
	expectedLocationRE := sessionURIRegexp
	expectedSessionStatus := StatusWaiting
	expectedNonceSize := 32

	qParams := url.Values{}
	qParams.Add("nonceSize", strconv.Itoa(expectedNonceSize))

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body ChallengeResponseSession
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Regexp(t, expectedLocationRE, w.Result().Header.Get("Location"))
	assert.Len(t, body.Nonce, expectedNonceSize)
	assert.Nil(t, body.Evidence)
	assert.Nil(t, body.Result)
	assert.Equal(t, expectedSessionStatus, body.Status)
}

func TestHandler_NewChallengeResponse_SetSessionFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sessionManagerError := "session manager says: doh!"

	expectedCode := http.StatusInternalServerError
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: sessionManagerError,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		SetSession(gomock.Any(), tenantID, gomock.Any(), ConfigSessionTTL).
		Return(errors.New(sessionManagerError))

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes)

	h := NewHandler(sm, v)

	qParams := url.Values{}
	qParams.Add("nonceSize", "32")

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, testNewSessionURL, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.URL.RawQuery = qParams.Encode()

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, url, http.NoBody)
	req.Header.Set("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_unsupported_evidence_format(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	verifierError := "no active plugin found for " + testUnsupportedMediaType

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusUnsupportedMediaType
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Unsupported Media Type",
		Status: http.StatusUnsupportedMediaType,
		Detail: verifierError,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		SupportedMediaTypes().
		Return(testSupportedMediaTypes)
	v.EXPECT().
		IsSupportedMediaType(testUnsupportedMediaType).
		Return(false)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testUnsupportedMediaType)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, testSupportedMediaTypesString, w.Result().Header.Get("Accept"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_bad_session_id_url(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badPath := path.Join(testSessionBaseURL, "1234")

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("invalid session id (%s) in path segment: invalid UUID length: 4", badPath),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, badPath, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_session_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	smErr := "session not found"

	expectedCode := http.StatusNotFound
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: smErr,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return(nil, errors.New(smErr))

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathNotFound, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_no_body(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "unable to read evidence from the request body",
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathNotFound, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_SubmitEvidence_process_evidence_failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	vmErr := "enqueueing evidence failed"

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testFailedSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	// we cannot assert on the serialised session object (=> gomock.Any()), but
	// it's not a problem because this is going to be checked anyway when
	// matching the response body
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)
	v.EXPECT().
		ProcessEvidence([]byte(testJSONBody), testSupportedMediaTypeA).
		Return(nil, errors.New(vmErr))

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_SubmitEvidence_process_ok_sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testCompleteSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)
	v.EXPECT().
		ProcessEvidence([]byte(testJSONBody), testSupportedMediaTypeA).
		Return([]byte(testResult), nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_SubmitEvidence_process_ok_async(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusAccepted
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testProcessingSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testSession), nil)
	sm.EXPECT().
		SetSession(testUUID, tenantID, gomock.Any(), ConfigSessionTTL).
		Return(nil)

	v := mock_deps.NewMockIVerifier(ctrl)
	v.EXPECT().
		IsSupportedMediaType(testSupportedMediaTypeA).
		Return(true)
	v.EXPECT().
		ProcessEvidence([]byte(testJSONBody), testSupportedMediaTypeA).
		Return(nil, nil)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPost, pathOK, strings.NewReader(testJSONBody))
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}

func TestHandler_GetSession_UnsupportedAccept(t *testing.T) {
	h := &Handler{}

	url := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusNotAcceptable
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Acceptable",
		Status: http.StatusNotAcceptable,
		Detail: fmt.Sprintf("the only supported output format is %s", ChallengeResponseSessionMediaType),
	}

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, url, http.NoBody)
	req.Header.Set("Accept", "application/unsupported+ber")

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetSession_bad_session_id_url(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	badPath := path.Join(testSessionBaseURL, "1234")

	expectedCode := http.StatusBadRequest
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("invalid session id (%s) in path segment: invalid UUID length: 4", badPath),
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, badPath, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetSession_session_not_found(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathNotFound := path.Join(testSessionBaseURL, testUUIDString)

	smErr := "session not found"

	expectedCode := http.StatusNotFound
	expectedType := "application/problem+json"
	expectedBody := problems.DefaultProblem{
		Type:   "about:blank",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: smErr,
	}

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return(nil, errors.New(smErr))

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, pathNotFound, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	var body problems.DefaultProblem
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.Equal(t, expectedBody, body)
}

func TestHandler_GetSession_ok(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pathOK := path.Join(testSessionBaseURL, testUUIDString)

	expectedCode := http.StatusOK
	expectedType := ChallengeResponseSessionMediaType
	expectedBody := testCompleteSession

	sm := mock_deps.NewMockISessionManager(ctrl)
	sm.EXPECT().
		GetSession(testUUID, tenantID).
		Return([]byte(testCompleteSession), nil)

	v := mock_deps.NewMockIVerifier(ctrl)

	h := NewHandler(sm, v)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, pathOK, http.NoBody)
	req.Header.Set("Accept", ChallengeResponseSessionMediaType)
	req.Header.Set("Content-Type", testSupportedMediaTypeA)

	NewRouter(h).ServeHTTP(w, req)

	body := w.Body.Bytes()

	assert.Equal(t, expectedCode, w.Code)
	assert.Equal(t, expectedType, w.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, expectedBody, string(body))
}
