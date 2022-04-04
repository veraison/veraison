// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

type IKVStore interface {
	Init(cfg Config) error
	Close() error
	Get(key string) ([]string, error)
	Set(key, val string) error
	Del(key string) error
	Add(key, val string) error
}
