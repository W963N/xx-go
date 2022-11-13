package main

import (
	"io"

	"github.com/naoina/toml"
)

type cmdConf struct {
	Root   string
	Output string
}

func loadConf(file io.Reader, config *cmdConf) error {
	return toml.NewDecoder(file).Decode(&config)
}
