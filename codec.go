// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package qthttptest

import (
	qt "github.com/frankban/quicktest"
	"gopkg.in/mgo.v2/bson"
	yaml "gopkg.in/yaml.v2"
)

// BSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as BSON, is equal to the given value. Rather than
// unmarshaling into something of the expected body type, we reform
// the expected body in BSON and back to interface{} so we can check
// the whole content. Otherwise we lose information when unmarshaling.
var BSONEquals = qt.CodecEquals(bson.Marshal, bson.Unmarshal)

// JSONEquals defines a checker that checks whether a byte slice, when
// unmarshaled as JSON, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in JSON and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
//
// Deprecated: use qt.JSONEquals instead.
var JSONEquals = qt.JSONEquals

// YAMLEquals defines a checker that checks whether a byte slice, when
// unmarshaled as YAML, is equal to the given value.
// Rather than unmarshaling into something of the expected
// body type, we reform the expected body in YAML and
// back to interface{}, so we can check the whole content.
// Otherwise we lose information when unmarshaling.
var YAMLEquals = qt.CodecEquals(yaml.Marshal, yaml.Unmarshal)
