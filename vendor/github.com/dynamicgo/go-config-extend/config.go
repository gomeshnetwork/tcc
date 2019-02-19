package extend

import (
	"encoding/json"

	"github.com/dynamicgo/go-config"
	"github.com/dynamicgo/go-config/source/memory"
)

// SubConfig get sub config
func SubConfig(conf config.Config, path ...string) (config.Config, error) {
	data := conf.Get(path...).Bytes()

	subConfig := config.NewConfig()

	if err := subConfig.Load(memory.NewSource(memory.WithData(data))); err != nil {
		return nil, err
	}

	return subConfig, nil
}

// SubConfigMap .
func SubConfigMap(conf config.Config, path ...string) (map[string]config.Config, error) {
	data := conf.Get(path...).Bytes()

	var dataSlice map[string]interface{}

	if err := json.Unmarshal(data, &dataSlice); err != nil {
		return nil, err
	}

	subConfigSlice := make(map[string]config.Config)

	for k, data := range dataSlice {

		buff, err := json.Marshal(data)

		if err != nil {
			return nil, err
		}

		subConfig := config.NewConfig()

		if err := subConfig.Load(memory.NewSource(memory.WithData(buff))); err != nil {
			return nil, err
		}

		subConfigSlice[k] = subConfig
	}

	return subConfigSlice, nil
}

// SubConfigSlice get subconfig slice
func SubConfigSlice(conf config.Config, path ...string) ([]config.Config, error) {

	data := conf.Get(path...).Bytes()

	var dataSlice []interface{}

	if err := json.Unmarshal(data, &dataSlice); err != nil {
		return nil, err
	}

	var subConfigSlice []config.Config

	for _, data := range dataSlice {

		buff, err := json.Marshal(data)

		if err != nil {
			return nil, err
		}

		subConfig := config.NewConfig()

		if err := subConfig.Load(memory.NewSource(memory.WithData(buff))); err != nil {
			return nil, err
		}

		subConfigSlice = append(subConfigSlice, subConfig)
	}

	return subConfigSlice, nil
}
