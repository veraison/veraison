// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"veraison/common"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/hashicorp/go-plugin"
)

const (
	mindepth = "1"
	maxdepth = "10"
	to       = ".."
	space    = " "
	newline  = "\n"
)

// ArangoDBConnParams holds the dynamic connection information for arangoDB
type ArangoDBConnParams struct {
	paramValid bool
	// HTTP Connection to the Data Base
	conn driver.Connection

	// Client
	client driver.Client

	// DB
	db driver.Database

	// Graph inside the DB
	graph driver.Graph
}

// ArangoEndorsementStore holds the DB specific details
type ArangoEndorsementStore struct {
	common.BaseEndorsementStore
	StoreName      string
	GraphName      string
	Login          string
	Password       string
	HwCollection   string
	SwCollection   string
	EdgeCollection string
	RelCollection  string
}

// PsaLinkObject represents an Edge Collection container in the DB
type PsaLinkObject struct {
	From     string `json:"_from"`
	To       string `json:"_to"`
	Relation string `json:"rel"`
}

//Resource describes the software information associated with a tag
type Resource struct {
	Type             string `json:"type"`
	MeasType         string `json:"arm.com-PSAMeasurementType"`
	Description      string `json:"arm.com-PSADescription"`
	MeasurementValue string `json:"arm.com-PSAMeasurementValue"`
	SignerID         string `json:"arm.com-PSASignerId"`
}

//ResourceCollection is used by Payload to describe the details of what may be installed on the device
type ResourceCollection struct {
	Resources []Resource `json:"resource"`
}

//PsaEntity describes the details of the role associated to the tag
type PsaEntity struct {
	Name  string   `json:"entity-name"`
	RegID string   `json:"reg-id"`
	Role  []string `json:"role"`
}

//PsaHardwareRot defines the specific details about the Hardware
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

// HardwareIdentityWrapper is the top lavel wrapper for fetching from arangoDB
type HardwareIdentityWrapper struct {
	ID      HardwareIdentity `json:"HardwareIdentity"`
	RawData json.RawMessage
}

// SoftwareIdentityWrapper is the top lavel wrapper for fetching from arangoDB
type SoftwareIdentityWrapper struct {
	ID      SoftwareIdentity `json:"SoftwareIdentity"`
	Rawdata json.RawMessage
}

//FirstIndex to be used in fetching resource
const FirstIndex = 0

// GetName method returns the Arango Store name
func (e *ArangoEndorsementStore) GetName() string {
	return "arangodb"
}

// Init is invoked to initialise the Arango Store params
func (e *ArangoEndorsementStore) Init(args common.EndorsementStoreParams) error {
	storeName, found := args["storeName"]
	if !found {
		return fmt.Errorf("ArangoDB Store Name is not specified inside FetcherParams")
	}
	e.StoreName = storeName

	graphName, found := args["graphName"]
	if !found {
		return fmt.Errorf("ArangoDB Graph Name is not specified inside FetcherParams")
	}
	e.GraphName = graphName

	login, found := args["login"]
	if !found {
		return fmt.Errorf("ArangoDB Login Name is not specified inside FetcherParams")
	}
	e.Login = login

	password, found := args["password"]
	if !found {
		return fmt.Errorf("ArangoDB Password is not specified inside FetcherParams")
	}
	e.Password = password

	// HW Collection below
	hwCollection, found := args["hwCollection"]
	if !found {
		return fmt.Errorf("ArangoDB HWID Collection is not specified inside FetcherParams")
	}
	e.HwCollection = hwCollection

	// SW Collection below
	swCollection, found := args["swCollection"]
	if !found {
		return fmt.Errorf("ArangoDB SWID Collection is not specified inside FetcherParams")
	}
	e.SwCollection = swCollection

	// Edge Collection Below
	edgeCollection, found := args["edgeCollection"]
	if !found {
		return fmt.Errorf("ArangoDB Edge Collection is not specified inside FetcherParams")
	}
	e.EdgeCollection = edgeCollection

	// Patch Collection Below
	relCollection, found := args["relCollection"]
	if !found {
		log.Printf("ArangoDB Patch Collection is not specified inside FetcherParams")
	}
	e.RelCollection = relCollection

	e.Queries = map[string]common.Query{
		"hardware_id":           e.GetHardwareID,
		"software_components":   e.GetSoftwareComponents,          // This should match to all SW Components associated to given measurements
		"all_sw_components":     e.GetAllSoftwareComponents,       // Fetch all SW Components associated to a platform
		"linked_sw_comp_latest": e.GetMostRecentSwVerOfLinkedComp, // For a given measurement, linked to Platform, fetch the most upto date SW component
		"sw_component_latest":   e.GetMostRecentSwVerOfAnySWComp,  // For any given measurement, fetch the most upto date SW component
	}

	// Patch Collection Below
	_, found = args["AltAlgorithm"]
	if found {
		e.Queries["software_components"] = e.GetAltSoftwareComponents
	}

	arangoConn := new(ArangoDBConnParams)
	ctx := context.Background()
	res, err := e.ConnectToArangoDB(ctx, arangoConn)
	if (err != nil) || (res != true) {
		return fmt.Errorf("Arango DB Connection Failed %v", err)
	}
	return nil
}

// ConnectToArangoDB is responsible for making connection to the arangoDB
func (e *ArangoEndorsementStore) ConnectToArangoDB(ctx context.Context, arangoConn *ArangoDBConnParams) (connStatus bool, err error) {
	// Create an HTTP Connection First to the Client

	ok := true
	arangoConn.conn, err = http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		return false, fmt.Errorf("Failed to create HTTP connection: %v", err)
	}

	// Create a client
	arangoConn.client, err = driver.NewClient(driver.ClientConfig{
		Connection:     arangoConn.conn,
		Authentication: driver.BasicAuthentication(e.Login, e.Password),
	})

	// Check if Database with a given name exists or not?
	ok, err = arangoConn.client.DatabaseExists(ctx, e.StoreName)
	if ok {
		arangoConn.db, err = arangoConn.client.Database(ctx, e.StoreName)
		if err != nil {
			return false, fmt.Errorf("Failed to connect to Endorsement database: %v", err)
		}
	}

	// Check if the Graph exists in the DB or not ?
	ok, err = arangoConn.db.GraphExists(ctx, e.GraphName)
	if ok {
		// Connect to the Graph now to get the Graph
		arangoConn.graph, err = arangoConn.db.Graph(ctx, e.GraphName)
		if err != nil {
			return false, fmt.Errorf("Failed to connect to Graph for Endorsement database: %v", err)
		}
	}
	log.Printf("Connection DB and Graph Succeeded!")
	arangoConn.paramValid = true
	return arangoConn.paramValid, nil
}

// Close method is uses to Close the DB
func (e *ArangoEndorsementStore) Close() (err error) {
	ctx := context.Background()
	arangoConn := new(ArangoDBConnParams)
	e.ConnectToArangoDB(ctx, arangoConn)

	if arangoConn.paramValid {
		err := arangoConn.graph.Remove(ctx)
		if err != nil {
			return fmt.Errorf("Unable to remove the graph: %s", err)
		}

		if err = arangoConn.db.Remove(ctx); err != nil {
			return fmt.Errorf("Failed to remove database: %s", err)
		}
		arangoConn.paramValid = false
	}
	return nil
}

//GetHwIDFromArangoDB queries the ArangoDB to get the correct HWID Container for a given platform id
func (e *ArangoEndorsementStore) GetHwIDFromArangoDB(platformID string, HwID *HardwareIdentity) error {
	ctx := context.Background()
	arangoConn := new(ArangoDBConnParams)
	e.ConnectToArangoDB(ctx, arangoConn)
	var HardwareInstance HardwareIdentityWrapper

	bindVars := map[string]interface{}{
		"platformId": platformID,
	}
	query := "FOR d IN " + e.HwCollection
	query += " FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d"
	log.Printf("Input Query = %s", query)

	cursor, err := arangoConn.db.Query(ctx, query, bindVars)
	if err != nil {
		return fmt.Errorf("Query Failed with reason %v", err)
	}
	defer cursor.Close()

	for {
		meta, err := cursor.ReadDocument(ctx, &HardwareInstance)
		log.Printf("Returned from readDocument %v ", err)

		if driver.IsNoMoreDocuments(err) {
			log.Printf("No More Documents Left")
			break
		} else {
			*HwID = HardwareInstance.ID
		}

		log.Printf("Got doc with key '%s' from query\n", meta.Key)
	}
	return nil
}

//GetSwIDTagsLinkedToHWTagFromDB fecthes a list of SWID's linked to a pltform identified by its unqie HwTag
func (e *ArangoEndorsementStore) GetSwIDTagsLinkedToHWTagFromDB(HwTag string) ([]SoftwareIdentity, error) {
	ctx := context.Background()
	arangoConn := new(ArangoDBConnParams)
	e.ConnectToArangoDB(ctx, arangoConn)
	var swidList []SoftwareIdentity

	HwColTag := e.HwCollection + "/" + HwTag
	HwColTag = "'" + HwColTag + "'"

	// no need to specify depth as only default depth 1 is required for SWID's linked to HW platform
	query := "FOR swid, link IN INBOUND " + HwColTag + space + e.EdgeCollection
	query += newline + " FILTER link.rel == 'psa-rot-compound' RETURN swid"

	//Fetch the SWID List associated linked this tag
	swidList, err := e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetSwIDTagsForQueryFromDB failed %v", err)
	}
	return swidList, nil
}

// GetHardwareID is the query function to fetch the Hardware Id associated to a platform
func (e *ArangoEndorsementStore) GetHardwareID(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var HardwareInstance HardwareIdentityWrapper
	arangoConn := new(ArangoDBConnParams)

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	log.Printf("Received Platform Id=%s", platformID)

	bindVars := map[string]interface{}{
		"platformId": platformID,
	}

	ctx := context.Background()
	e.ConnectToArangoDB(ctx, arangoConn)

	query := "FOR d IN " + e.HwCollection
	query += " FILTER d.HardwareIdentity.`psa-hardware-rot`.`implementation-id` == @platformId RETURN d"
	log.Printf("Input Query = %s", query)

	cursor, err := arangoConn.db.Query(ctx, query, bindVars)
	if err != nil {
		return nil, fmt.Errorf("Query Failed with reason %v", err)
	}
	defer cursor.Close()

	hwMap := make(map[string]string)
	for {
		meta, err := cursor.ReadDocument(ctx, &HardwareInstance)
		log.Printf("Returned from readDocument %v ", err)

		if driver.IsNoMoreDocuments(err) {
			log.Printf("No More Documents Left")
			break
		} else {
			hwMap["hw_id"] = HardwareInstance.ID.HardwareRot.HwVer
			log.Printf("Detected HW  Id %s", HardwareInstance.ID.HardwareRot.HwVer)
		}

		log.Printf("Got doc with key '%s' from query\n", meta.Key)
	}
	// Everything went ok
	log.Printf("Returned Hardware id = %s", hwMap["hw_id"])
	result = append(result, hwMap["hw_id"])

	return result, nil
}

// IsMeasurementInList function to check whether given measurement is part of the measurement slice/list,
// return true if in the list, else return false. Return error, if there is some issue
func IsMeasurementInList(measure string, measurements []string) (IsPresent bool, err error) {
	IsPresent = false
	err = nil
	if measure == "" {
		err = fmt.Errorf("Invalid Input NULL measurement")
		return IsPresent, err
	}
	if len(measurements) == 0 {
		err = fmt.Errorf("Invalid Query Measurement arguments, no measurements present")
		return IsPresent, err
	}
	for _, measureValue := range measurements {
		if measure == measureValue {
			IsPresent = true
			break
		}
	}
	return IsPresent, err
}

//ExtractPlatformID function extracts a given platform id from supplied query arguments
func ExtractPlatformID(args common.QueryArgs) (string, error) {
	var platformID string
	platformIDArg, ok := args["platform_id"]
	if !ok {
		return "NULL", fmt.Errorf("Missing mandatory query argument 'platform_id'")
	}
	switch platformIDArg.(type) {
	case string:
		platformID = platformIDArg.(string)
	case []interface{}:
		platformID, ok = platformIDArg.([]interface{})[0].(string)
		if !ok {
			return "NULL", fmt.Errorf("Unexpected type for 'platform_id'; must be a string")
		}
	default:
		t := reflect.TypeOf(platformIDArg)
		return "NULL", fmt.Errorf("Unexpected type for 'platform_id'; must be a string; found: %v", t)
	}
	return platformID, nil
}

// ExtractSoftwareComponents function is used to extract the Software Components from the given query parameter
func ExtractSoftwareComponents(args common.QueryArgs) (meas []string, err error) {
	var measurements []string
	measurementsArg, ok := args["measurements"]
	if !ok {
		return nil, fmt.Errorf("Missing mandatory query argument 'measurements'")
	}
	switch measurementsArg.(type) {
	case []interface{}:
		for _, elt := range measurementsArg.([]interface{}) {
			measure, ok := elt.(string)
			if !ok {
				return nil, fmt.Errorf("Unexpected type for 'measurements'; must be a []string")
			}
			measurements = append(measurements, measure)
		}
	case []string:
		measurements = measurementsArg.([]string)
	default:
		return nil, fmt.Errorf("Unexpected type for 'measurements'; must be a []string")
	}
	return measurements, nil
}

// GetSoftwareComponents implements the query function to fetch the Software Components
func (e *ArangoEndorsementStore) GetSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var finalList [][]SoftwareIdentity
	var schemeEndorsements []map[string]string

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	measurements, err := ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("Failed to extract software compoenents from query params %v", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	finalList, err = e.GetAllswidForPlatform(platformID)
	if err != nil {
		// Handle Error
		return nil, fmt.Errorf("Failed to extract SW Components for the platform Id %v", err)
	}
	// Now we have the PSA SWID List associated to this platform
	for _, swidList := range finalList {
		matched := false
		for _, swid := range swidList {
			log.Printf("Got SWID with Tag = %s", swid.TagID)
			SwMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			matched, err = IsMeasurementInList(SwMeasurement, measurements)
			if err != nil {
				return nil, fmt.Errorf("Error happened while matching measurements %v ", err)
			}
			if matched {
				// We have the successful swid structure
				// Generate software map from the swid
				swMap := map[string]string{
					"measurement":          SwMeasurement,
					"sw_component_type":    swid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
					"sw_component_version": swid.SoftwareVersion,
					"signer_id":            swid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
				}
				schemeEndorsements = append(schemeEndorsements, swMap)
				break
			}
		}
		if !matched {
			log.Printf("No matched component for platform linked swid = %s", swidList[FirstIndex].TagID)
		}
	}
	result = append(result, schemeEndorsements)
	return result, nil
}

//GetAltSoftwareComponents implements SW Optimised Alternative to reduce the
// number of query to fetch the Software Components. It ONLY fetches Relations
// for those SWID's whose measurements were not matched to SWID's linked to platform
// SWID's which have already matched are excluded from querying as one measurement match per
// SW component is expected.
func (e *ArangoEndorsementStore) GetAltSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var swidList []SoftwareIdentity
	var schemeEndorsements []map[string]string
	var HwID HardwareIdentity

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	measurements, err := ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("Failed to extract software compoenents from query params %v", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	// Fetch the HW ID container from the DB associated to this platform
	err = e.GetHwIDFromArangoDB(platformID, &HwID)
	if err != nil {
		return nil, fmt.Errorf("Given platform Identity failed %v", err)
	}

	HwTag := HwID.TagID
	//Fetch the SWID List linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(HwTag)
	if err != nil {
		return nil, fmt.Errorf("Query GetAllSoftwareComponents failed %v", err)
	}
	var unmatchedSwid []SoftwareIdentity

	// Now we have the PSA SWID List associated to this platform
	for _, swid := range swidList {
		log.Printf("Got SWID with Tag = %s", swid.TagID)
		SwMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
		matched, err := IsMeasurementInList(SwMeasurement, measurements)
		if err != nil {
			return nil, fmt.Errorf("Error happened while matching measurements %v ", err)
		}
		if matched {
			// We have the successful swid structure
			// Generate following map from swid
			swMap := map[string]string{
				"measurement":          SwMeasurement,
				"sw_component_type":    swid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
				"sw_component_version": swid.SoftwareVersion,
				"signer_id":            swid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
			}
			schemeEndorsements = append(schemeEndorsements, swMap)
		} else {
			unmatchedSwid = append(unmatchedSwid, swid)
		}
	}
	//Check Unmatched SWID's for their patches and updates
	for _, swid := range unmatchedSwid {
		// Give me Every SWID linked to this SWID in the DB
		log.Printf("Umatched SWID = %s", swid.TagID)
		relList, err := e.GetAllSwRelForBaseTag(swid.TagID)
		if err != nil {
			return nil, fmt.Errorf("Error occured while fetching relations %v", err)
		}
		for _, relswid := range relList {
			log.Printf("Got Linked SWID with Tag = %s", relswid.TagID)
			SwMeasurement := relswid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			matched, err := IsMeasurementInList(SwMeasurement, measurements)
			if err != nil {
				return nil, fmt.Errorf("Error happened while matching measurements %v ", err)
			}
			if matched {
				// We have the successful swid structure here
				// Generate following map from swid
				swMap := map[string]string{
					"measurement":          SwMeasurement,
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

// InputSwidMatches checks whether a given SWID is present in a SWID list
func InputSwidMatches(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//GetSwIDTagsForQueryFromDB is the most generic function that is been used to fetch
// SWID's based on given query as an argument to the function.
//Please note returned swid list contains only naked nodes, it does not contain any links/relations among swid.
func (e *ArangoEndorsementStore) GetSwIDTagsForQueryFromDB(query string) ([]SoftwareIdentity, error) {
	ctx := context.Background()
	arangoConn := new(ArangoDBConnParams)
	e.ConnectToArangoDB(ctx, arangoConn)
	var swidList []SoftwareIdentity

	log.Printf("Supplied Query = %s", query)
	cursor, err := arangoConn.db.Query(ctx, query, nil)
	if err != nil {
		// handle error
		return nil, fmt.Errorf(" \nDB Query Failed with Reason = \n : %v", err)
	}
	defer cursor.Close()
	for {
		var Swdocwrap SoftwareIdentityWrapper
		var Swdoc SoftwareIdentity

		// Please note below API, ReadDocument, will return an Empty ' ' meta if the query does not yield any thing
		meta, err := cursor.ReadDocument(ctx, &Swdocwrap)

		// Please note that returned err could also be no more document cause.
		if driver.IsNoMoreDocuments(err) {
			log.Printf("No More Documents Left")
			break
		} else if err != nil {
			// handle other errors
			return nil, fmt.Errorf("Document Reading Failed with Reason = : %v", err)
		}

		if meta.Key == "" {
			log.Printf("Query did not returned any result for the links")
			return nil, fmt.Errorf("No results for the links detected")
		}
		Swdoc = Swdocwrap.ID
		swidList = append(swidList, Swdoc)
	}
	return swidList, nil
}

//GetAllSoftwareComponents implements the query function to fetch ALL software components associated
// to a platform.  What ALL means is that include software components which are linked to
// platform id and for each of these linked softwarecomponents, fetch associated patches and software
// updates presentin the DB.
func (e *ArangoEndorsementStore) GetAllSoftwareComponents(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var result []interface{}
	var finalList [][]SoftwareIdentity
	var schemeEndorsements []map[string]string

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	finalList, err = e.GetAllswidForPlatform(platformID)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract swid list for a platform Id %v", err)
	}

	for _, swidList := range finalList {
		// Give me Every SWID linked to this SWID in the DB
		for _, swid := range swidList {
			log.Printf("Fetched SWID = %s", swid.TagID)
			SwMeasurement := swid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
			swMap := map[string]string{
				"measurement":          SwMeasurement,
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

//GetAllswidForPlatform fetches all the SWID's for a given platformID
func (e *ArangoEndorsementStore) GetAllswidForPlatform(platformID string) ([][]SoftwareIdentity, error) {
	var HwID HardwareIdentity
	var finalList [][]SoftwareIdentity

	// Fetch the HW ID container from the DB associated to this platform
	err := e.GetHwIDFromArangoDB(platformID, &HwID)
	if err != nil {
		return nil, fmt.Errorf("Given platform Identity failed %v", err)
	}

	HwTag := HwID.TagID
	//Fetch the SWID List linked to this HW tag, based on Verif scheme
	swidList, err := e.GetSwIDTagsLinkedToHWTagFromDB(HwTag)
	if err != nil {
		return nil, fmt.Errorf("Query GetAllSoftwareComponents failed %v", err)
	}
	for _, swid := range swidList {
		log.Printf("Fetched SWID = %s", swid.TagID)
		var indexSwidList []SoftwareIdentity

		// first swid is always the one linked with platform
		indexSwidList = append(indexSwidList, swid)

		// Get every SWID linked to the base SWID in the DB
		relList, err := e.GetAllSwRelForBaseTag(swid.TagID)
		if err != nil {
			return nil, fmt.Errorf("Error occured while fetching relations %v", err)
		}
		for _, relswid := range relList {
			indexSwidList = append(indexSwidList, relswid)
		}
		finalList = append(finalList, indexSwidList)
	}
	return finalList, nil
}

//GetAllSwRelForBaseTag fetches all sw patches and updates for a given SWID
func (e *ArangoEndorsementStore) GetAllSwRelForBaseTag(SwIDTag string) (SwIDList []SoftwareIdentity, err error) {
	var swidList []SoftwareIdentity
	SwColTag := e.SwCollection + "/" + SwIDTag
	SwColTag = "'" + SwColTag + "'"

	query := "FOR swid IN " + mindepth + to + maxdepth + " ANY " + SwColTag + space
	query += e.RelCollection + newline + " RETURN swid"
	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	swidList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetAllSwRelForBaseTag failed %v", err)
	}
	return swidList, nil
}

//GetAllSwPatchesForSwIDTag fetches all sw patches for a given SWID
func (e *ArangoEndorsementStore) GetAllSwPatchesForSwIDTag(SwIDTag string) (SwIDList []SoftwareIdentity, err error) {
	var swidList []SoftwareIdentity
	SwColTag := e.SwCollection + "/" + SwIDTag
	SwColTag = "'" + SwColTag + "'"

	query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY " + SwColTag + space
	query += e.RelCollection + newline + "FILTER link.rel == 'patches' RETURN swid"
	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	swidList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetAllSwPatchesForSwIDTag failed %v", err)
	}
	return swidList, nil
}

//GetAllSwUpdatesForSwIDTag fetches all sw updates for a given SWID
func (e *ArangoEndorsementStore) GetAllSwUpdatesForSwIDTag(SwIDTag string) (SwIDList []SoftwareIdentity, err error) {
	var swidList []SoftwareIdentity
	SwColTag := e.SwCollection + "/" + SwIDTag
	SwColTag = "'" + SwColTag + "'"

	query := "FOR swid, link IN " + mindepth + to + maxdepth + " ANY " + SwColTag + space
	query += e.EdgeCollection + newline + "FILTER link.rel == 'updates' RETURN swid"
	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	swidList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetAllSwUpdatesForSwIDTag failed %v", err)
	}
	return swidList, nil
}

//GetPatchListForBaseSwVersion fetches only the list of forward patches for a specific base revision of SWID
func (e *ArangoEndorsementStore) GetPatchListForBaseSwVersion(SwTag string) (PatchList []SoftwareIdentity, err error) {
	SwColTag := e.SwCollection + "/" + SwTag
	SwColTag = "'" + SwColTag + "'"
	// SWID(C) --patches--> SWID(B) --patches--> SWID(A), returns B, C for A passed in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
	query += SwColTag + space + e.RelCollection + newline
	query += "PRUNE link.rel == 'updates'" + newline
	query += "FILTER link.rel == 'patches' RETURN swid"
	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	PatchList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetPatchListForBaseSwVersion failed %v", err)
	}
	return PatchList, nil
}

//GetBaseSwVersionFromPatch fetches only the Base SWID associated to this patch version
func (e *ArangoEndorsementStore) GetBaseSwVersionFromPatch(Patchswid SoftwareIdentity) (Baseswid SoftwareIdentity, err error) {
	var baseswid SoftwareIdentity
	var swidList []SoftwareIdentity

	SwColTag := e.SwCollection + "/" + Patchswid.TagID
	SwColTag = "'" + SwColTag + "'"
	// SWID(C) --patches--> SWID(B) --patches--> SWID(A), returns A when C or B supplied in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " OUTBOUND "
	query += SwColTag + space + e.RelCollection + newline
	query += "PRUNE link.rel == 'updates'" + newline
	query += "FILTER link.rel == 'patches' RETURN swid"
	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	swidList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return baseswid, fmt.Errorf("Query GetBaseSwVersionFromPatch failed %v", err)
	}
	if len(swidList) == 0 {
		// No base to the Input SWID, hence the patch itself is the Base
		return Patchswid, nil
	}
	baseswid = swidList[len(swidList)-1]
	return baseswid, nil
}

//GetUpdatesListForBaseSwVersion fetches only the list of forward updates for a specific base revision of SWID
func (e *ArangoEndorsementStore) GetUpdatesListForBaseSwVersion(Baseswid SoftwareIdentity) (UpdateList []SoftwareIdentity, err error) {
	SwColTag := e.SwCollection + "/" + Baseswid.TagID
	SwColTag = "'" + SwColTag + "'"
	// SWID(C) --updates--> SWID(B) --updates--> SWID(A), returns B, C for A passed in
	query := "FOR swid, link IN " + mindepth + to + maxdepth + " INBOUND "
	query += SwColTag + space + e.RelCollection + newline
	query += "PRUNE link.rel == 'patches'" + newline
	query += "FILTER link.rel == 'updates' RETURN swid"

	log.Printf("Constructed Query = %s", query)

	//Fetch the SWID List associated linked this tag
	UpdateList, err = e.GetSwIDTagsForQueryFromDB(query)
	if err != nil {
		return nil, fmt.Errorf("Query GetUpdatesListForBaseSwVersion failed %v", err)
	}
	if len(UpdateList) == 0 {
		// No update list to the Input SWID, hence the given update is most recent
		UpdateList = append(UpdateList, Baseswid)
	}
	return UpdateList, nil
}

//GetMostRecentSwVerOfLinkedComp for a given Platform ID and
// measurement for a SW Component, linked to HW Platform
// fetch most recent SW version to a System.
func (e *ArangoEndorsementStore) GetMostRecentSwVerOfLinkedComp(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var measurements []string
	var result []interface{}
	var schemeEndorsements []map[string]string
	var swidList []SoftwareIdentity
	var HwID HardwareIdentity

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	measurements, err = ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("Failed to extract software compoenents from query params %v", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	err = e.GetHwIDFromArangoDB(platformID, &HwID)
	if err != nil {
		return nil, fmt.Errorf("Given platform Identity failed %v", err)
	}

	HwTag := HwID.TagID
	//Fetch the SWID List linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(HwTag)
	if err != nil {
		return nil, fmt.Errorf("Query GetMostRecentSwVerOfLinkedComp failed %v", err)
	}
	// Only fetch the first measurement
	meas := measurements[0]
	requiredswid, matched := checkMeasMatchInSwIDList(swidList, meas)
	if !matched {
		return nil, fmt.Errorf("Query Failed Given platform Identity has no matching measurements %v", err)
	}
	// Now we have the required swid, use it to find the best SW version to return

	// Note by design it is assured that linked SW component (to platform) is always the base.
	latestSwid := requiredswid
	updateList, err := e.GetUpdatesListForBaseSwVersion(latestSwid)
	if err != nil {
		return nil, fmt.Errorf("Unable to fecth queries from ArangoDB %v", err)
	}
	if (len(updateList)) != 0 {
		latestSwid = updateList[len(updateList)-1]
	}
	// If Update List present new latestSwid becomes root of update
	// get the best patch on this list
	patchList, err := e.GetPatchListForBaseSwVersion(latestSwid.TagID)
	if err != nil {
		return nil, fmt.Errorf("Unable to fecth queries from ArangoDB %v", err)
	}
	if (len(patchList)) != 0 {
		latestSwid = patchList[len(patchList)-1]
	}
	SwMeasurement := latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue

	// We have the successful SWID structure here
	// Generate the map for the given swid
	swMap := map[string]string{
		"measurement":          SwMeasurement,
		"sw_component_type":    latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasType,
		"sw_component_version": latestSwid.SoftwareVersion,
		"signer_id":            latestSwid.Payload[FirstIndex].Resources[FirstIndex].SignerID,
	}
	schemeEndorsements = append(schemeEndorsements, swMap)
	result = append(result, schemeEndorsements)
	return result, nil
}

//checkMeasMatchInSwIDList is a utility function which checks for
//given measurement to be present in the SWID List
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

//GetMostRecentSwVerOfAnySWComp is the Master Blaster
// For a given Platform ID and measurement for a SW Component, which itself could be
// any patch or an update, fetch most recent SW version (recent patch on recent update branch)
func (e *ArangoEndorsementStore) GetMostRecentSwVerOfAnySWComp(args common.QueryArgs) (common.QueryResult, error) {
	var platformID string
	var measurements []string
	var result []interface{}
	var schemeEndorsements []map[string]string
	var swidList []SoftwareIdentity
	var HwID HardwareIdentity
	var matched, patchmatched bool

	platformID, err := ExtractPlatformID(args)
	if err != nil {
		// Handle Platform Id Error
		return nil, fmt.Errorf("Failed to extract platform_id from query params %v", err)
	}
	measurements, err = ExtractSoftwareComponents(args)
	if err != nil {
		// Handle Measurement Error
		return nil, fmt.Errorf("Failed to extract software compoenents from query params %v", err)
	}
	// If no measurements provided, we automatically "match" an empty set of components
	if len(measurements) == 0 {
		return common.QueryResult{[]interface{}{}}, nil
	}

	err = e.GetHwIDFromArangoDB(platformID, &HwID)
	if err != nil {
		return nil, fmt.Errorf("Given platform Identity failed %v", err)
	}

	HwTag := HwID.TagID
	//Fetch the SWID List linked to this tag
	swidList, err = e.GetSwIDTagsLinkedToHWTagFromDB(HwTag)
	if err != nil {
		return nil, fmt.Errorf("Query GetMostRecentSwVerOfLinkedComp failed %v", err)
	}
	// Only fetch the first measurement
	meas := measurements[0]
	requiredswid, matched := checkMeasMatchInSwIDList(swidList, meas)
	if !matched {
		/* ok, the given measurement is definitely not linked to Platform,
		check the Patches and Updates List for each swid, for a match? */
		for _, swid := range swidList {
			relList, err := e.GetAllSwPatchesForSwIDTag(swid.TagID)
			if err != nil {
				return nil, fmt.Errorf("Query GetMostRecentSwVerOfAnySWComp failed %v", err)
			}
			requiredswid, patchmatched = checkMeasMatchInSwIDList(relList, meas)
			matched = patchmatched
			if matched {
				break
			} else { /* Given measurement does not match patches, check the updates */
				relList, err := e.GetAllSwUpdatesForSwIDTag(swid.TagID)
				if err != nil {
					return nil, fmt.Errorf("Query GetMostRecentSwVerOfAnySWComp failed %v", err)
				}
				requiredswid, matched = checkMeasMatchInSwIDList(relList, meas)
			}
		}
	}
	if !matched {
		return nil, fmt.Errorf("Query failed for given platform Identity, supplied measurement not in DB %v", err)
	}
	// Now we have the required swid, use it to find the best SW version to return
	if patchmatched { // for a patch walk to its base
		requiredswid, err = e.GetBaseSwVersionFromPatch(requiredswid)
		if err != nil {
			return nil, fmt.Errorf("Query GetMostRecentSwVerOfAnySWComp failed %v", err)
		}
	}
	latestSwid := requiredswid
	// fetch the most recent update on this base
	updateList, err := e.GetUpdatesListForBaseSwVersion(latestSwid)
	if err != nil {
		return nil, fmt.Errorf("Unable to fecth queries from ArangoDB %v", err)
	}
	if (len(updateList)) != 0 {
		latestSwid = updateList[len(updateList)-1]
	}
	// If Update List present new latestSwid becomes root of update
	// get the best patch on this list
	patchList, err := e.GetPatchListForBaseSwVersion(latestSwid.TagID)
	if err != nil {
		return nil, fmt.Errorf("Unable to fecth queries from ArangoDB %v", err)
	}
	if (len(patchList)) != 0 {
		latestSwid = patchList[len(patchList)-1]
	}
	SwMeasurement := latestSwid.Payload[FirstIndex].Resources[FirstIndex].MeasurementValue
	swMap := map[string]string{
		"measurement":          SwMeasurement,
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
		"endorsementstore": &common.EndorsementStorePlugin{
			Impl: &ArangoEndorsementStore{},
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}
