package jolokia

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/ghodss/yaml"
)

// LoadConfig reads a file and returns the contained config
func LoadConfig(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if ext := path.Ext(file); ext == ".yaml" || ext == ".yml" {
		b, err = yaml.YAMLToJSON(b)
		if err != nil {
			return nil, err
		}
	}

	config := &Config{}
	if err = json.Unmarshal(b, config); err != nil {
		return nil, err
	}

	return config, nil
}
