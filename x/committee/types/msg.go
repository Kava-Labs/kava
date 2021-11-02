package types

import (
	"encoding/json"
	fmt "fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// VoteTypeFromString returns a VoteType from a string. It returns an error
// if the string is invalid.
func VoteTypeFromString(str string) (VoteType, error) {
	switch strings.ToLower(str) {
	case "yes", "y":
		return Yes, nil

	case "abstain", "a":
		return Abstain, nil

	case "no", "n":
		return No, nil

	default:
		return VoteType(0xff), fmt.Errorf("'%s' is not a valid vote type", str)
	}
}

// Marshal needed for protobuf compatibility.
func (vt VoteType) Marshal() ([]byte, error) {
	return []byte{byte(vt)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (vt *VoteType) Unmarshal(data []byte) error {
	*vt = VoteType(data[0])
	return nil
}

// Marshals to JSON using string.
func (vt VoteType) MarshalJSON() ([]byte, error) {
	return json.Marshal(vt.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (vt *VoteType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteTypeFromString(s)
	if err != nil {
		return err
	}

	*vt = bz2
	return nil
}

// Marshals to YAML using string.
func (vt VoteType) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(vt.String())
}

// UnmarshalJSON decodes from YAML assuming Bech32 encoding.
func (vt *VoteType) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteTypeFromString(s)
	if err != nil {
		return err
	}

	*vt = bz2
	return nil
}

// String implements the Stringer interface.
func (vt VoteType) String() string {
	switch vt {
	case Yes:
		return "Yes"
	case Abstain:
		return "Abstain"
	case No:
		return "No"
	default:
		return ""
	}
}

func (vt VoteType) Validate() error {
	if vt <= 0 || vt > 3 {
		return fmt.Errorf("invalid vote type: %d", vt)
	}
	return nil
}
