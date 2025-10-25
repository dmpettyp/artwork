package imagegraph

type NodeConfig map[string]interface{}

func (nc NodeConfig) Exists(key string) bool {
	_, ok := nc[key]
	return ok
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
