package config

import "fmt"

type LocalAllowList struct {
	Interfaces map[string]bool `yaml:"interfaces,omitempty"`
	Cidrs      map[string]bool `yaml:"-"`
}

func (l *LocalAllowList) UnmarshalYAML(unmarshal func(any) error) error {
	var raw map[string]any
	if err := unmarshal(&raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		return nil
	}
	for key, value := range raw {
		if key == "interfaces" {
			interfaces, err := decodeBoolMap(value, "local_allow_list.interfaces")
			if err != nil {
				return err
			}
			l.Interfaces = interfaces
			continue
		}
		flag, ok := value.(bool)
		if !ok {
			return fmt.Errorf("local_allow_list: expected boolean for %q", key)
		}
		if l.Cidrs == nil {
			l.Cidrs = make(map[string]bool)
		}
		l.Cidrs[key] = flag
	}
	return nil
}

func (l LocalAllowList) MarshalYAML() (any, error) {
	if len(l.Interfaces) == 0 && len(l.Cidrs) == 0 {
		return map[string]any{}, nil
	}
	out := make(map[string]any, len(l.Cidrs)+1)
	if len(l.Interfaces) > 0 {
		out["interfaces"] = l.Interfaces
	}
	for key, value := range l.Cidrs {
		out[key] = value
	}
	return out, nil
}

func decodeBoolMap(value any, path string) (map[string]bool, error) {
	switch typed := value.(type) {
	case map[string]bool:
		return typed, nil
	case map[string]any:
		out := make(map[string]bool, len(typed))
		for key, val := range typed {
			flag, ok := val.(bool)
			if !ok {
				return nil, fmt.Errorf("%s: expected boolean for %q", path, key)
			}
			out[key] = flag
		}
		return out, nil
	case map[any]any:
		out := make(map[string]bool, len(typed))
		for key, val := range typed {
			keyStr, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("%s: expected string key", path)
			}
			flag, ok := val.(bool)
			if !ok {
				return nil, fmt.Errorf("%s: expected boolean for %q", path, keyStr)
			}
			out[keyStr] = flag
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s: expected map", path)
	}
}
