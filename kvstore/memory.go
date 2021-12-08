// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package kvstore

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"text/tabwriter"
)

var (
	lk = sync.RWMutex{}
)

type Memory struct {
	Type Type
	Data map[string]string
}

func (o *Memory) Init(cfg Config) error {
	if err := o.Type.SetFromConfig(cfg); err != nil {
		return err
	}

	o.Data = make(map[string]string)

	return nil
}

func (o *Memory) Term() error {
	// no-op (the map is garbage collected)
	return nil
}

func (o Memory) Get(key string) (string, error) {
	if o.Data == nil {
		return "", errors.New("memory store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return "", err
	}

	lk.RLock()
	defer lk.RUnlock()

	val, ok := o.Data[key]
	if !ok {
		return "", fmt.Errorf("key %q not found", key)
	}

	return val, nil
}

func (o *Memory) Set(key string, val string) error {
	if o.Data == nil {
		return errors.New("memory store uninitialized")
	}

	if err := sanitizeKV(key, val); err != nil {
		return err
	}

	lk.Lock()
	defer lk.Unlock()

	_, ok := o.Data[key]
	if ok {
		return fmt.Errorf("key %q already exist, delete it first", key)
	}

	o.Data[key] = val

	return nil
}

func (o *Memory) Del(key string) error {
	if o.Data == nil {
		return errors.New("memory store uninitialized")
	}

	if err := sanitizeK(key); err != nil {
		return err
	}

	lk.Lock()
	defer lk.Unlock()

	delete(o.Data, key)

	return nil
}

func (o Memory) dump() string {
	var b bytes.Buffer

	w := tabwriter.NewWriter(&b, 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "Key\tVal")
	fmt.Fprintln(w, "---\t---")

	lk.RLock()
	defer lk.RUnlock()

	// stabilize output order
	sortedKeys := make([]string, 0, len(o.Data))
	for k := range o.Data {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		fmt.Fprintf(w, "%s\t%s\n", k, o.Data[k])
	}

	w.Flush()

	return b.String()
}

func (o Memory) Dump() {
	fmt.Println(o.dump())
}
