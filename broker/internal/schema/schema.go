// Package schema validates agent output against the versioned JSON
// schemas (Council doc 04 §8). Agent output is untrusted until it passes
// here; a second malformed response is an infrastructure error, never an
// approval (doc 04 §7.2).
package schema

import (
	"encoding/json"
	"fmt"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type re2 struct{ re *regexp2.Regexp }

func (r re2) MatchString(s string) bool { ok, err := r.re.MatchString(s); return err == nil && ok }
func (r re2) String() string            { return r.re.String() }

func engine(pattern string) (jsonschema.Regexp, error) {
	c, err := regexp2.Compile(pattern, regexp2.ECMAScript)
	if err != nil {
		return nil, err
	}
	c.MatchTimeout = 2_000_000_000
	return re2{c}, nil
}

// Validator holds one compiled schema.
type Validator struct{ s *jsonschema.Schema }

// New compiles a schema file into a validator.
func New(schemaPath string) (*Validator, error) {
	c := jsonschema.NewCompiler()
	c.UseRegexpEngine(engine)
	s, err := c.Compile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("compile %s: %w", schemaPath, err)
	}
	return &Validator{s}, nil
}

// ValidateBytes checks raw JSON against the schema.
func (v *Validator) ValidateBytes(raw []byte) error {
	var val any
	if err := json.Unmarshal(raw, &val); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return v.s.Validate(val)
}
