package customs

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Formatter is used to output inspection results. For example, you could have
// an stdout formatter for local development and a GitHub formatter to post
// results to a Pull Request.
type Formatter interface {
	Format(source string, i *Import, r Result) error
}

type Configuration struct {
	// ConcurrentInspections is the number of inspections to run concurrently.
	Concurrency int
	// Formatter is used to output the customs.Result
	Formatter  Formatter
	Inspectors map[string]string
}

type yamlConfiguration struct {
	Customs struct {
		Concurrency int    `yaml:"concurrency"`
		Formatter   string `yaml:"formatter"`
		Inspectors  map[string]struct {
			Command string `yaml:"command"`
		} `yaml:"inspectors"`
	} `yaml:"customs"`
}

// ParseConfig accepts a reader that should return YAML configuration for
// customs. It returns the parsed configuration.
func ParseConfig(r io.Reader, c *Configuration, formatters map[string]Formatter) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("could not read configuration file: %w", err)
	}

	var yamlConfig yamlConfiguration
	err = yaml.Unmarshal(content, &yamlConfig)
	if err != nil {
		return fmt.Errorf("could not parse configuration file: %w", err)
	}

	if yamlConfig.Customs.Concurrency > 0 {
		c.Concurrency = yamlConfig.Customs.Concurrency
	}

	if yamlConfig.Customs.Formatter != "" {
		formatter, ok := formatters[yamlConfig.Customs.Formatter]
		if !ok {
			return fmt.Errorf("could not find formatter '%s'", yamlConfig.Customs.Formatter)
		}
		c.Formatter = formatter
	}

	if c.Inspectors == nil {
		c.Inspectors = make(map[string]string, len(yamlConfig.Customs.Inspectors))
	}
	for name, inspector := range yamlConfig.Customs.Inspectors {
		c.Inspectors[name] = inspector.Command
	}

	return nil
}
