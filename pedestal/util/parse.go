package util

import (
	"strconv"
	"strings"
)

func ParseInfoKey(key string) (name string, version float32) {
	nv := strings.Split(key, "@")
	if len(nv) == 1 {
		return key, -1
	}
	if len(nv) != 2 {
		return "", 0
	}
	ff, _ := strconv.ParseFloat(nv[1], 10)
	return nv[0], float32(ff)
}
