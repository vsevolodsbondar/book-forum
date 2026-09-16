package helper

import (
	"flag"
)

func IsSeeding() bool {
	initiate := flag.Bool("init", false, "adds premade data for db. To use it: --init")
	flag.Parse()

	return *initiate
}
