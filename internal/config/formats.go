package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	FormatYAML = "yaml"
	FormatJSON = "json"

	ExtensionYML  = ".yml"
	ExtensionYAML = ".yaml"
	ExtensionJSON = ".json"

	DefaultYAMLFilename            = "migration_config" + ExtensionYML
	DefaultYAMLAlternativeFilename = "migration_config" + ExtensionYAML
	DefaultJSONFilename            = "migration_config" + ExtensionJSON
)

func readConfigFile(path, workingDir string) ([]byte, string, error) {
	if path != "" {
		if workingDir != "" {
			path = filepath.Join(workingDir, path)
		}
		content, err := os.ReadFile(path)
		return content, path, err
	}
	for _, name := range []string{DefaultYAMLFilename, DefaultYAMLAlternativeFilename, DefaultJSONFilename} {
		candidate := filepath.Join(workingDir, name)
		content, err := os.ReadFile(candidate)
		if os.IsNotExist(err) {
			if _, statErr := os.Lstat(candidate); os.IsNotExist(statErr) {
				continue
			}
			return nil, candidate, errors.New("cannot read discovered configuration file")
		}
		return content, candidate, err
	}
	return nil, "", &os.PathError{
		Op:   "open",
		Path: DefaultYAMLFilename + ", " + DefaultYAMLAlternativeFilename + ", " + DefaultJSONFilename,
		Err:  os.ErrNotExist,
	}
}

func decodeConfig(path string, content []byte, cfg *FileConfig) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ExtensionYML, ExtensionYAML:
		return decodeYAML(content, cfg)
	default:
		if err := json.Unmarshal(content, cfg); err != nil {
			return formatJSONConfigError(path, content, err)
		}
		return nil
	}
}

// YAML errors can contain scalar values, including credentials. Report only a
// location extracted from the diagnostic, never its message or cause chain.
var yamlErrorLine = regexp.MustCompile(`line ([0-9]+)(?:[: ,]|$)`)

func invalidYAML(err error) error {
	if match := yamlErrorLine.FindStringSubmatch(err.Error()); len(match) == 2 {
		return fmt.Errorf("invalid YAML configuration at line %s", match[1])
	}
	return errors.New("invalid YAML configuration")
}

func decodeYAML(content []byte, cfg *FileConfig) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return invalidYAML(err)
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return errors.New("YAML configuration must be a mapping")
	}
	if err := document.Decode(cfg); err != nil {
		return invalidYAML(err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("YAML configuration must contain exactly one document")
	}
	return nil
}
