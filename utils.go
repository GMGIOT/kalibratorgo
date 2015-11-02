package main

import (
	"regexp"
	"strconv"
)

var unqote = regexp.MustCompile(`^"?([0-9A-Fa-fxX]*)"?$`)

func DecodeHex(hexNumper string) (int64, error) {
	if unqote.Match(([]byte)(hexNumper)) {
		hexNumper = unqote.FindStringSubmatch(hexNumper)[1]
	}

	return strconv.ParseInt(hexNumper, 0, 64);
}
