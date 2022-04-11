// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"context"
	"fmt"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

const (
	defaultEP        = "http://localhost:8529"
	defaultStoreName = "ArangoEndStore"
	defaultCollName  = "ArangoEndorsements"
)

// ArangoDBparams are the DB parameters required for smooth operation of ArangoDB
type ArangoDBparams struct {
	ConEndPoint    string
	StoreName      string
	CollectionName string
	Login          string
	Password       string
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
}

// ArangoStore is an Arango Instance of Store
type ArangoStore struct {
	isInitialised bool
	dbparams      ArangoDBparams
	connvars      ArangoDBConnVars
}

func (o *ArangoStore) Init(cfg Config) error {

	ConnEP, err := cfg.ReadVarString(ConEndPoint)
	if err != nil {
		switch err {
		case ErrMissingDirective:
			o.dbparams.ConEndPoint = defaultEP
		default:
			return fmt.Errorf("%w: %q", err, ConEndPoint)
		}
	} else {
		o.dbparams.ConEndPoint = ConnEP
	}
	StName, err := cfg.ReadVarString(StoreName)
	if err != nil {
		switch err {
		case ErrMissingDirective:
			o.dbparams.StoreName = defaultStoreName
		default:
			return fmt.Errorf("%w: %q", err, StoreName)
		}
	} else {
		o.dbparams.StoreName = StName
	}

	ColName, err := cfg.ReadVarString(CollName)
	if err != nil {
		switch err {
		case ErrMissingDirective:
			o.dbparams.CollectionName = defaultCollName
		default:
			return fmt.Errorf("%w: %q", err, ColName)
		}
	} else {
		o.dbparams.CollectionName = ColName
	}

	Login, err := cfg.ReadVarString(Login)
	if err != nil {
		return fmt.Errorf("%w: %q", err, Login)
	}
	o.dbparams.Login = Login

	Password, err := cfg.ReadVarString(Password)
	if err != nil {
		return fmt.Errorf("%w: %q", err, Password)
	}
	o.dbparams.Password = Password

	ctx := context.Background()
	if err = o.connect(ctx); err != nil {
		return fmt.Errorf("initialisation failed unable to connect to DB %v", err)
	}
	o.isInitialised = true
	return nil
}

func (o *ArangoStore) connect(ctx context.Context) error {
	var ok bool
	var err error

	// Create an HTTP Connection First to the Client
	o.connvars.conn, err = http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{o.dbparams.ConEndPoint},
	})
	if err != nil {
		return fmt.Errorf("failed to create HTTP connection: %w", err)
	}

	// Create a client
	o.connvars.client, err = driver.NewClient(driver.ClientConfig{
		Connection:     o.connvars.conn,
		Authentication: driver.BasicAuthentication(o.dbparams.Login, o.dbparams.Password),
	})
	if err != nil {
		return fmt.Errorf("failed to create a client: %w", err)
	}

	// Now that we have created the client create the collection only once

	// Check if Database with a given name exists or not
	ok, err = o.connvars.client.DatabaseExists(ctx, o.dbparams.StoreName)
	if err != nil {
		return fmt.Errorf("failure while checking whether DB exists: %w", err)
	}
	if !ok && o.isInitialised {
		return fmt.Errorf("endorsement database %s does not exist", o.dbparams.StoreName)
	}

	if !o.isInitialised {
		// Create the data base now, should happen only once during system initialisation
		o.connvars.db, err = o.connvars.client.CreateDatabase(nil, o.dbparams.StoreName, nil)
		if err != nil {
			return fmt.Errorf("failed to create database :%v", err)
		}

		// Create the collection
		_, err := o.connvars.db.CreateCollection(nil, o.dbparams.CollectionName, nil)
		if err != nil {
			fmt.Errorf("failed to create collection")
		}
	} else {
		o.connvars.db, err = o.connvars.client.Database(ctx, o.dbparams.StoreName)
		if err != nil {
			return fmt.Errorf("failed to connect to the Endorsement database: %w", err)
		}
		coll_exists, err := o.connvars.db.CollectionExists(nil, o.dbparams.CollectionName, nil)

		if err != nil {
			return fmt.Errorf("failed to check the collection")
		}
		if !coll_exists {
			return fmt.Errorf("unable to locate the collection: %v", o.dbparams.CollectionName)
		}
	}
	return nil
}

// Close shuts down the store
func (o *ArangoStore) Close() error {
	ctx := context.Background()
	if err := o.connect(ctx); err != nil {
		return err
	}

	// Check do we have to remove any collections before removing the actual DB ?
	if err := o.connvars.db.Remove(ctx); err != nil {
		return fmt.Errorf("failed to remove database: %w", err)
	}
	o.isInitialised = false
	return nil
}
