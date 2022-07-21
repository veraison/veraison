// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package sessionmanager

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

const DefaultTTL = time.Minute

type SessionManagerTTLCache struct {
	cache *ttlcache.Cache[string, json.RawMessage]
}

func NewSessionManagerTTLCache() *SessionManagerTTLCache {
	return &SessionManagerTTLCache{}
}

func (o *SessionManagerTTLCache) Init(cfg Config) error {
	var (
		ttl time.Duration = DefaultTTL
		err error
	)

	if v, ok := cfg["ttl"]; ok {
		if ttl, err = time.ParseDuration(v); err != nil {
			return fmt.Errorf("invalid ttl: %q", v)
		}
	}

	o.cache = ttlcache.New[string, json.RawMessage](
		ttlcache.WithTTL[string, json.RawMessage](ttl),
	)

	go o.cache.Start()

	return nil
}

func (o *SessionManagerTTLCache) Close() error {
	o.cache.Stop()

	return nil
}

func (o *SessionManagerTTLCache) SetSession(id uuid.UUID, tenant string, session json.RawMessage, ttl time.Duration) error {
	_ = o.cache.Set(makeKey(id, tenant), session, ttl)

	return nil
}
func (o *SessionManagerTTLCache) DelSession(id uuid.UUID, tenant string) error {
	o.cache.Delete(makeKey(id, tenant))

	return nil
}

func (o *SessionManagerTTLCache) GetSession(id uuid.UUID, tenant string) (json.RawMessage, error) {
	if item := o.cache.Get(makeKey(id, tenant)); item != nil {
		return item.Value(), nil
	}

	return nil, fmt.Errorf("session not found for (id, tenant)=(%s, %s)", id, tenant)
}

func makeKey(id uuid.UUID, tenant string) string {
	// session://{tenant}/{uuid}
	u := url.URL{
		Scheme: "session",
		Host:   tenant,
		Path:   id.String(),
	}

	return u.String()
}
