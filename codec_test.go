// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package qthttptest_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/juju/qthttptest"
)

type Inner struct {
	First  string
	Second int             `json:",omitempty" yaml:",omitempty"`
	Third  map[string]bool `json:",omitempty" yaml:",omitempty"`
}

type Outer struct {
	First  float64
	Second []*Inner `json:"Last,omitempty" yaml:"last,omitempty"`
}

func TestJSONEquals(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		descr       string
		obtained    string
		expected    *Outer
		expectError string
	}{{
		descr:    "very simple",
		obtained: `{"First": 47.11}`,
		expected: &Outer{
			First: 47.11,
		},
	}, {
		descr:    "nested",
		obtained: `{"First": 47.11, "Last": [{"First": "Hello", "Second": 42}]}`,
		expected: &Outer{
			First: 47.11,
			Second: []*Inner{
				{First: "Hello", Second: 42},
			},
		},
	}, {
		descr: "nested with newline",
		obtained: `{"First": 47.11, "Last": [{"First": "Hello", "Second": 42},
			{"First": "World", "Third": {"T": true, "F": false}}]}`,
		expected: &Outer{
			First: 47.11,
			Second: []*Inner{
				{First: "Hello", Second: 42},
				{First: "World", Third: map[string]bool{
					"F": false,
					"T": true,
				}},
			},
		},
	}, {
		descr:    "illegal field",
		obtained: `{"NotThere": 47.11}`,
		expected: &Outer{
			First: 47.11,
		},
		expectError: `values are not equal`,
	}, {
		descr:       "illegal optained content",
		obtained:    `{"NotThere": `,
		expectError: `cannot unmarshal obtained contents: unexpected end of JSON input; .*`,
	}}
	for _, test := range tests {
		c.Run(test.descr, func(c *qt.C) {
			err := qthttptest.JSONEquals.Check(test.obtained, []interface{}{test.expected}, nopNote)
			if test.expectError != "" {
				c.Assert(err, qt.ErrorMatches, test.expectError)
			} else {
				c.Assert(err, qt.Equals, nil)
			}
		})
	}

	// Test non-string input.
	err := qthttptest.JSONEquals.Check(true, []interface{}{true}, nopNote)
	c.Assert(err, qt.ErrorMatches, "bad check: expected string, got bool")
	c.Assert(err, qt.Satisfies, qt.IsBadCheck)
}

func nopNote(key string, value interface{}) {}

func TestYAMLEquals(t *testing.T) {
	c := qt.New(t)
	tests := []struct {
		descr       string
		obtained    string
		expected    *Outer
		expectError string
	}{{
		descr:    "very simple",
		obtained: `first: 47.11`,
		expected: &Outer{
			First: 47.11,
		},
	}, {
		descr:    "nested",
		obtained: `{first: 47.11, last: [{first: 'Hello', second: 42}]}`,
		expected: &Outer{
			First: 47.11,
			Second: []*Inner{
				{First: "Hello", Second: 42},
			},
		},
	}, {
		descr: "nested with newline",
		obtained: `{first: 47.11, last: [{first: 'Hello', second: 42},
			{first: 'World', third: {t: true, f: false}}]}`,
		expected: &Outer{
			First: 47.11,
			Second: []*Inner{
				{First: "Hello", Second: 42},
				{First: "World", Third: map[string]bool{
					"f": false,
					"t": true,
				}},
			},
		},
	}, {
		descr:    "illegal field",
		obtained: `{"NotThere": 47.11}`,
		expected: &Outer{
			First: 47.11,
		},
		expectError: `values are not equal`,
	}, {
		descr:       "illegal obtained content",
		obtained:    `{"NotThere": `,
		expectError: `cannot unmarshal obtained contents: yaml: line 1: .*`,
	}}
	for _, test := range tests {
		c.Run(test.descr, func(c *qt.C) {
			err := qthttptest.YAMLEquals.Check(test.obtained, []interface{}{test.expected}, nopNote)
			if test.expectError != "" {
				c.Assert(err, qt.ErrorMatches, test.expectError)
			} else {
				c.Assert(err, qt.Equals, nil)
			}
		})
	}

	// Test non-string input.
	err := qthttptest.YAMLEquals.Check(true, []interface{}{true}, nopNote)
	c.Assert(err, qt.ErrorMatches, "bad check: expected string, got bool")
	c.Assert(err, qt.Satisfies, qt.IsBadCheck)
}
