package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

const (
	defaultEP = "http://localhost:8529"
)

// ArangoDBparams are the DB parameters required for smooth operation of ArangoDB
type ArangoDBparams struct {
	ConEndPoint    string
	StoreName      string
	GraphName      string
	Login          string
	Password       string
	HwCollection   string
	SwCollection   string
	EdgeCollection string
	RelCollection  string
}

// ArangoDBConnVars holds the dynamic connection information for arangoDB
type ArangoDBConnVars struct {
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

// ArangoStore is an Arango Instance of Store
type ArangoStore struct {
	isInitialised bool
	dbparams      ArangoDBparams
	connvars      ArangoDBConnVars
}

// NewArangoStore instantiates a new Arango store
func NewArangoStore(dbparams ArangoDBparams) (*ArangoStore, error) {
	as := &ArangoStore{}
	if err := as.init(dbparams); err != nil {
		return nil, err
	}
	return as, nil
}

func checkEndPoint(ep string) error {
	u, err := url.Parse(ep)
	if err != nil {
		return fmt.Errorf("failed to parse URL %s: %w", ep, err)
	} else if !u.IsAbs() {
		return fmt.Errorf("supplied URL %s is not absolute", ep)
	}

	return nil
}

func (e *ArangoStore) init(dbparams ArangoDBparams) error {
	e.dbparams = dbparams
	if dbparams.ConEndPoint == "" {
		e.dbparams.ConEndPoint = defaultEP
	} else if err := checkEndPoint(dbparams.ConEndPoint); err != nil {
		return fmt.Errorf("init failed, no valid connection endpoint: %w", err)
	} else {
		e.dbparams.ConEndPoint = dbparams.ConEndPoint
	}
	e.isInitialised = true
	return nil
}

// Connect is responsible for making connection to the arangoDB
func (e *ArangoStore) Connect(ctx context.Context) error {
	var ok bool
	var err error

	// Create an HTTP Connection First to the Client
	e.connvars.conn, err = http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{e.dbparams.ConEndPoint},
	})
	if err != nil {
		return fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	// Create a client
	e.connvars.client, err = driver.NewClient(driver.ClientConfig{
		Connection:     e.connvars.conn,
		Authentication: driver.BasicAuthentication(e.dbparams.Login, e.dbparams.Password),
	})
	if err != nil {
		return fmt.Errorf("failed to create a client: %w", err)
	}

	// Check if Database with a given name exists or not
	ok, err = e.connvars.client.DatabaseExists(ctx, e.dbparams.StoreName)
	if err != nil {
		return fmt.Errorf("failure while checking whether DB exists: %w", err)
	}
	if !ok {
		return fmt.Errorf("endorsement database %s does not exist", e.dbparams.StoreName)
	}
	e.connvars.db, err = e.connvars.client.Database(ctx, e.dbparams.StoreName)
	if err != nil {
		return fmt.Errorf("failed to connect to the Endorsement database: %w", err)
	}

	// Check if the Graph exists in the DB
	ok, err = e.connvars.db.GraphExists(ctx, e.dbparams.GraphName)
	if err != nil {
		return fmt.Errorf("failure while checking whether graph exists: %w", err)
	}
	if !ok {
		return fmt.Errorf("endorsement graph %s does not exist", e.dbparams.GraphName)
	}
	// Connect to the Graph
	e.connvars.graph, err = e.connvars.db.Graph(ctx, e.dbparams.GraphName)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", e.dbparams.GraphName, err)
	}

	log.Printf("connection to %s DB and %s graph succeeded!", e.dbparams.StoreName, e.dbparams.GraphName)
	e.connvars.paramValid = true
	return nil
}

// IsInitialised returns whether store is initialized or not?
func (e *ArangoStore) IsInitialised() bool {
	return e.isInitialised
}

// GetQueryParam returns a suitable query parameter based on supplied input
func (e *ArangoStore) GetQueryParam(input string) (string, error) {
	switch input {
	case HW:
		return e.dbparams.HwCollection, nil
	case SW:
		return e.dbparams.SwCollection, nil
	case Edge:
		return e.dbparams.EdgeCollection, nil
	case Rel:
		return e.dbparams.RelCollection, nil
	default:
		return "", fmt.Errorf("invalid input to GetQueryParam")
	}
}

// Close closes the ArangoDB Store
func (e *ArangoStore) Close(ctx context.Context) error {
	if err := e.connvars.graph.Remove(ctx); err != nil {
		return fmt.Errorf("unable to remove the graph: %w", err)
	}
	if err := e.connvars.db.Remove(ctx); err != nil {
		return fmt.Errorf("failed to remove database: %w", err)
	}
	return nil
}

// RunQuery implements the interface to fetch all documents associated with a query
func (e *ArangoStore) RunQuery(ctx context.Context, query string, bindVars map[string]interface{}, document interface{}) ([]interface{}, error) {
	log.Printf("input query = %s", query)
	// Empty DocList
	docList := []interface{}{}

	cursor, err := e.connvars.db.Query(ctx, query, bindVars)
	if err != nil {
		cursor.Close()
		return nil, fmt.Errorf("query failed with reason %w", err)
	}
	defer cursor.Close()

	for {
		meta, rc := cursor.ReadDocument(ctx, document)
		log.Printf("returned from readDocument %v", rc)

		if driver.IsNoMoreDocuments(rc) {
			log.Printf("no more documents Left")
			break
		} else if rc != nil {
			err = rc
			break
		} else {
			switch t := document.(type) {
			case *HardwareIdentityWrapper:
				var lHwIDWrapper HardwareIdentityWrapper
				lHwIDWrapper.ID = t.ID
				lHwIDWrapper.ID.Entity = make([]PsaEntity, len(t.ID.Entity))
				copy(lHwIDWrapper.ID.Entity, t.ID.Entity)

				// Append the document to the list
				docList = append(docList, lHwIDWrapper)
			case *SoftwareIdentityWrapper:
				var lSwIDWrapper SoftwareIdentityWrapper
				lSwIDWrapper.ID = t.ID
				lSwIDWrapper.ID.Entity = make([]PsaEntity, len(t.ID.Entity))
				copy(lSwIDWrapper.ID.Entity, t.ID.Entity)
				lSwIDWrapper.ID.Payload = make([]ResourceCollection, len(t.ID.Payload))
				for index, resource := range t.ID.Payload {
					lSwIDWrapper.ID.Payload[index].Resources = make([]Resource, len(resource.Resources))
					copy(lSwIDWrapper.ID.Payload[index].Resources, resource.Resources)
				}

				// Append the document to the list
				docList = append(docList, lSwIDWrapper)
			default:
				return nil, fmt.Errorf("document fetch failed: unsupported doc type")
			}
		}
		log.Printf("got doc with key '%s' from query", meta.Key)
	}
	return docList, err
}
