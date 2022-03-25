// Copyright 2021-2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func (c *ParamDescription) Apply(value interface{}) error {
	valueKind := reflect.ValueOf(value).Kind()
	kind := reflect.Kind(c.Kind)

	if valueKind != kind {
		switch t := value.(type) {
		case float64:
			if kind == reflect.Int || kind == reflect.Int32 || kind == reflect.Int64 ||
				kind == reflect.Int16 || kind == reflect.Int8 {
				if t != float64(int(t)) {
					return fmt.Errorf("expected type %q, but found %q", kind.String(), valueKind.String())
				}
			} else if kind == reflect.Uint || kind == reflect.Uint32 || kind == reflect.Uint64 ||
				kind == reflect.Uint16 || kind == reflect.Uint8 {
				if t != float64(int(t)) || t < 0 {
					return fmt.Errorf("expected type %q, but found %q", kind.String(), valueKind.String())
				}
			} else {
				return fmt.Errorf("expected type %q, but found %q", kind.String(), valueKind.String())
			}

		default:
			return fmt.Errorf("expected type %q, but found %q", kind.String(), valueKind.String())
		}
	}
	return nil
}

func NewParamStore(name string) *ParamStore {
	store := new(ParamStore)
	err := store.Init(name)
	if err != nil {
		return nil
	}
	return store
}

func (s *ParamStore) Init(name string) error {
	data, err := structpb.NewStruct(map[string]interface{}{})
	if err != nil {
		return err
	}

	s.Name = name
	s.Data = data
	s.Params = make(map[string]*ParamDescription)
	s.IsFrozen = false

	return nil
}

// DefineParam adds a definition for a parameter expected by the ParamStore.
// Values are retrieved from config sources and validated based on these
// definitions.
// name - the name of the parameter; used to retrieve it from the store
// kind - the type of the maters (defined by the reflect.Kind enum) used for validating values
// path - specifies where to look for the param's value inside config sources
// necessity - specifies whether the param is optional or required
func (s *ParamStore) DefineParam(name string, kind reflect.Kind, path string, necessity ParamNecessity) error {
	if s.IsFrozen {
		return fmt.Errorf("cannot define parameter %q -- store %q is frozen", name, s.Name)
	}

	s.Params[name] = &ParamDescription{
		Kind:     uint32(kind),
		Path:     path,
		Required: necessity,
	}
	if necessity == ParamNecessity_REQUIRED {
		s.Required = append(s.Required, name)
	}

	return nil
}

func (s *ParamStore) AddParamDefinitions(definitions map[string]*ParamDescription) error {
	for name, desc := range definitions {
		if err := s.DefineParam(name, reflect.Kind(desc.Kind), desc.Path, desc.Required); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates parameter values against the defined constraints, returning
// an error if some of the constraints have not been met. The constraints that are validated:
// - If a constraint exists for the parameter name, then the type of the value
//   that has been set matches the Kind specified in the constraint.
// - All mandatory parameters have been set.
// - If strict argument is set to true, no parameter values have been set for
//   which a constraint does not exist.
func (s *ParamStore) Validate(strict bool) error {
	data := s.Data.AsMap()
	for key, value := range data {
		constraint, ok := s.Params[key]
		if ok {
			if err := constraint.Apply(value); err != nil {
				return fmt.Errorf("constraint failed for %q: %s", key, err.Error())
			}
		} else if strict {
			return fmt.Errorf("unexpected parameter: %q", key)
		}

	}

	var missingRequired []string
	for _, req := range s.Required {
		if _, ok := data[req]; !ok {
			missingRequired = append(missingRequired, req)
		}
	}

	if len(missingRequired) > 0 {
		missingText := strings.Join(missingRequired, ", ")
		return fmt.Errorf("missing required parameter(s): %v", missingText)
	}

	return nil
}

// PopulateFromViper extracts values for the defined parameters from the
// specified Viper instance. If nil is specified, then the default Viper
// instance is used.
func (s *ParamStore) PopulateFromViper(v *viper.Viper) error {
	if v == nil {
		v = viper.GetViper()
	}

	for name, desc := range s.Params {
		if !v.IsSet(desc.Path) {
			continue
		}

		var value interface{}

		switch reflect.Kind(desc.Kind) {
		case reflect.Int:
			value = v.GetInt(desc.Path)
		case reflect.Int32:
			value = v.GetInt32(desc.Path)
		case reflect.Int64:
			value = v.GetInt64(desc.Path)
		case reflect.String:
			value = v.GetString(desc.Path)
		case reflect.Bool:
			value = v.GetBool(desc.Path)
		case reflect.Slice:
			slice := v.GetStringSlice(desc.Path)
			var genSlice []interface{}
			for _, elt := range slice {
				genSlice = append(genSlice, elt)
			}
			value = genSlice
		case reflect.Map:
			value = v.GetStringMap(desc.Path)
		default:
			value = v.Get(desc.Path)
		}

		var err error
		s.Data.Fields[name], err = structpb.NewValue(value)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: allow using paths for nested map.
func (s *ParamStore) PopulateFromMap(m map[string]interface{}) error {
	if len(s.Params) == 0 {
		return fmt.Errorf("no parameters have been defined")
	}

	for name, desc := range s.Params {
		value, ok := m[name]
		if !ok {
			continue
		}

		var err error
		var genVal interface{}

		switch reflect.Kind(desc.Kind) {
		case reflect.Int:
			genVal, err = toInt(value)
		case reflect.Int32:
			genVal, err = toInt32(value)
		case reflect.Int64:
			genVal, err = toInt64(value)
		case reflect.String:
			genVal, err = toString(value)
		case reflect.Bool:
			genVal, err = toBool(value)
		case reflect.Slice:
			genVal, err = toSlice(value)
		case reflect.Map:
			genVal, err = toStringMap(value)
		default:
			genVal, err = value, nil
		}

		if err != nil {
			return err
		}

		s.Data.Fields[name], err = structpb.NewValue(genVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ParamStore) PopulateFromStringMapString(sm map[string]string) error {
	m := make(map[string]interface{}, len(sm))
	for k, v := range sm {
		m[k] = v
	}

	return s.PopulateFromMap(m)
}

// Freeze prevents additional parameter defintions from being added (thus
// "freezing" the definition of this store). It does *not* prevent parameter
// values from being set.
func (s *ParamStore) Freeze() {
	s.IsFrozen = true
}

// Clear currently set parameter values (but not constraint definitions).
func (s *ParamStore) Clear() error {
	var err error
	s.Data, err = structpb.NewStruct(map[string]interface{}{})
	return err
}

func (s *ParamStore) GetParamNames() []string {
	var names []string

	for name := range s.Params {
		names = append(names, name)
	}

	return names
}

func (s *ParamStore) GetBool(key string) bool {
	v, _ := s.TryGetBool(key)
	return v
}

func (s *ParamStore) GetInt(key string) int {
	v, _ := s.TryGetInt(key)
	return v
}

func (s *ParamStore) GetString(key string) string {
	v, _ := s.TryGetString(key)
	return v
}

func (s *ParamStore) GetStringSlice(key string) []string {
	v, _ := s.TryGetStringSlice(key)
	return v
}

func (s *ParamStore) GetStringMapString(key string) map[string]string {
	v, _ := s.TryGetStringMapString(key)
	return v
}

func (s *ParamStore) SetBool(key string, value bool) error {
	return s.set(key, value)
}

func (s *ParamStore) SetInt(key string, value int) error {
	return s.set(key, value)
}

func (s *ParamStore) SetString(key string, value string) error {
	return s.set(key, value)
}

func (s *ParamStore) SetStringSlice(key string, value []string) error {
	var genVal []interface{}
	for _, v := range value {
		genVal = append(genVal, v)
	}

	return s.set(key, genVal)
}

func (s *ParamStore) SetStringMapString(key string, value map[string]string) error {
	genVal := make(map[string]interface{})
	for k, v := range value {
		genVal[k] = v
	}

	return s.set(key, genVal)
}

func (s *ParamStore) TryGetBool(key string) (bool, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return false, fmt.Errorf("key not found: %q", key)
	}

	return toBool(v)
}

func toBool(v interface{}) (bool, error) {
	switch t := v.(type) {
	case bool:
		return t, nil
	case string:
		return strconv.ParseBool(t)
	default:
		return false, fmt.Errorf("not a boolean value: %v (%T)", v, v)
	}
}

func (s *ParamStore) TryGetInt(key string) (int, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return 0, fmt.Errorf("key not found: %q", key)
	}

	return toInt(v)
}

func toInt(v interface{}) (int, error) {
	switch t := v.(type) {
	case int:
		return t, nil
	case int8:
		return int(t), nil
	case int16:
		return int(t), nil
	case int32:
		return int(t), nil
	case int64:
		return int(t), nil
	case float32:
		if t == float32(int(t)) {
			return int(t), nil
		}
		return 0, fmt.Errorf("float32 value %v cannot be represented as an int", t)
	case float64:
		if t == float64(int(t)) {
			return int(t), nil
		}
		return 0, fmt.Errorf("float64 value %v cannot be represented as an int", t)
	default:
		return 0, fmt.Errorf("not an int: %v", v)
	}
}

func toInt32(v interface{}) (int32, error) {
	switch t := v.(type) {
	case int:
		if t <= math.MaxInt32 && t >= math.MinInt32 {
			return int32(t), nil
		}
		return 0, fmt.Errorf("value outside int32 range: %d", t)
	case int8:
		return int32(t), nil
	case int16:
		return int32(t), nil
	case int32:
		return t, nil
	case int64:
		if t <= math.MaxInt32 && t >= math.MinInt32 {
			return int32(t), nil
		}
		return 0, fmt.Errorf("value outside int32 range: %d", t)
	default:
		return 0, fmt.Errorf("not an int32: %v", v)
	}
}

func toInt64(v interface{}) (int64, error) {
	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	default:
		return 0, fmt.Errorf("not an int64: %v", v)
	}
}

func (s *ParamStore) TryGetString(key string) (string, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return "", fmt.Errorf("key not found: %q", key)
	}

	res, err := toString(v)
	if err != nil {
		err = fmt.Errorf("cannot get value for %q: %s", key, err)
	}

	return res, err
}

func toString(v interface{}) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case fmt.Stringer:
		return t.String(), nil
	default:
		return "", fmt.Errorf("cannot convert to string: %v (%T)", v, v)
	}

}

func (s *ParamStore) TryGetStringSlice(key string) ([]string, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %q", key)
	}

	res, err := toStringSlice(v)
	if err != nil {
		err = fmt.Errorf("cannot get value for %q: %s", key, err)
	}

	return res, err
}

func toStringSlice(v interface{}) ([]string, error) {
	switch t := v.(type) {
	case []string:
		return t, nil
	case [][]byte:
		var res []string
		for _, bs := range t {
			res = append(res, string(bs))
		}
		return res, nil
	case []interface{}:
		var res []string
		for _, bs := range t {
			res = append(res, fmt.Sprint(bs))
		}
		return res, nil
	default:
		return nil, fmt.Errorf("cannot convert to []string: %v (%T)", v, v)
	}
}

func toSlice(v interface{}) ([]interface{}, error) {
	switch t := v.(type) {
	case []string:
		var res []interface{}
		for _, bs := range t {
			res = append(res, fmt.Sprint(bs))
		}
		return res, nil
	case [][]byte:
		var res []interface{}
		for _, bs := range t {
			res = append(res, bs)
		}
		return res, nil
	case []interface{}:
		return t, nil
	default:
		return nil, fmt.Errorf("cannot convert to []string: %v (%T)", v, v)
	}
}

func (s *ParamStore) TryGetStringMap(key string) (map[string]interface{}, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %q", key)
	}

	res, err := toStringMap(v)
	if err != nil {
		err = fmt.Errorf("cannot get value for %q: %s", key, err)
	}

	return res, err
}

func toStringMap(i interface{}) (map[string]interface{}, error) {
	switch t := i.(type) {
	case map[string]string:
		out := make(map[string]interface{})
		for k, v := range t {
			out[k] = fmt.Sprintf("%v", v)
		}
		return out, nil
	case map[string]interface{}:
		return t, nil
	default:
		return nil, fmt.Errorf("cannot convert to map[string]interface{}: %v (%T)", i, i)
	}
}

func (s *ParamStore) TryGetStringMapString(key string) (map[string]string, error) {
	v, ok := s.Data.AsMap()[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %q", key)
	}

	res, err := toStringMapString(v)
	if err != nil {
		err = fmt.Errorf("cannot get value for %q: %s", key, err)
	}

	return res, err
}

func toStringMapString(i interface{}) (map[string]string, error) {
	switch t := i.(type) {
	case map[string]string:
		return t, nil
	case map[string]interface{}:
		out := make(map[string]string)
		for k, v := range t {
			out[k] = fmt.Sprintf("%v", v)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("cannot convert to map[string]string: %v (%T)", i, i)
	}
}

func (s *ParamStore) set(key string, value interface{}) error {
	if _, ok := s.Data.Fields[key]; ok {
		return fmt.Errorf("duplicate key: %q", key)
	}

	var err error
	s.Data.Fields[key], err = structpb.NewValue(value)
	return err
}
