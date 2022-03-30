// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

//go:build !codeanalysis
// +build !codeanalysis

package main

import (
	"errors"
	"testing"

	mock_deps "arangodbendorsements/mocks"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/veraison/common"
)

type prepareFunc func(int) ([]string, [][]interface{})

const (
	hwTag       = "example.acme.roadrunner-hw-v1-0-0"
	blBaseTag   = "example.acme.roadrunner-sw-bl-v1-0-0"
	pRotBaseTag = "example.acme.roadrunner-sw-prot-v1-0-0"
	aRotBaseTag = "example.acme.roadrunner-sw-arot-v1-0-0"
	appBaseTag  = "example.acme.roadrunner-sw-app-v1-0-0"
	blPTagV1    = "example.acme.roadrunner-sw-bl-v1-0-1"
	pRotPV1     = "example.acme.roadrunner-sw-prot-v1-0-1"
	aRotPV1     = "example.acme.roadrunner-sw-arot-v1-0-1"
	appPV1      = "example.acme.roadrunner-sw-app-v1-0-1"
	armResType  = "arm.com-PSAMeasuredSoftwareComponent"
)
const (
	noError       = ""
	blMeasVal     = "76543210fedcba9817161514131211101f1e1d1c1b1a1916"
	pRotMeasVal   = "76543210fedcba9817161514131211101f1e1d1c1b1a1917"
	aRotMeasVal   = "76543210fedcba9817161514131211101f1e1d1c1b1a1918"
	appMeasVal    = "76543210fedcba9817161514131211101f1e1d1c1b1a1919"
	appPMeasVal   = "76543210fedcba9817161514131211101f1e1d1c1b1a1920"
	umatchMeasVal = "76543210fedcba9817161514131211101f1e1d1c1b1a1AA"
	blPMeasVal    = "76543210fedcba9817161514131211101f1e1d1c1b1a192f"
	blPv2MeasVal  = "76543210fedcba9817161514131211101f1e1d1c1b1a19f2"
	pRotPMeasVal  = "76543210fedcba9817161514131211101f1e1d1c1b1a2017"
	aRotPMeasVal  = "76543210fedcba9817161514131211101f1e1d1c1b1a2018"
	tPlatformID   = "76543210fedcba9817161514131211101f1e1d1c1b1a1918"
	tImplID       = "76543210fedcba9817161514131211101f1e1d1c1b1a1918"
	tSignerID1    = "76543210fedcba9817161514131211101f1e1d1c1b1a1918"
	tSignerID2    = "76543210fedcba9817161514131211101f1e1d1c1b1a1920"
	tSignerID3    = "76543210fedcba9817161514131211101f1e1d1c1b1a1921"
	firstIndex    = 0
	secondIndex   = 1
)

// TestInitStore is the main test function to test store initialization
func TestInitStore(t *testing.T) {
	for _, test := range []struct {
		desc    string
		wantErr string
	}{
		{
			desc: "successful store init",
		},
		{
			desc:    "store init failed",
			wantErr: "uninitialized DB store provided in FetcherParams",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)
			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}

			if test.wantErr == noError {
				ms.EXPECT().IsInitialised().Return(true)
				ms.EXPECT().Connect(gomock.Any()).Return(nil)
			} else {
				ms.EXPECT().IsInitialised().Return(false)
			}

			err := fetcher.Init(argList)
			if err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

		})
	}
}

// TestConnectStore is the main test function to test store connection
func TestConnectStore(t *testing.T) {
	for _, test := range []struct {
		desc    string
		wantErr string
	}{
		{
			desc: "successful store connect",
		},
		{
			desc:    "store connect failed",
			wantErr: "DB connection failed: connection timed out",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)
			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			ms.EXPECT().IsInitialised().Return(true)
			if test.wantErr == noError {
				ms.EXPECT().Connect(gomock.Any()).Return(nil)
			} else {
				ms.EXPECT().Connect(gomock.Any()).Return(errors.New("connection timed out"))
			}

			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

		})
	}
}

// TestQueryHardwareID is the main function to test Query of Hardware ID
func TestQueryHardwareID(t *testing.T) {
	for _, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		wantQuery string
		qRsp      [1]HardwareIdentityWrapper
		wantErr   string
	}{
		{
			desc: "successful query of Hardware ID",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
			},
			wantQuery: "FOR d IN hwid_collection FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d",
			qRsp:      [1]HardwareIdentityWrapper{HardwareIdentityWrapper{ID: HardwareIdentity{TagID: hwTag, Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator"}}}, HardwareRot: PsaHardwareRot{ImplementationID: tImplID, HwVer: "acme-rr-trap"}}}},
		},
		{
			desc: "failed query of Hardware ID",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
			},
			wantQuery: "FOR d IN hwid_collection FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d",
			qRsp:      [1]HardwareIdentityWrapper{},
			wantErr:   "for given platform id=76543210fedcba9817161514131211101f1e1d1c1b1a1918, failed to fetch the hwID from DB: query failed: invalid syntax",
		},
		{
			desc: "invalid platform query, wrong type",
			qArgs: map[string]interface{}{
				"platform_id": 523,
			},
			wantErr: "failed to extract platform_id from query params: unexpected type for 'platform_id'; must be a string; found: int",
		},
		{
			desc: "invalid platform query, incorrect key",
			qArgs: map[string]interface{}{
				"platfor_id": "12345678",
			},
			wantErr: "failed to extract platform_id from query params: missing mandatory query argument 'platform_id'",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			var retErr error
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)
			docList := []interface{}{}

			if test.wantErr != noError {
				retErr = errors.New("invalid syntax")
			} else {
				retErr = nil
			}
			docList = append(docList, test.qRsp[0])
			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
			)
			if test.wantQuery != "" {
				gomock.InOrder(
					ms.EXPECT().Connect(gomock.Any()).Return(nil),
					ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
					ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery), gomock.Any(), gomock.Any()).Return(docList, retErr),
				)
			}

			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetHardwareId query
			qd := common.QueryDescriptor{
				Name:       "hardware_id",
				Args:       test.qArgs,
				Constraint: common.QcOne,
			}

			qr, err := fetcher.GetEndorsements(qd)
			if err == nil {
				assert.Equal(t, 1, len(qr["hardware_id"]),
					"hardware_id constraint of exactly 1 match was not met")
				assert.Equal(t, "acme-rr-trap", qr["hardware_id"][0],
					"hardware_id failed to match")
			} else {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}
		})
	}
}

// prepareRoTdata is the function that prepares the queries to be checked
// and documents to be returned for the RoT data setting
func prepareRoTdata(index int) ([]string, [][]interface{}) {
	var queryArray []string // mark it as Max Queries
	var docArray [][]interface{}
	switch index {
	case firstIndex:
		// For the first array element
		qvar := "FOR d IN hwid_collection FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d"
		queryArray = append(queryArray, qvar)
		docList := []interface{}{}
		var qRsp = [1]HardwareIdentityWrapper{HardwareIdentityWrapper{ID: HardwareIdentity{TagID: hwTag, Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator"}}}, HardwareRot: PsaHardwareRot{ImplementationID: tImplID, HwVer: "acme-rr-trap"}}}}
		docList = append(docList, qRsp[0])
		docArray = append(docArray, docList)
		query := "FOR swid, link IN INBOUND " + "'hwid_collection/example.acme.roadrunner-hw-v1-0-0'" + " " + "edge_verif_scheme"
		query += "\n" + " FILTER link.rel == 'psa-rot-compound' RETURN swid"
		queryArray = append(queryArray, query)
		var swRsp = [4]SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blBaseTag, TagVersion: 0, SoftwareName: "Roadrunner boot loader", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "BL", Description: "TF-M_SHA256MemPreXIP", MeasurementValue: blMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: pRotBaseTag, TagVersion: 0, SoftwareName: "Roadrunner PRoT", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: pRotMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: aRotBaseTag, TagVersion: 0, SoftwareName: "Roadrunner ARoT", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M2", Description: "", MeasurementValue: aRotMeasVal, SignerID: ""}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: appBaseTag, TagVersion: 0, SoftwareName: "Roadrunner App", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M3", Description: "", MeasurementValue: appMeasVal, SignerID: ""}}}}}},
		}
		docList = []interface{}{}
		for _, swid := range swRsp {
			docList = append(docList, swid)
		}
		docArray = append(docArray, docList)

	case secondIndex:
		// For second test, the measurement does not match the Platform RoT, but one of the patches
		// For the first array element
		qvar := "FOR d IN hwid_collection FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d"
		queryArray = append(queryArray, qvar)
		docList := []interface{}{}
		var qRsp = [1]HardwareIdentityWrapper{HardwareIdentityWrapper{ID: HardwareIdentity{TagID: hwTag, Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator"}}}, HardwareRot: PsaHardwareRot{ImplementationID: tImplID, HwVer: "acme-rr-trap"}}}}
		docList = append(docList, qRsp[0])
		docArray = append(docArray, docList)
		query := "FOR swid, link IN INBOUND " + "'hwid_collection/example.acme.roadrunner-hw-v1-0-0'" + " " + "edge_verif_scheme"
		query += "\n" + " FILTER link.rel == 'psa-rot-compound' RETURN swid"
		queryArray = append(queryArray, query)
		var swRsp = [4]SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blBaseTag, TagVersion: 0, SoftwareName: "Roadrunner boot loader", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "BL", Description: "TF-M_SHA256MemPreXIP", MeasurementValue: blMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: pRotBaseTag, TagVersion: 0, SoftwareName: "Roadrunner PRoT", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: pRotMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: aRotBaseTag, TagVersion: 0, SoftwareName: "Roadrunner ARoT", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M2", Description: "", MeasurementValue: aRotMeasVal, SignerID: ""}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: appBaseTag, TagVersion: 0, SoftwareName: "Roadrunner App", SoftwareVersion: "1.0.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M3", Description: "", MeasurementValue: appMeasVal, SignerID: ""}}}}}},
		}
		docList = []interface{}{}
		for _, swid := range swRsp {
			docList = append(docList, swid)
		}
		docArray = append(docArray, docList)
		// return patch node with matching measurements, when queried for all relations
		query = "FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-app-v1-0-0'" + space
		query += "edge_rel_scheme" + newline + " RETURN swid"
		queryArray = append(queryArray, query)
		swRsp1 := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: appPV1, TagVersion: 0, SoftwareName: "Roadrunner App Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M3", Description: "", MeasurementValue: appPMeasVal, SignerID: ""}}}}}},
		}
		docList = []interface{}{}
		for _, swid := range swRsp1 {
			docList = append(docList, swid)
		}
		docArray = append(docArray, docList)
	}
	return queryArray, docArray
}

// prepareGetSoftComp is the function that prepares the queries to be checked
// and documents to be returned for GetSoftwareComponent query
func prepareGetSoftComp(index int) ([]string, [][]interface{}) {
	var queryArray []string // mark it as Max Queries
	var docArray [][]interface{}
	switch index {
	case firstIndex, secondIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)
		qArray := []string{
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-prot-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-arot-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-app-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
		}
		queryArray = append(queryArray, qArray...)
		swRsp := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blPTagV1, TagVersion: 0, SoftwareName: "Roadrunner Boot Loader patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "BL", Description: "TF-M_SHA256MemPreXIP", MeasurementValue: blPMeasVal, SignerID: tSignerID2}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: pRotPV1, TagVersion: 0, SoftwareName: "Roadrunner PRoT Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: pRotPMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: aRotPV1, TagVersion: 0, SoftwareName: "Roadrunner ARoT Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M2", Description: "", MeasurementValue: aRotPMeasVal, SignerID: tSignerID2}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: appPV1, TagVersion: 0, SoftwareName: "Roadrunner App Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M3", Description: "", MeasurementValue: appPMeasVal, SignerID: tSignerID3}}}}}},
		}
		// populate one document per query above
		for _, swdoc := range swRsp {
			docList := []interface{}{}
			docList = append(docList, swdoc)
			docArray = append(docArray, docList)
		}
	}
	return queryArray, docArray
}

// TestQueryGetSoftwareComponents is the main function to test fetch of Software Components,
// based on input measurements
func TestQueryGetSoftwareComponents(t *testing.T) {
	for index, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		prepare   prepareFunc
		wantQuery []string
		qRsp      [][]interface{}
		wantErr   string
	}{
		{
			desc: "successful test the query to fetch matching SW components",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
					aRotMeasVal,
					appMeasVal,
				},
			},
			prepare: prepareGetSoftComp,
		},
		{
			desc: "mis-match in fetching SW components",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
					aRotMeasVal,
					umatchMeasVal,
				},
			},
			prepare: prepareGetSoftComp,
			wantErr: "no matched component for platform linked swid",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			var retErr error
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)
			test.wantQuery, test.qRsp = test.prepare(index)

			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[0]), gomock.Any(), gomock.Any()).Return(test.qRsp[0], retErr),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().GetQueryParam(Edge).Return("edge_verif_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[1]), gomock.Any(), gomock.Any()).Return(test.qRsp[1], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[2]), gomock.Any(), gomock.Any()).Return(test.qRsp[2], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[3]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[4]), gomock.Any(), gomock.Any()).Return(test.qRsp[4], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[5]), gomock.Any(), gomock.Any()).Return(test.qRsp[5], retErr),
			)

			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetAllSoftwareComponents query
			qd := common.QueryDescriptor{
				Name: "software_components",
				Args: test.qArgs,
			}

			qr, err := fetcher.GetEndorsements(qd)
			if err == nil {
				assert.NotEmpty(t, qr["software_components"], "Did not match software components")
				assert.Equal(t, 1, len(qr["software_components"]),
					"software_components constraint of exactly 1 match was not met")
			} else {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}
		})
	}
}

// TestQueryAltSoftwareComponents is the main function to test the alternative software
// implementation for GetSoftComponents
func TestQueryAltSoftwareComponents(t *testing.T) {
	for index, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		prepare   prepareFunc
		wantQuery []string
		qRsp      [][]interface{}
		wantErr   string
	}{
		{
			desc: "successful test alternative algorithm to fetch matching SW components, to platform RoT",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
					aRotMeasVal,
					appMeasVal,
				},
			},
			prepare: prepareRoTdata,
		},
		{
			desc: "successful test alternative algorithm to fetch matching SW components, to one of patches",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
					aRotMeasVal,
					appPMeasVal,
				},
			},
			prepare: prepareRoTdata,
		},
		{
			desc: "failed to fetch SW components, invalid platform id",
			qArgs: map[string]interface{}{
				"platform_id": 123,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
				},
			},
			prepare: prepareRoTdata,
			wantErr: "failed to extract platform_id from query params: unexpected type for 'platform_id'; must be a string; found: int",
		},
		{
			desc: "failed to fetch SW components, invalid platform id key",
			qArgs: map[string]interface{}{
				"platfom_id": 123,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
				},
			},
			prepare: prepareRoTdata,
			wantErr: "failed to extract platform_id from query params: missing mandatory query argument 'platform_id'",
		},
		{
			desc: "failed to fetch SW components, invalid measurements",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []int{
					234,
					432,
				},
			},
			prepare: prepareRoTdata,
			wantErr: "failed to extract software components from query params: unexpected type for 'measurements'; must be []string, found: []int",
		},
		{
			desc: "failed to fetch SW components, invalid key",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurement": []string{
					"123456",
					"678910",
					"122334",
					"162789",
				},
			},
			prepare: prepareRoTdata,
			wantErr: "failed to extract software components from query params: missing mandatory query argument 'measurements'",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			var retErr error
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)

			if test.wantErr != noError {
				retErr = errors.New("invalid syntax")
			} else {
				retErr = nil
			}
			test.wantQuery, test.qRsp = test.prepare(index)
			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
			)
			if test.wantErr == noError {
				gomock.InOrder(
					ms.EXPECT().Connect(gomock.Any()).Return(nil),
					ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
					ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[0]), gomock.Any(), gomock.Any()).Return(test.qRsp[0], retErr),
					ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
					ms.EXPECT().GetQueryParam(Edge).Return("edge_verif_scheme", nil),
					ms.EXPECT().Connect(gomock.Any()).Return(nil),
					ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[1]), gomock.Any(), gomock.Any()).Return(test.qRsp[1], retErr),
				)
				if index == 1 {
					gomock.InOrder(
						ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
						ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
						ms.EXPECT().Connect(gomock.Any()).Return(nil),
						ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[2]), gomock.Any(), gomock.Any()).Return(test.qRsp[2], retErr),
					)
				}
			}
			argList := common.EndorsementBackendParams{
				"AltAlgorithm":  "Normal",
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetSoftwareComponents query
			qd := common.QueryDescriptor{
				Name: "software_components",
				Args: test.qArgs,
			}

			qr, err := fetcher.GetEndorsements(qd)
			if err == nil {
				assert.NotEmpty(t, qr["software_components"], "Did not match software components")
				assert.Equal(t, 1, len(qr["software_components"]),
					"software_components constraint of exactly 1 match was not met")
			} else {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}
		})
	}
}

// prepareAllSoftComp is the function that prepares the queries to be checked
// and documents to be returned for fetching All Software Components query
func prepareAllSoftComp(index int) ([]string, [][]interface{}) {
	var queryArray []string // mark it as Max Queries
	var docArray [][]interface{}
	switch index {
	case firstIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)
		qArray := []string{
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-prot-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-arot-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
			"FOR swid IN " + mindepth + to + maxdepth + " ANY " + "'swid_collection/example.acme.roadrunner-sw-app-v1-0-0'" + space + "edge_rel_scheme" + newline + " RETURN swid",
		}
		queryArray = append(queryArray, qArray...)
		swRsp := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blPTagV1, TagVersion: 0, SoftwareName: "Roadrunner boot loader patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "BL", Description: "TF-M_SHA256MemPreXIP", MeasurementValue: "76543210fedcba9817161514131211101f1e1d1c1b1a1920", SignerID: "tSignerID2"}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: pRotPV1, TagVersion: 0, SoftwareName: "Roadrunner PRoT Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: pRotPMeasVal, SignerID: tSignerID1}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: aRotPV1, TagVersion: 0, SoftwareName: "Roadrunner ARoT Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M2", Description: "", MeasurementValue: aRotPMeasVal, SignerID: "tSignerID2"}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: appPV1, TagVersion: 0, SoftwareName: "Roadrunner App Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M3", Description: "", MeasurementValue: "76543210fedcba9817161514131211101f1e1d1c1b1a2019", SignerID: tSignerID3}}}}}},
		}
		// populate one document per query above
		for _, swdoc := range swRsp {
			docList := []interface{}{}
			docList = append(docList, swdoc)
			docArray = append(docArray, docList)
		}

	}
	return queryArray, docArray
}

// TestQueryAllSoftwareComponents is the main function to test query of all the
// software components based on the supplied input measurements
func TestQueryAllSoftwareComponents(t *testing.T) {
	for index, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		prepare   prepareFunc
		wantQuery []string
		qRsp      [][]interface{}
		wantErr   string
	}{
		{
			desc: "successful test for fetching all sw components for a given measurement",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blMeasVal,
					pRotMeasVal,
					aRotMeasVal,
					appMeasVal,
				},
			},
			prepare: prepareAllSoftComp,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			var retErr error
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)

			if test.wantErr != noError {
				retErr = errors.New("invalid syntax")
			} else {
				retErr = nil
			}
			test.wantQuery, test.qRsp = test.prepare(index)

			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[0]), gomock.Any(), gomock.Any()).Return(test.qRsp[0], retErr),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().GetQueryParam(Edge).Return("edge_verif_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[1]), gomock.Any(), gomock.Any()).Return(test.qRsp[1], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[2]), gomock.Any(), gomock.Any()).Return(test.qRsp[2], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[3]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[4]), gomock.Any(), gomock.Any()).Return(test.qRsp[4], retErr),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[5]), gomock.Any(), gomock.Any()).Return(test.qRsp[5], retErr),
			)

			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetAllSoftwareComponents query
			qd := common.QueryDescriptor{
				Name: "all_sw_components",
				Args: test.qArgs,
			}

			qr, err := fetcher.GetEndorsements(qd)
			assert.Nil(t, err)
			assert.NotEmpty(t, qr["all_sw_components"], "Did not match software components")
			assert.Equal(t, 1, len(qr["all_sw_components"]),
				"all_sw_components constraint of exactly 1 match was not met")
		})
	}
}

// prepareLinkedSoftComp is the function that prepares the queries to be checked
// and documents to be returned for the LinkedSoftwareComponent query
func prepareLinkedSoftComp(index int) ([]string, [][]interface{}) {
	var queryArray []string // mark it as Max Queries
	var docArray [][]interface{}
	switch index {
	case firstIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)
		// First query to fecth the latest of base
		query1 := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
		query1 += "'swid_collection/example.acme.roadrunner-sw-prot-v1-0-0'" + space + "edge_rel_scheme" + newline
		query1 += "PRUNE link.rel == 'patches'" + newline
		query1 += "FILTER link.rel == 'updates' RETURN swid"

		// Next query is to fetch the uptodate patch for the latest base
		query2 := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
		query2 += "'swid_collection/example.acme.roadrunner-sw-prot-v1-2-0'" + space + "edge_rel_scheme" + newline
		query2 += "PRUNE link.rel == 'updates'" + newline
		query2 += "FILTER link.rel == 'patches' RETURN swid"

		queryArray = append(queryArray, query1, query2)
		swRsp := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: "example.acme.roadrunner-sw-prot-v1-2-0", TagVersion: 0, SoftwareName: "Roadrunner PRoT Base V2", SoftwareVersion: "1.2.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: "76543210fedcba9817161514131211101f1e1d1c1b1a2217", SignerID: "76543210fedcba9817161514131211101f1e1d1c1b1a1913"}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: "example.acme.roadrunner-sw-prot-v1-2-2", TagVersion: 0, SoftwareName: "Roadrunner PRoT Patch V2", SoftwareVersion: "1.2.2", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: "76543210fedcba9817161514131211101f1e1d1c1b1a3317", SignerID: tSignerID1}}}}}},
		}
		// populate onde document per query above
		for _, swdoc := range swRsp {
			docList := []interface{}{}
			docList = append(docList, swdoc)
			docArray = append(docArray, docList)
		}
	case secondIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)
	}
	return queryArray, docArray
}

// TestQueryLatestLinkedSwComponent is the main function to test Query of fetching
// software components from supplied measurements linked to platform
func TestQueryLatestLinkedSwComponent(t *testing.T) {
	for index, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		prepare   prepareFunc
		wantQuery []string
		qRsp      [][]interface{}
		wantErr   string
	}{
		{
			desc: "successful query of measurement linked to PSA RoT SWID",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					pRotMeasVal,
				},
			},
			prepare: prepareLinkedSoftComp,
		},
		{
			desc: "unsuccessful query, no measurements for PSA RoT in DB",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					"76543210fedcba9817161514131211101f1e1d1c1b1a19AA",
				},
			},
			prepare: prepareLinkedSoftComp,
			wantErr: "query failed platform identity=76543210fedcba9817161514131211101f1e1d1c1b1a1918, has no matching measurements",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)

			test.wantQuery, test.qRsp = test.prepare(index)

			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[0]), gomock.Any(), gomock.Any()).Return(test.qRsp[0], nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().GetQueryParam(Edge).Return("edge_verif_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[1]), gomock.Any(), gomock.Any()).Return(test.qRsp[1], nil),
			)
			if test.wantErr == noError {
				gomock.InOrder(
					ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
					ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
					ms.EXPECT().Connect(gomock.Any()).Return(nil),
					ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[2]), gomock.Any(), gomock.Any()).Return(test.qRsp[2], nil),
					ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
					ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
					ms.EXPECT().Connect(gomock.Any()).Return(nil),
					ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[3]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], nil),
				)
			}
			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetLinkedSoftwareComponents query
			qd := common.QueryDescriptor{
				Name: "linked_sw_comp_latest",
				Args: test.qArgs,
			}
			qr, err := fetcher.GetEndorsements(qd)
			if err == nil {

				assert.NotEmpty(t, qr["linked_sw_comp_latest"], "Did not match software components")
				assert.Equal(t, 1, len(qr["linked_sw_comp_latest"]),
					"linked_sw_comp_latest constraint of exactly 1 match was not met")
			} else {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}
		})
	}
}

// prepareMostRecentSoftComp is the function that prepares the queries to be checked
// and documents to be returned for fecthing most recent software component query
func prepareMostRecentSoftComp(index int) ([]string, [][]interface{}) {
	var queryArray []string // mark it as Max Queries
	var docArray [][]interface{}
	switch index {
	case firstIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)
		// Set the common part of query first
		query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY "
		query += "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-0'" + space + "edge_rel_scheme" + newline

		// First query is to fetch all the patches
		query1 := query + "FILTER link.rel == 'patches' RETURN swid"

		// Next query is to fetch all the updates
		query2 := query + "FILTER link.rel == 'updates' RETURN swid"

		queryArray = append(queryArray, query1, query2)
		swRsp := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blPTagV1, TagVersion: 0, SoftwareName: "Roadrunner BL Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: "76543210fedcba9817161514131211101f1e1d1c1b1a2217", SignerID: "76543210fedcba9817161514131211101f1e1d1c1b1a1913"}}}}}},
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: "example.acme.roadrunner-sw-bl-v1-2-0", TagVersion: 0, SoftwareName: "Roadrunner BL Base V2", SoftwareVersion: "1.2.0", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: blPv2MeasVal, SignerID: tSignerID1}}}}}},
		}
		// populate one document per query above
		for _, swdoc := range swRsp {
			docList := []interface{}{}
			docList = append(docList, swdoc)
			docArray = append(docArray, docList)
		}
		// Set the common part of query first
		query = "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
		query += "'swid_collection/example.acme.roadrunner-sw-bl-v1-2-0'" + space + "edge_rel_scheme" + newline

		// Query to get the patches
		query1 = query + "PRUNE link.rel == 'patches'" + newline
		query1 += "FILTER link.rel == 'updates' RETURN swid"
		queryArray = append(queryArray, query1)
		// above query is expected to return an empty document slice
		docList := []interface{}{}
		docArray = append(docArray, docList)

		// Query to get the updates
		query2 = query + "PRUNE link.rel == 'updates'" + newline
		query2 += "FILTER link.rel == 'patches' RETURN swid"
		queryArray = append(queryArray, query2)
		// above query is expected to return an empty document slice
		docList = []interface{}{}
		docArray = append(docArray, docList)

	case secondIndex:
		// the following function sets the naked nodes linked to a platform id, through psa-rot
		queryArray, docArray = prepareRoTdata(0)

		query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY "
		query += "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-0'" + space + "edge_rel_scheme" + newline
		query += "FILTER link.rel == 'patches' RETURN swid"

		queryArray = append(queryArray, query)
		swRsp := []SoftwareIdentityWrapper{
			SoftwareIdentityWrapper{ID: SoftwareIdentity{TagID: blPTagV1, TagVersion: 0, SoftwareName: "Roadrunner BL Patch V1", SoftwareVersion: "1.0.1", Entity: []PsaEntity{PsaEntity{Name: "ACME Ltd", RegID: "acme.example", Role: []string{"tagCreator", "aggregator"}}}, Payload: []ResourceCollection{ResourceCollection{Resources: []Resource{Resource{Type: armResType, MeasType: "M1", Description: "", MeasurementValue: blPv2MeasVal, SignerID: "76543210fedcba9817161514131211101f1e1d1c1b1a1913"}}}}}},
		}
		// populate one document per query above
		for _, swdoc := range swRsp {
			docList := []interface{}{}
			docList = append(docList, swdoc)
			docArray = append(docArray, docList)
		}

		query = "FOR swid, link IN " + mindepth + to + maxdepth + " OUTBOUND "
		query += "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-1'" + space + "edge_rel_scheme" + newline
		query += "PRUNE link.rel == 'updates'" + newline
		query += "FILTER link.rel == 'patches' RETURN swid"
		queryArray = append(queryArray, query)
		docList := []interface{}{}
		docArray = append(docArray, docList)

		// Set the common part of query first
		query = "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
		query += "'swid_collection/example.acme.roadrunner-sw-bl-v1-0-1'" + space + "edge_rel_scheme" + newline

		// Query to get the patches
		query1 := query + "PRUNE link.rel == 'patches'" + newline
		query1 += "FILTER link.rel == 'updates' RETURN swid"
		queryArray = append(queryArray, query1)
		// above query is expected to return an empty document slice
		docList = []interface{}{}
		docArray = append(docArray, docList)

		// Query to get the updates
		query2 := query + "PRUNE link.rel == 'updates'" + newline
		query2 += "FILTER link.rel == 'patches' RETURN swid"
		queryArray = append(queryArray, query2)
		// above query is expected to return an empty document slice
		docList = []interface{}{}
		docArray = append(docArray, docList)
	}
	return queryArray, docArray
}

// TestQueryMostRecentSwComp is the main function to test Query of most recent
// software component for a supplied measurement
func TestQueryMostRecentSwComp(t *testing.T) {
	for index, test := range []struct {
		desc      string
		qArgs     map[string]interface{}
		prepare   prepareFunc
		wantQuery []string
		qRsp      [][]interface{}
		wantErr   string
	}{
		{
			desc: "successful query of measurement linked to an update",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blPv2MeasVal,
				},
			},
			prepare: prepareMostRecentSoftComp,
		},
		{
			desc: "successful query of measurement linked to a patch",
			qArgs: map[string]interface{}{
				"platform_id": tPlatformID,
				"measurements": []string{
					blPv2MeasVal,
				},
			},
			prepare: prepareMostRecentSoftComp,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ms := mock_deps.NewMockStore(ctrl)

			test.wantQuery, test.qRsp = test.prepare(index)

			gomock.InOrder(
				ms.EXPECT().IsInitialised().Return(true),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[0]), gomock.Any(), gomock.Any()).Return(test.qRsp[0], nil),
				ms.EXPECT().GetQueryParam(HW).Return("hwid_collection", nil),
				ms.EXPECT().GetQueryParam(Edge).Return("edge_verif_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[1]), gomock.Any(), gomock.Any()).Return(test.qRsp[1], nil),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[2]), gomock.Any(), gomock.Any()).Return(test.qRsp[2], nil),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[3]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], nil),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[4]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], nil),
				ms.EXPECT().GetQueryParam(SW).Return("swid_collection", nil),
				ms.EXPECT().GetQueryParam(Rel).Return("edge_rel_scheme", nil),
				ms.EXPECT().Connect(gomock.Any()).Return(nil),
				ms.EXPECT().RunQuery(gomock.Any(), gomock.Eq(test.wantQuery[5]), gomock.Any(), gomock.Any()).Return(test.qRsp[3], nil),
			)

			argList := common.EndorsementBackendParams{
				"storeInstance": ms,
			}
			fetcher := &EndorsementStore{}
			if err := fetcher.Init(argList); err != nil {
				assert.EqualErrorf(t, err, test.wantErr, "received error got != want (%s, %s)", err.Error(), test.wantErr)
			}

			// set the required query parameter for GetSoftwareComponentLatest query
			qd := common.QueryDescriptor{
				Name: "sw_component_latest",
				Args: test.qArgs,
			}
			qr, err := fetcher.GetEndorsements(qd)
			assert.Nil(t, err)
			assert.NotEmpty(t, qr["sw_component_latest"], "Did not match software components")
			assert.Equal(t, 1, len(qr["sw_component_latest"]),
				"sw_component_latest constraint of exactly 1 match was not met")
		})
	}
}
