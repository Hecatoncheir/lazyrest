package environment

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	DefaultPublicFile  = "http-client.env.json"
	DefaultPrivateFile = "http-client.private.env.json"
	DefaultDotEnvFile  = ".env"
)

var variableNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)

type Config struct {
	Name        string
	PublicFile  string
	PrivateFile string
	DotEnvFile  string
}

type Environment struct {
	Name            string
	Values          map[string]string
	SecretVariables []string
}

func Load(rootDirectory string, config Config) (Environment, error) {
	result := Environment{Name: config.Name, Values: map[string]string{}}
	secretVariables := map[string]struct{}{}
	dotEnvFile := config.DotEnvFile
	if dotEnvFile == "" {
		dotEnvFile = DefaultDotEnvFile
	}
	dotEnvValues, _, err := loadDotEnv(resolvePath(rootDirectory, dotEnvFile))
	if err != nil {
		return Environment{}, err
	}
	for key, value := range dotEnvValues {
		result.Values[key] = value
		secretVariables[key] = struct{}{}
	}
	if config.Name == "" {
		result.SecretVariables = sortedVariables(secretVariables)
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
			delete(secretVariables, key)
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
			secretVariables[key] = struct{}{}
		}
	}
	if !found {
		return Environment{}, fmt.Errorf("environment %q was not found in %s or %s", config.Name, publicFile, privateFile)
	}

	result.SecretVariables = sortedVariables(secretVariables)
	return result, nil
}

func sortedVariables(variables map[string]struct{}) []string {
	result := make([]string, 0, len(variables))
	for name := range variables {
		result = append(result, name)
	}
	slices.Sort(result)
	return result
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
	defer func() { _ = file.Close() }()

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

func loadDotEnv(path string) (map[string]string, bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("open dotenv file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		separator := strings.IndexByte(line, '=')
		if separator < 1 {
			return nil, false, fmt.Errorf("parse dotenv file %s:%d: expected NAME=value", path, lineNumber)
		}
		name := strings.TrimSpace(line[:separator])
		value, err := parseDotEnvValue(strings.TrimSpace(line[separator+1:]))
		if err != nil {
			return nil, false, fmt.Errorf("parse dotenv file %s:%d: %w", path, lineNumber, err)
		}
		if err := setVariable(values, name, value); err != nil {
			return nil, false, fmt.Errorf("parse dotenv file %s:%d: %w", path, lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, false, fmt.Errorf("read dotenv file %s: %w", path, err)
	}
	return values, true, nil
}

func parseDotEnvValue(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	if value[0] == '\'' {
		closing := strings.IndexByte(value[1:], '\'')
		if closing < 0 {
			return "", fmt.Errorf("unterminated single-quoted value")
		}
		closing++
		if err := validateDotEnvSuffix(value[closing+1:]); err != nil {
			return "", err
		}
		return value[1:closing], nil
	}
	if value[0] == '"' {
		var result strings.Builder
		escaped := false
		for index := 1; index < len(value); index++ {
			character := value[index]
			if escaped {
				switch character {
				case 'n':
					result.WriteByte('\n')
				case 'r':
					result.WriteByte('\r')
				case 't':
					result.WriteByte('\t')
				case '\\', '"':
					result.WriteByte(character)
				default:
					result.WriteByte('\\')
					result.WriteByte(character)
				}
				escaped = false
				continue
			}
			if character == '\\' {
				escaped = true
				continue
			}
			if character == '"' {
				if err := validateDotEnvSuffix(value[index+1:]); err != nil {
					return "", err
				}
				return result.String(), nil
			}
			result.WriteByte(character)
		}
		return "", fmt.Errorf("unterminated double-quoted value")
	}

	for index, character := range value {
		if character == '#' && index > 0 && (value[index-1] == ' ' || value[index-1] == '\t') {
			value = value[:index]
			break
		}
	}
	return strings.TrimSpace(value), nil
}

func validateDotEnvSuffix(suffix string) error {
	suffix = strings.TrimSpace(suffix)
	if suffix == "" || strings.HasPrefix(suffix, "#") {
		return nil
	}
	return fmt.Errorf("unexpected text after quoted value")
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
