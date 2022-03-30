// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/go-plugin"
	"github.com/veraison/common"
)

const (
	// HW Collection
	HW = "HW"
	// SW Collection
	SW = "SW"
	// Edge Collection
	Edge = "Edge"
	// Rel Collection
	Rel = "Rel"
)
const (
	mindepth      = "1"
	maxdepth      = "10"
	to            = ".."
	space         = " "
	newline       = "\n"
	filterPatches = "FILTER link.rel == 'patches' "
	filterUpdates = "FILTER link.rel == 'updates' "
)

// EndorsementStore is the abstraction of Endorsement store
type EndorsementStore struct {
	common.BaseEndorsementBackend
	store Store
}

// Resource describes the software information associated with a tag
type Resource struct {
	Type             string `json:"type"`
	MeasType         string `json:"arm.com-PSAMeasurementType"`
	Description      string `json:"arm.com-PSADescription"`
	MeasurementValue string `json:"arm.com-PSAMeasurementValue"`
	SignerID         string `json:"arm.com-PSASignerId"`
}

// ResourceCollection is used by Payload to describe the details of what may be installed on the device
type ResourceCollection struct {
	Resources []Resource `json:"resource"`
}

// PsaEntity describes the details of the role associated to the tag
type PsaEntity struct {
	Name  string   `json:"entity-name"`
	RegID string   `json:"reg-id"`
	Role  []string `json:"role"`
}

// PsaHardwareRot defines the specific details about the Hardware
type PsaHardwareRot struct {
	ImplementationID string `json:"implementation-id"`
	HwVer            string `json:"hw-ver"`
}

// SoftwareIdentity structure to hold software data
type SoftwareIdentity struct {
	TagID           string               `json:"tag-id"`
	TagVersion      int                  `json:"tag-version"`
	SoftwareName    string               `json:"software-name"`
	SoftwareVersion string               `json:"software-version"`
	Entity          []PsaEntity          `json:"entity"`
	Payload         []ResourceCollection `json:"payload"`
}

// HardwareIdentity is the structure for Arando HW ID container inside Db
type HardwareIdentity struct {
	TagID       string         `json:"tag-id"`
	Entity      []PsaEntity    `json:"entity"`
	HardwareRot PsaHardwareRot `json:"psa-hardware-rot"`
}

// HardwareIdentityWrapper is the top lavel wrapper for fetching from DB
type HardwareIdentityWrapper struct {
	ID      HardwareIdentity `json:"HardwareIdentity"`
	RawData json.RawMessage
}

// SoftwareIdentityWrapper is the top lavel wrapper for fetching from DB
type SoftwareIdentityWrapper struct {
	ID      SoftwareIdentity `json:"SoftwareIdentity"`
	Rawdata json.RawMessage
}

// FirstIndex to be used in fetching resource
const FirstIndex = 0

// GetName method returns the Arango Store name
func (e *EndorsementStore) GetName() string {
	return "ARANGODB"
}

// Init is invoked to initialize the store params
func (e *EndorsementStore) Init(args common.EndorsementBackendParams) error {

	e.Queries = map[string]common.Query{
		"hardware_id":           e.GetHardwareID,
		"software_components":   e.GetSoftwareComponents,          // This should match to all SW Components associated to given measurements
		"all_sw_components":     e.GetAllSoftwareComponents,       // Fetch all SW Components associated to a platform
		"linked_sw_comp_latest": e.GetMostRecentSwVerOfLinkedComp, // For a given measurement, linked to Platform, fetch the most upto date SW component
		"sw_component_latest":   e.GetMostRecentSwVerOfAnySWComp,  // For any given measurement, fetch the most upto date SW component
	}

	// Patch Collection Below
	_, found := args["AltAlgorithm"]
	if found {
		e.Queries["software_components"] = e.GetAltSoftwareComponents
	}

	store, found := args["storeInstance"]
	if !found {
		return fmt.Errorf("DB store instance is not provided inside FetcherParams")
	}
	e.store = store.(Store)
	if !e.store.IsInitialised() {
		return fmt.Errorf("uninitialized DB store provided in FetcherParams")
	}
	ctx := context.Background()
	if err := e.store.Connect(ctx); err != nil {
		return fmt.Errorf("DB connection failed: %w", err)
	}
	return nil
}

// Close shuts down the store
func (e *EndorsementStore) Close() error {
	ctx := context.Background()
	if err := e.store.Connect(ctx); err != nil {
		return err
	}
	if err := e.store.Close(ctx); err != nil {
		return err
	}
	return nil
}

// GetHwIDFromDB queries the DB to get the correct HWID Container for a given platform id
func (e *EndorsementStore) GetHwIDFromDB(platformID string, hwID *HardwareIdentity) error {
	var hardwareInstance HardwareIdentityWrapper
	ctx := context.Background()
	if err := e.store.Connect(ctx); err != nil {
		return err
	}

	queryArgs := map[string]interface{}{
		"platformId": platformID,
	}

	hwCollection, _ := e.store.GetQueryParam(HW)
	query := "FOR d IN " + hwCollection
	query += " FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d"
	log.Printf("input query = %s", query)
	resultList, err := e.store.RunQuery(ctx, query, queryArgs, &hardwareInstance)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for _, resultValue := range resultList {
		hardwareInstance = resultValue.(HardwareIdentityWrapper)
		break
	}
	*hwID = hardwareInstance.ID
	return nil
}

// GetSwIDTagsLinkedToHWTagFromDB fetches a list of SWID's linked to a pltform identified by its unqie HwTag
func (e *EndorsementStore) GetSwIDTagsLinkedToHWTagFromDB(hwTag string) ([]SoftwareIdentity, error) {
	hwCollection, _ := e.store.GetQueryParam(HW)
	hwColTag := hwCollection + "/" + hwTag
	hwColTag = "'" + hwColTag + "'"

	// no need to specify depth as only default depth 1 is required for SWID's linked to HW platform
	edgeCollection, _ := e.store.GetQueryParam(Edge)
	query := "FOR swid, link IN INBOUND " + hwColTag + space + edgeCollection
	query += newline + " FILTER link.rel == 'psa-rot-compound' RETURN swid"

	// Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetSwIDTagsForQueryFromDB failed: %w", err)
	}
	return swidList, nil
}

// GetHardwareID is the query function to fetch the Hardware Id associated to a platform
func (e *EndorsementStore) GetHardwareID(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var hwID HardwareIdentity
	hwMap := make(map[string]string)

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	// Fetch the HW ID container from the DB associated to this platform
	err = e.GetHwIDFromDB(platformID, &hwID)
	if err != nil {
		return nil, fmt.Errorf("for given platform id=%s, failed to fetch the hwID from DB: %w", platformID, err)
	}
	hwMap["hw_id"] = hwID.HardwareRot.HwVer

	// Everything went ok
	result = append(result, hwMap["hw_id"])
	return result, nil
}

// IsMeasurementInList function to check whether given measurement is part of the measurement slice/list,
// return true if in the list, else return false. Return error, if there is some issue
func IsMeasurementInList(measure string, measurements []string) (bool, error) {
	IsPresent := false
	if measure == "" {
		return IsPresent, fmt.Errorf("invalid input NULL measurement")
	}
	if len(measurements) == 0 {
		return IsPresent, fmt.Errorf("invalid query measurement arguments, no measurements present")
	}
	for _, measureValue := range measurements {
		if measure == measureValue {
			IsPresent = true
			break
		}
	}
	return IsPresent, nil
}

// ExtractPlatformID extracts a given platform id from supplied query arguments
func ExtractPlatformID(args common.QueryArgs) (string, error) {
	var platformID string
	platformIDArg, ok := args["platform_id"]
	if !ok {
		return platformID, fmt.Errorf("missing mandatory query argument 'platform_id'")
	}
	switch v := platformIDArg.(type) {
	case string:
		platformID = v
	case []interface{}:
		platformID, ok = v[0].(string)
		if !ok {
			return platformID, fmt.Errorf("unexpected type for 'platform_id'; must be a string")
		}
	default:
		return platformID, fmt.Errorf("unexpected type for 'platform_id'; must be a string; found: %T", v)
	}
	return platformID, nil
}

// ExtractSoftwareComponents function is used to extract the Software Components from the given query parameter
func ExtractSoftwareComponents(args common.QueryArgs) ([]string, error) {
	var measurements []string
	measurementsArg, ok := args["measurements"]
	if !ok {
		return nil, fmt.Errorf("missing mandatory query argument 'measurements'")
	}
	switch v := measurementsArg.(type) {
	case []interface{}:
		for _, elt := range v {
			measure, ok := elt.(string)
			if !ok {
				return nil, fmt.Errorf("unexpected type for 'measurements'; must be a []string")
			}
			measurements = append(measurements, measure)
		}
	case []string:
		measurements = v
	default:
		return nil, fmt.Errorf("unexpected type for 'measurements'; must be []string, found: %T", v)
	}
	return measurements, nil
}

// GetSoftwareComponents implements the query function to fetch the Software Components
func (e *EndorsementStore) GetSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var finalList [][]SoftwareIdentity
	var schemeEndorsements []map[string]string
	matched := false
	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	measurements, err := ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("failed to extract software components from query params: %w", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	finalList, err = e.GetAllswidForPlatform(platformID)
	if err != nil {
		// Handle Error
		return nil, fmt.Errorf("failed to extract swIDs for platform Id=%s %w", platformID, err)
	}
	// Now we have the PSA SWID List associated to this platform
	for _, swidList := range finalList {
		matched = false
		for _, swid := range swidList {
			log.Printf("got SWID with tag = %s", swid.TagID)
			swMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			matched, err = IsMeasurementInList(swMeasurement, measurements)
			if err != nil {
				return nil, fmt.Errorf("error happened while matching measurements: %w ", err)
			}
			if matched {
				// We have the successful swid structure
				// Generate software map from the swid
				swMap := map[string]string{
					"measurement":          swMeasurement,
					"sw_component_type":    swid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
					"sw_component_version": swid.SoftwareVersion,
					"signer_id":            swid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
				}
				schemeEndorsements = append(schemeEndorsements, swMap)
				break
			}
		}
		if !matched {
			log.Printf("no matched component for platform linked swid = %s", swidList[FirstIndex].TagID)
			break
		}
	}
	if !matched {
		return nil, fmt.Errorf("no matched component for platform linked swid")
	}
	result = append(result, schemeEndorsements)
	return result, nil
}

// GetAltSoftwareComponents implements SW optimized alternative to reduce the
// number of query to fetch the software components. It ONLY fetches relations
// for those SWID's whose measurements were not matched to SWID's linked to platform.
// SWID's which have already matched are excluded from querying as one measurement match per
// SW component is expected.
func (e *EndorsementStore) GetAltSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var swidList []SoftwareIdentity
	var schemeEndorsements []map[string]string
	var hwID HardwareIdentity

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	measurements, err := ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("failed to extract software components from query params: %w", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	// Fetch the HW ID container from the DB associated to this platform
	err = e.GetHwIDFromDB(platformID, &hwID)
	if err != nil {
		return nil, fmt.Errorf("for given platform id=%s, failed to fetch the hwID from DB: %w", platformID, err)
	}

	hwTag := hwID.TagID
	// Fetch the SWID List linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(hwTag)
	if err != nil {
		return nil, fmt.Errorf("query GetAllSoftwareComponents failed: %w", err)
	}
	var unmatchedSwid []SoftwareIdentity

	// Now we have the PSA SWID List associated to this platform
	for _, swid := range swidList {
		log.Printf("got SWID with tag = %s", swid.TagID)
		swMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
		matched, err := IsMeasurementInList(swMeasurement, measurements)
		if err != nil {
			return nil, fmt.Errorf("error happened while matching measurements: %w ", err)
		}
		if matched {
			// We have the successful swid structure
			// Generate following map from swid
			swMap := map[string]string{
				"measurement":          swMeasurement,
				"sw_component_type":    swid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
				"sw_component_version": swid.SoftwareVersion,
				"signer_id":            swid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
			}
			schemeEndorsements = append(schemeEndorsements, swMap)
		} else {
			unmatchedSwid = append(unmatchedSwid, swid)
		}
	}
	// Check Unmatched SWID's for their patches and updates
	for _, swid := range unmatchedSwid {
		// Give me Every SWID linked to this SWID in the DB
		log.Printf("umatched SWID = %s", swid.TagID)
		relList, err := e.GetAllSwRelForBaseTag(swid.TagID)
		if err != nil {
			return nil, fmt.Errorf("error occurred while fetching relations: %w", err)
		}
		for _, relswid := range relList {
			log.Printf("got linked SWID with tag = %s", relswid.TagID)
			swMeasurement := relswid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			matched, err := IsMeasurementInList(swMeasurement, measurements)
			if err != nil {
				return nil, fmt.Errorf("error happened while matching measurements: %w", err)
			}
			if matched {
				// We have the successful swid structure here
				// Generate following map from swid
				swMap := map[string]string{
					"measurement":          swMeasurement,
					"sw_component_type":    relswid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
					"sw_component_version": relswid.SoftwareVersion,
					"signer_id":            relswid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
				}
				schemeEndorsements = append(schemeEndorsements, swMap)
			}
		}
	}
	result = append(result, schemeEndorsements)
	return result, nil
}

// GetSwIDTagsForQueryFromDB is the most generic function that is been used to fetch
// SWID's based on given query as an argument to the function.
// Please note returned swid list contains only naked nodes, it does not contain any links/relations among swid.
func (e *EndorsementStore) GetSwIDTagsForQueryFromDB(query string) ([]SoftwareIdentity, error) {
	var swidList []SoftwareIdentity
	var swDocWrap SoftwareIdentityWrapper
	ctx := context.Background()

	if err := e.store.Connect(ctx); err != nil {
		return nil, err
	}

	log.Printf("supplied query = %s", query)

	swDocWrapList, err := e.store.RunQuery(ctx, query, nil, &swDocWrap)
	if err != nil {
		return nil, fmt.Errorf("query failed to fetch swIDs: %w", err)
	}

	for _, swDocWrap := range swDocWrapList {
		softwareInstance := swDocWrap.(SoftwareIdentityWrapper)
		swidList = append(swidList, softwareInstance.ID)
	}
	return swidList, nil
}

// GetAllSoftwareComponents implements the query function to fetch ALL software components associated
// to a platform.  What ALL means is that include software components which are linked to
// platform id and for each of these linked softwarecomponents, fetch associated patches and software
// updates presentin the DB.
func (e *EndorsementStore) GetAllSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var finalList [][]SoftwareIdentity
	var schemeEndorsements []map[string]string

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	finalList, err = e.GetAllswidForPlatform(platformID)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract swid list for a platform Id=%s %w", platformID, err)
	}

	for _, swidList := range finalList {
		// Give me Every SWID linked to this SWID in the DB
		for _, swid := range swidList {
			log.Printf("fetched SWID = %s", swid.TagID)
			swMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			swMap := map[string]string{
				"measurement":          swMeasurement,
				"sw_component_type":    swid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
				"sw_component_version": swid.SoftwareVersion,
				"signer_id":            swid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
			}
			schemeEndorsements = append(schemeEndorsements, swMap)
		}
	}
	result = append(result, schemeEndorsements)
	return result, nil
}

// GetAllswidForPlatform fetches all the SWID's for a given platformID
func (e *EndorsementStore) GetAllswidForPlatform(platformID string) ([][]SoftwareIdentity, error) {
	var hwID HardwareIdentity
	var finalList [][]SoftwareIdentity

	// Fetch the HW ID container from the DB associated to this platform
	err := e.GetHwIDFromDB(platformID, &hwID)
	if err != nil {
		return nil, fmt.Errorf("for given platform id=%s, failed to fetch the hwID from DB: %w", platformID, err)
	}

	hwTag := hwID.TagID
	// Fetch the SWID List linked to this HW tag, based on Verif scheme
	swidList, err := e.GetSwIDTagsLinkedToHWTagFromDB(hwTag)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swIDs linked to hwTag=%s reason: %w", hwTag, err)
	}
	for _, swid := range swidList {
		log.Printf("fetched SWID = %s", swid.TagID)
		var indexSwidList []SoftwareIdentity

		// first swid is always the one linked with platform
		indexSwidList = append(indexSwidList, swid)

		// Get every SWID linked to the base SWID in the DB
		relList, err := e.GetAllSwRelForBaseTag(swid.TagID)
		if err != nil {
			return nil, fmt.Errorf("could not fetch all sw relations for base sw tag=%s: %w", swid.TagID, err)
		}
		indexSwidList = append(indexSwidList, relList...)
		finalList = append(finalList, indexSwidList)
	}
	return finalList, nil
}

// GetAllSwRelForBaseTag fetches all sw patches and updates for a given SWID
func (e *EndorsementStore) GetAllSwRelForBaseTag(swIDTag string) ([]SoftwareIdentity, error) {
	var swidList []SoftwareIdentity
	swCollection, _ := e.store.GetQueryParam(SW)
	swColTag := swCollection + "/" + swIDTag
	swColTag = "'" + swColTag + "'"

	query := "FOR swid IN " + mindepth + to + maxdepth + " ANY " + swColTag + space
	relCollection, _ := e.store.GetQueryParam(Rel)
	query += relCollection + newline + " RETURN swid"
	log.Printf("constructed query = %s", query)

	// Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetAllSwRelForBaseTag=%s failed: %w", query, err)
	}
	return swidList, nil
}

// GetAllSwPatchesForSwIDTag fetches all sw patches for a given SWID
func (e *EndorsementStore) GetAllSwPatchesForSwIDTag(swIDTag string) ([]SoftwareIdentity, error) {
	var swidList []SoftwareIdentity
	swCollection, _ := e.store.GetQueryParam(SW)
	relCollection, _ := e.store.GetQueryParam(Rel)
	swColTag := swCollection + "/" + swIDTag
	swColTag = "'" + swColTag + "'"

	query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY " + swColTag + space
	query += relCollection + newline + filterPatches + "RETURN swid"
	log.Printf("constructed query = %s", query)

	// Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetAllSwPatchesForSwIDTag=%s failed: %w", query, err)
	}
	return swidList, nil
}

// GetAllSwUpdatesForSwIDTag fetches all sw updates for a given SWID
func (e *EndorsementStore) GetAllSwUpdatesForSwIDTag(swIDTag string) ([]SoftwareIdentity, error) {
	var swidList []SoftwareIdentity
	swCollection, _ := e.store.GetQueryParam(SW)
	swColTag := swCollection + "/" + swIDTag
	swColTag = "'" + swColTag + "'"

	relCollection, _ := e.store.GetQueryParam(Rel)
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY " + swColTag + space
	query += relCollection + newline + filterUpdates + "RETURN swid"
	log.Printf("constructed query = %s", query)

	// Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetAllSwUpdatesForSwIDTag=%s failed: %w", query, err)
	}
	return swidList, nil
}

// GetPatchListForBaseSwVersion fetches only the list of forward patches for a specific base revision of SWID
func (e *EndorsementStore) GetPatchListForBaseSwVersion(SwTag string) ([]SoftwareIdentity, error) {
	swCollection, _ := e.store.GetQueryParam(SW)
	relCollection, _ := e.store.GetQueryParam(Rel)
	swColTag := swCollection + "/" + SwTag
	swColTag = "'" + swColTag + "'"
	// SWID(C) --patches--> SWID(B) --patches--> SWID(A), returns B, C for A passed in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
	query += swColTag + space + relCollection + newline
	query += "PRUNE link.rel == 'updates'" + newline
	query += filterPatches + "RETURN swid"
	log.Printf("Constructed Query = %s", query)

	// Fetch the SWID List associated linked this tag
	PatchList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetPatchListForBaseSwVersion=%s failed: %w", query, err)
	}
	return PatchList, nil
}

// GetBaseSwVersionFromPatch fetches only the Base SWID associated to this patch version
func (e *EndorsementStore) GetBaseSwVersionFromPatch(Patchswid SoftwareIdentity) (SoftwareIdentity, error) {
	var baseswid SoftwareIdentity
	var swidList []SoftwareIdentity
	swCollection, _ := e.store.GetQueryParam(SW)
	relCollection, _ := e.store.GetQueryParam(Rel)
	swColTag := swCollection + "/" + Patchswid.TagID
	swColTag = "'" + swColTag + "'"
	// SWID(C) --patches--> SWID(B) --patches--> SWID(A), returns A when C or B supplied in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " OUTBOUND "
	query += swColTag + space + relCollection + newline
	query += "PRUNE link.rel == 'updates'" + newline
	query += filterPatches + "RETURN swid"
	log.Printf("constructed query = %s", query)

	// Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return baseswid, fmt.Errorf("query GetBaseSwVersionFromPatch=%s failed: %w", query, err)
	}
	if len(swidList) == 0 {
		// No base to the Input SWID, hence the patch itself is the Base
		return Patchswid, nil
	}
	baseswid = swidList[len(swidList)-1]
	return baseswid, nil
}

// GetUpdatesListForBaseSwVersion fetches only the list of forward updates for a specific base revision of SWID
func (e *EndorsementStore) GetUpdatesListForBaseSwVersion(Baseswid SoftwareIdentity) ([]SoftwareIdentity, error) {
	swCollection, _ := e.store.GetQueryParam(SW)
	relCollection, _ := e.store.GetQueryParam(Rel)
	swColTag := swCollection + "/" + Baseswid.TagID
	swColTag = "'" + swColTag + "'"
	// SWID(C) --updates--> SWID(B) --updates--> SWID(A), returns B, C for A passed in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
	query += swColTag + space + relCollection + newline
	query += "PRUNE link.rel == 'patches'" + newline
	query += filterUpdates + "RETURN swid"

	log.Printf("constructed query = %s", query)

	// Fetch the SWID List associated linked this tag
	UpdateList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("query GetUpdatesListForBaseSwVersion=%s failed: %w", query, err)
	}
	if len(UpdateList) == 0 {
		// No update list to the Input SWID, hence the given update is most recent
		UpdateList = append(UpdateList, Baseswid)
	}
	return UpdateList, nil
}

// GetMostRecentSwVerOfLinkedComp for a given Platform ID and
// measurement for a SW Component, linked to HW Platform
// fetch most recent SW version to a System.
func (e *EndorsementStore) GetMostRecentSwVerOfLinkedComp(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var measurements []string
	var result []interface{}
	var schemeEndorsements []map[string]string
	var swidList []SoftwareIdentity
	var hwID HardwareIdentity

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	measurements, err = ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("failed to extract software components from query params: %w", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	err = e.GetHwIDFromDB(platformID, &hwID)
	if err != nil {
		return nil, fmt.Errorf("for given platform id=%s, failed to fetch the hwID from DB %w", platformID, err)
	}

	hwTag := hwID.TagID
	// Fetch the swid list linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(hwTag)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swid list for hwTag=%s,  %w", hwTag, err)
	}
	// Only fetch the first measurement
	meas := measurements[0]
	requiredswid, matched := checkMeasMatchInSwIDList(swidList, meas)
	if !matched {
		return nil, fmt.Errorf("query failed platform identity=%s, has no matching measurements", platformID)
	}

	// Now we have the required swid, use it to find the best SW version to return
	// Note by design it is assured that linked SW component (to platform) is always the base.
	latestSwid := requiredswid
	updateList, err := e.GetUpdatesListForBaseSwVersion(latestSwid)
	if err != nil {
		return nil, fmt.Errorf("query failed as unable to fetch update list for base sw version tag %s, %w", latestSwid.TagID, err)
	}
	if (len(updateList)) != 0 {
		latestSwid = updateList[len(updateList)-1]
	}
	// If Update List present new latestSwid becomes root of update
	// get the best patch on this list
	patchList, err := e.GetPatchListForBaseSwVersion(latestSwid.TagID)
	if err != nil {
		return nil, fmt.Errorf("query failed as unable to fetch patch list for base sw version tag %s, %w", latestSwid.TagID, err)
	}
	if (len(patchList)) != 0 {
		latestSwid = patchList[len(patchList)-1]
	}
	swMeasurement := latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue

	// We have the successful SWID structure here
	// Generate the map for the given swid
	swMap := map[string]string{
		"measurement":          swMeasurement,
		"sw_component_type":    latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
		"sw_component_version": latestSwid.SoftwareVersion,
		"signer_id":            latestSwid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
	}
	schemeEndorsements = append(schemeEndorsements, swMap)
	result = append(result, schemeEndorsements)
	return result, nil
}

// checkMeasMatchInSwIDList is a utility function which checks for
// given measurement to be present in the SWID List
func checkMeasMatchInSwIDList(swidList []SoftwareIdentity, meas string) (SoftwareIdentity, bool) {
	var locSwID SoftwareIdentity
	for _, swid := range swidList {
		// Check for SWID Measurement to match the incoming measurement
		if swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue == meas {
			// Got the desired SWID
			return swid, true
		}
	}
	return locSwID, false
}

// GetMostRecentSwVerOfAnySWComp is the Master Blaster
// For a given Platform ID and measurement for a SW Component, which itself could be
// any patch or an update, fetch most recent SW version (recent patch on recent update branch)
func (e *EndorsementStore) GetMostRecentSwVerOfAnySWComp(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var measurements []string
	var result []interface{}
	var schemeEndorsements []map[string]string
	var swidList []SoftwareIdentity
	var relList []SoftwareIdentity
	var hwID HardwareIdentity
	var matched, patchmatched bool

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("failed to extract platform_id from query params: %w", err)
	}
	measurements, err = ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("failed to extract software components from query params: %w", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	err = e.GetHwIDFromDB(platformID, &hwID)
	if err != nil {
		return nil, fmt.Errorf("for given platform id=%s, failed to fetch the hwID from DB: %w", platformID, err)
	}

	hwTag := hwID.TagID
	// Fetch the SWID List linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(hwTag)
	if err != nil {
		return nil, fmt.Errorf("query failed unable to fetch swid tags linked to hwTag for hwTag=%s: %w", hwTag, err)
	}
	// Only fetch the first measurement
	meas := measurements[0]
	requiredswid, matched := checkMeasMatchInSwIDList(swidList, meas)
	if !matched {
		/* ok, the given measurement is definitely not linked to platform,
		check the Patches and Updates List for each swid, for a match? */
		for _, swid := range swidList {
			relList, err = e.GetAllSwPatchesForSwIDTag(swid.TagID)
			if err != nil {
				return nil, fmt.Errorf("query failed unable to fetch all sw patches for swTag=%s: %w", swid.TagID, err)
			}
			requiredswid, patchmatched = checkMeasMatchInSwIDList(relList, meas)
			matched = patchmatched
			if matched {
				break
			} else { /* Given measurement does not match patches, check the updates */
				relList, err = e.GetAllSwUpdatesForSwIDTag(swid.TagID)
				if err != nil {
					return nil, fmt.Errorf("query failed unable to fetch all sw updates for swTag=%s: %w", swid.TagID, err)
				}
				requiredswid, matched = checkMeasMatchInSwIDList(relList, meas)
				if matched {
					break
				}
			}
		}
	}
	if !matched {
		return nil, fmt.Errorf("query failed for platform identity = %s, supplied measurement not in DB: ", platformID)
	}
	// Now we have the required swid, use it to find the best SW version to return
	if patchmatched { // for a patch walk to its base
		requiredswid, err = e.GetBaseSwVersionFromPatch(requiredswid)
		if err != nil {
			return nil, fmt.Errorf("query failed as unable to locate base sw version for patch swid tag %s: %w", requiredswid.TagID, err)
		}
	}
	latestSwid := requiredswid
	// fetch the most recent update on this base
	updateList, err := e.GetUpdatesListForBaseSwVersion(latestSwid)
	if err != nil {
		return nil, fmt.Errorf("query failed as unable to fetch update list for base sw version tag %s: %w", latestSwid.TagID, err)
	}
	if (len(updateList)) != 0 {
		latestSwid = updateList[len(updateList)-1]
	}
	// If Update List present new latestSwid becomes root of update
	// get the best patch on this list
	patchList, err := e.GetPatchListForBaseSwVersion(latestSwid.TagID)
	if err != nil {
		return nil, fmt.Errorf("query failed as unable to fetch patch list for base sw version tag %s, %w", latestSwid.TagID, err)
	}
	if (len(patchList)) != 0 {
		latestSwid = patchList[len(patchList)-1]
	}
	swMeasurement := latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
	swMap := map[string]string{
		"measurement":          swMeasurement,
		"sw_component_type":    latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
		"sw_component_version": latestSwid.SoftwareVersion,
		"signer_id":            latestSwid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
	}
	schemeEndorsements = append(schemeEndorsements, swMap)
	result = append(result, schemeEndorsements)
	return result, nil
}
func main() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "VERAISON_PLUGIN",
		MagicCookieValue: "VERAISON",
	}

	var pluginMap = map[string]plugin.Plugin{
		"endorsementstore": &common.EndorsementBackendPlugin{
			Impl: &EndorsementStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
