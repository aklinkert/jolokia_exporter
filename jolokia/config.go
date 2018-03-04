package jolokia

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/ghodss/yaml"
	"strings"
	"sort"
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

	fixMbeanNames(config)

	return config, nil
}

// fixMeanNames sorts the request string of a mbean, e.g. from
// java.lang:type=GarbageCollector,name=* to java.lang:name=*,type=GarbageCollector
func fixMbeanNames(config *Config) {
	for index, m := range config.Metrics {
		parts := strings.Split(m.Source.Mbean, ":")
		if len(parts) == 1 {
			continue
		}

		fields := strings.Split(parts[1], ",")
		sort.Strings(fields)
		config.Metrics[index].Source.Mbean = strings.Join([]string{parts[0], strings.Join(fields, ",")}, ":")
	}
}
