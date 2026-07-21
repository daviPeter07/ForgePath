package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	Editor   Editor                   `json:"editor"`
	Projects map[string]ProjectConfig `json:"projects"`
}

type Editor struct {
	Name       string `json:"name,omitempty"`
	Executable string `json:"executable,omitempty"`
}

type ProjectConfig struct {
	Command []string `json:"command,omitempty"`
}

func Default() Config {
	return Config{
		Projects: make(map[string]ProjectConfig),
	}
}

func DefaultPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "forgepath", "config.json"), nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return Config{}, fmt.Errorf("decode config %q: root must be a JSON object", path)
	}

	configuration := Default()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&configuration); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}
	if err := ensureEOF(decoder); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}
	if err := configuration.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}
	if configuration.Projects == nil {
		configuration.Projects = make(map[string]ProjectConfig)
	}
	return configuration, nil
}

func Init(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	removeIncomplete := true
	defer func() {
		_ = file.Close()
		if removeIncomplete {
			_ = os.Remove(path)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(Default()); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	removeIncomplete = false
	return nil
}

func (configuration Config) Validate() error {
	for name, project := range configuration.Projects {
		if name == "" {
			return fmt.Errorf("project name cannot be empty")
		}
		if len(project.Command) == 0 || project.Command[0] == "" {
			return fmt.Errorf("project %q command cannot be empty", name)
		}
	}
	return nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
