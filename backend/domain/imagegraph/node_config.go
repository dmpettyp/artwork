package imagegraph

import "fmt"

type NodeConfig map[string]any

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
		return 0.0, fmt.Errorf("float config key %q is not a float", key)
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

func (nc NodeConfig) GetIntOptional(key string) (*int, error) {
	v, ok := nc[key]
	if !ok {
		return nil, nil // Key doesn't exist - this is OK for optional
	}

	vInt, ok := v.(int)

	if !ok {
		return nil, fmt.Errorf("config key %q is not an int", key)
	}

	return &vInt, nil
}

func (nc NodeConfig) GetFloatOptional(key string) (*float64, error) {
	v, ok := nc[key]
	if !ok {
		return nil, nil // Key doesn't exist - this is OK for optional
	}

	vFloat, ok := v.(float64)

	if !ok {
		return nil, fmt.Errorf("config key %q is not a float", key)
	}

	return &vFloat, nil
}

func (nc NodeConfig) GetStringOptional(key string) (*string, error) {
	v, ok := nc[key]
	if !ok {
		return nil, nil // Key doesn't exist - this is OK for optional
	}

	vString, ok := v.(string)

	if !ok {
		return nil, fmt.Errorf("config key %q is not a string", key)
	}

	return &vString, nil
}

func (nc NodeConfig) Each(f func(string, any) error) error {
	for k, v := range nc {
		err := f(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
