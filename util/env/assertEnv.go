package env

import (
	"os"
	"strconv"

	l "github.com/sprisa/west/util/log"
	"golang.org/x/exp/constraints"
)

func Assert(name string) string {
	value := os.Getenv(name)
	if value == "" {
		l.Log.Panic().Str("name", name).Msg("Expected ENV value")
	}
	return value
}

func Maybe(name string) *string {
	value := os.Getenv(name)
	if value == "" {
		return nil
	}
	return &value
}

func Bool(name string) bool {
	value := os.Getenv(name)
	return value == "true" || value == "1"
}

func WithDefault(name string, def string) string {
	value := os.Getenv(name)
	if value == "" {
		// TODO: Only run this in prod mode. It breaks Ent gen for some reason
		// logger.Log.Warnf("ENV value not supplied for %s. Using default %s", name, def)
		return def
	}
	return value
}

func AssertAndParse[T any](name string, parse func(val string) T) T {
	return parse(Assert(name))
}

func Parse[T any](name string, parse func(val string) T) T {
	value := os.Getenv(name)
	return parse(value)
}

func IntWithDefault[T constraints.Integer](name string, def T) T {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	t, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return T(t)
}

func Int[T constraints.Integer](name string) T {
	return IntWithDefault(name, T(0))
}
