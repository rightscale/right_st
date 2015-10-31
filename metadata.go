package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	Single InputType = iota
	Array
)

var (
	comment       = regexp.MustCompile(`^\s*(?:#|//|--)\s?(.*)$`)
	metadataStart = regexp.MustCompile(`^\s*(?:#|//|--)\s?(\s*-{3}\s*)$`)
	metadataEnd   = regexp.MustCompile(`^\s*(?:#|//|--)\s?(\s*\.{3}\s*)$`)
)

type RightScriptMetadata struct {
	Name        string                   `yaml:"RightScript Name"`
	Description string                   `yaml:"Description,omitempty"`
	Inputs      map[string]InputMetadata `yaml:"Inputs,omitempty"`
	Attachments []string                 `yaml:"Attachments,omitempty"`
}

type InputMetadata struct {
	Category       string       `yaml:"Category,omitempty"`
	Description    string       `yaml:"Description,omitempty"`
	InputType      InputType    `yaml:"Input Type,omitempty"`
	Required       bool         `yaml:"Required,omitempty"`
	Advanced       bool         `yaml:"Advanced,omitempty"`
	Default        *InputValue  `yaml:"Default,omitempty"`
	PossibleValues []InputValue `yaml:"Possible Values,omitempty"`
}

type InputType int

type InputValue struct {
	Type  string
	Value string
}

func ParseRightScriptMetadata(script *os.File) (*RightScriptMetadata, error) {
	defer script.Seek(0, os.SEEK_SET)
	scanner := bufio.NewScanner(script)
	var buffer bytes.Buffer
	inMetadata := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case inMetadata:
			submatches := metadataEnd.FindStringSubmatch(line)
			if submatches != nil {
				buffer.WriteString(submatches[1] + "\n")
				inMetadata = false
				break
			}
			submatches = comment.FindStringSubmatch(line)
			if submatches != nil {
				buffer.WriteString(submatches[1] + "\n")
			}
		case metadataStart.MatchString(line):
			submatches := metadataStart.FindStringSubmatch(line)
			buffer.WriteString(submatches[1] + "\n")
			inMetadata = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if inMetadata {
		return nil, fmt.Errorf("Unterminated RightScript metadata comment")
	}
	var metadata RightScriptMetadata
	// TODO: https://github.com/go-yaml/yaml/issues/136
	err := yaml.Unmarshal(buffer.Bytes(), &metadata)
	if err != nil {
		// TODO: adjust line numbers, etc.
		return &metadata, err
	}
	return &metadata, nil
}

func (i *InputType) MarshalYAML() (interface{}, error) {
	switch *i {
	case Single:
		return "single", nil
	case Array:
		return "array", nil
	default:
		return "", fmt.Errorf("Invalid input type value: %d", *i)
	}
}

func (i *InputType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	err := unmarshal(&value)
	if err != nil {
		return err
	}
	switch value {
	case "single":
		*i = Single
	case "array":
		*i = Array
	default:
		return fmt.Errorf("Invalid input type value: %s", value)
	}
	return nil
}

func (i *InputValue) MarshalYAML() (interface{}, error) {
	return i.Type + ":" + i.Value, nil
}

func (i *InputValue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var value string
	err := unmarshal(&value)
	if err != nil {
		return err
	}
	values := strings.SplitN(value, ":", 2)
	if len(values) < 2 {
		return fmt.Errorf("Invalid input value: %s", value)
	}
	*i = InputValue{Type: values[0], Value: values[1]}
	return nil
}
