package ioutil

import "os"

func StdinAvailable() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode()&os.ModeCharDevice) == 0 || stat.Size() > 0
}
