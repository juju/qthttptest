// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package qthttptest

import (
	"encoding/json"
	"errors"
	"fmt"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/mgo.v2/bson"
	yaml "gopkg.in/yaml.v2"
)

type codecEqualChecker struct {
	name      string
	marshal   func(interface{}) ([]byte, error)
	unmarshal func([]byte, interface{}) error
}

// BSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as BSON, is equal to the given value. Rather than
// unmarshaling into something of the expected body type, we reform
// the expected body in BSON and back to interface{} so we can check
// the whole content. Otherwise we lose information when unmarshaling.
var BSONEquals = &codecEqualChecker{
	marshal:   bson.Marshal,
	unmarshal: bson.Unmarshal,
}

// JSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as JSON, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in JSON and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var JSONEquals = &codecEqualChecker{
	marshal:   json.Marshal,
	unmarshal: json.Unmarshal,
}

// YAMLEquals defines a checker that checks whether a byte slice, when
// unmarshaled as YAML, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in YAML and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var YAMLEquals = &codecEqualChecker{
	marshal:   yaml.Marshal,
	unmarshal: yaml.Unmarshal,
}

func (checker *codecEqualChecker) ArgNames() []string {
	return []string{"got", "want"}
}

func (checker *codecEqualChecker) Check(got interface{}, args []interface{}, note func(key string, value interface{})) error {
	gotContent, ok := got.(string)
	if !ok {
		return qt.BadCheckf("expected string, got %T", got)
	}
	expectContent := args[0]
	expectContentBytes, err := checker.marshal(expectContent)
	if err != nil {
		return qt.BadCheckf("cannot marshal expected contents: %v", err)
	}
	var expectContentVal interface{}
	if err := checker.unmarshal(expectContentBytes, &expectContentVal); err != nil {
		return fmt.Errorf("cannot unmarshal expected contents: %v", err)
	}

	var gotContentVal interface{}
	if err := checker.unmarshal([]byte(gotContent), &gotContentVal); err != nil {
		return fmt.Errorf("cannot unmarshal obtained contents: %v; %q", err, gotContent)
	}
	if diff := cmp.Diff(gotContentVal, expectContentVal); diff != "" {
		note("diff (-got +want)", qt.Unquoted(diff))
		return errors.New("values are not equal")
	}
	return nil
}
