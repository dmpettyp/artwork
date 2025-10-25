package imagegraph

import "fmt"

type NodeConfig map[string]interface{}

func (nc NodeConfig) Exists(key string) bool {
	_, ok := nc[key]
	return ok
}

func (nc NodeConfig) GetInt(key string) (int, error) {
	v, ok := nc[key]
	if !ok {
		return 0, fmt.Errorf("int config key %q does not exist", key)
	}

	vInt, ok := v.(int)

	if !ok {
		return 0, fmt.Errorf("int config key %q is not an int", key)
	}

	return vInt, nil
}

func (nc NodeConfig) GetFloat(key string) (float64, error) {
	v, ok := nc[key]
	if !ok {
		return 0.0, fmt.Errorf("float config key %q does not exist", key)
	}

	vFloat, ok := v.(float64)

	if !ok {
		return 0.0, fmt.Errorf("flaot config key %q is not a flaot", key)
	}

	return vFloat, nil
}

func (nc NodeConfig) GetString(key string) (string, error) {
	v, ok := nc[key]
	if !ok {
		return "", fmt.Errorf("string config key %q does not exist", key)
	}

	vString, ok := v.(string)

	if !ok {
		return "", fmt.Errorf("string config key %q is not a string", key)
	}

	return vString, nil
}

func (nc NodeConfig) Each(f func(string, interface{}) error) error {
	for k, v := range nc {
		err := f(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
