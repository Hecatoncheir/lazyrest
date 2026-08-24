package environment

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
)

const (
	DefaultPublicFile  = "http-client.env.json"
	DefaultPrivateFile = "http-client.private.env.json"
)

var variableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)

type Config struct {
	Name        string
	PublicFile  string
	PrivateFile string
}

type Environment struct {
	Name            string
	Values          map[string]string
	SecretVariables []string
}

func Load(rootDirectory string, config Config) (Environment, error) {
	result := Environment{Name: config.Name, Values: map[string]string{}}
	if config.Name == "" {
		return result, nil
	}

	publicFile := config.PublicFile
	if publicFile == "" {
		publicFile = DefaultPublicFile
	}
	privateFile := config.PrivateFile
	if privateFile == "" {
		privateFile = DefaultPrivateFile
	}

	found := false
	publicValues, publicFound, err := loadProfile(resolvePath(rootDirectory, publicFile), config.Name)
	if err != nil {
		return Environment{}, err
	}
	if publicFound {
		found = true
		for key, value := range publicValues {
			result.Values[key] = value
		}
	}

	privateValues, privateFound, err := loadProfile(resolvePath(rootDirectory, privateFile), config.Name)
	if err != nil {
		return Environment{}, err
	}
	if privateFound {
		found = true
		for key, value := range privateValues {
			result.Values[key] = value
			result.SecretVariables = append(result.SecretVariables, key)
		}
	}
	if !found {
		return Environment{}, fmt.Errorf("environment %q was not found in %s or %s", config.Name, publicFile, privateFile)
	}

	slices.Sort(result.SecretVariables)
	return result, nil
}

func resolvePath(rootDirectory, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(rootDirectory, path)
}

func loadProfile(path, name string) (map[string]string, bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("open environment file %s: %w", path, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	var profiles map[string]any
	if err := decoder.Decode(&profiles); err != nil {
		return nil, false, fmt.Errorf("decode environment file %s: %w", path, err)
	}
	profile, ok := profiles[name]
	if !ok {
		return nil, false, nil
	}
	object, ok := profile.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("environment %q in %s must be a JSON object", name, path)
	}
	values := make(map[string]string)
	if err := flatten(values, "", object); err != nil {
		return nil, false, fmt.Errorf("environment %q in %s: %w", name, path, err)
	}
	return values, true, nil
}

func flatten(output map[string]string, prefix string, object map[string]any) error {
	for key, value := range object {
		name := key
		if prefix != "" {
			name = prefix + "." + key
		}
		switch typed := value.(type) {
		case map[string]any:
			if err := flatten(output, name, typed); err != nil {
				return err
			}
		case string:
			if err := setVariable(output, name, typed); err != nil {
				return err
			}
		case json.Number:
			if err := setVariable(output, name, typed.String()); err != nil {
				return err
			}
		case bool:
			if err := setVariable(output, name, strconv.FormatBool(typed)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("variable %q must be a string, number, boolean, or object", name)
		}
	}
	return nil
}

func setVariable(output map[string]string, name, value string) error {
	if !variableNamePattern.MatchString(name) {
		return fmt.Errorf("variable name %q is invalid", name)
	}
	if _, exists := output[name]; exists {
		return fmt.Errorf("variable %q is defined more than once", name)
	}
	output[name] = value
	return nil
}
