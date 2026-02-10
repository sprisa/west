package config

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestLocalAllowListUnmarshal(t *testing.T) {
	data := []byte(`lighthouse:
  local_allow_list:
    interfaces:
      tun0: false
      'docker.*': false
    '10.0.0.0/8': true
`)
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	interfaces := map[string]bool{
		"tun0":     false,
		"docker.*": false,
	}
	if !reflect.DeepEqual(cfg.Lighthouse.LocalAllowList.Interfaces, interfaces) {
		t.Fatalf("interfaces mismatch: %#v", cfg.Lighthouse.LocalAllowList.Interfaces)
	}
	if !reflect.DeepEqual(cfg.Lighthouse.LocalAllowList.Cidrs, map[string]bool{"10.0.0.0/8": true}) {
		t.Fatalf("cidrs mismatch: %#v", cfg.Lighthouse.LocalAllowList.Cidrs)
	}
}

func TestLocalAllowListUnmarshalInterfaceError(t *testing.T) {
	data := []byte(`lighthouse:
  local_allow_list:
    interfaces:
      tun0: nope
`)
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err == nil {
		t.Fatalf("expected error for non-boolean interface entry")
	}
}

func TestLocalAllowListUnmarshalCIDRError(t *testing.T) {
	data := []byte(`lighthouse:
  local_allow_list:
    '10.0.0.0/8': nope
`)
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err == nil {
		t.Fatalf("expected error for non-boolean cidr entry")
	}
}

func TestLocalAllowListMarshal(t *testing.T) {
	cfg := Config{
		Lighthouse: Lighthouse{
			LocalAllowList: LocalAllowList{
				Interfaces: map[string]bool{
					"tun0":     false,
					"docker.*": false,
				},
				Cidrs: map[string]bool{
					"10.0.0.0/8": true,
				},
			},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded Config
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(decoded.Lighthouse.LocalAllowList.Interfaces, cfg.Lighthouse.LocalAllowList.Interfaces) {
		t.Fatalf("interfaces mismatch: %#v", decoded.Lighthouse.LocalAllowList.Interfaces)
	}
	if !reflect.DeepEqual(decoded.Lighthouse.LocalAllowList.Cidrs, cfg.Lighthouse.LocalAllowList.Cidrs) {
		t.Fatalf("cidrs mismatch: %#v", decoded.Lighthouse.LocalAllowList.Cidrs)
	}
}

func TestDecodeBoolMapAnyKeys(t *testing.T) {
	input := map[any]any{
		"tun0": true,
	}
	got, err := decodeBoolMap(input, "local_allow_list.interfaces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, map[string]bool{"tun0": true}) {
		t.Fatalf("unexpected map: %#v", got)
	}
}

func TestDecodeBoolMapAnyKeysInvalidKey(t *testing.T) {
	input := map[any]any{
		1: true,
	}
	_, err := decodeBoolMap(input, "local_allow_list.interfaces")
	if err == nil {
		t.Fatalf("expected error for non-string key")
	}
}

func TestDecodeBoolMapInvalidType(t *testing.T) {
	_, err := decodeBoolMap("nope", "local_allow_list.interfaces")
	if err == nil {
		t.Fatalf("expected error for non-map value")
	}
}

func TestDecodeBoolMapAnyKeysInvalidValue(t *testing.T) {
	input := map[any]any{
		"tun0": "nope",
	}
	_, err := decodeBoolMap(input, "local_allow_list.interfaces")
	if err == nil {
		t.Fatalf("expected error for non-boolean value")
	}
}

func TestDecodeBoolMapStringBoolPassthrough(t *testing.T) {
	input := map[string]bool{
		"tun0": true,
	}
	got, err := decodeBoolMap(input, "local_allow_list.interfaces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, input) {
		t.Fatalf("unexpected map: %#v", got)
	}
}

func TestLocalAllowListMarshalEmpty(t *testing.T) {
	list := LocalAllowList{}
	got, err := list.MarshalYAML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("unexpected value: %#v", got)
	}
}

func TestLocalAllowListUnmarshalError(t *testing.T) {
	list := LocalAllowList{}
	err := list.UnmarshalYAML(func(any) error {
		return fmt.Errorf("boom")
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
