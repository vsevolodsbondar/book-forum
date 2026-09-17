package helper

import (
	"flag"
)

<<<<<<< HEAD
func IsSeeding() bool {
	initiate := flag.Bool("init", false, "adds premade data for db. To use it: --init")
	flag.Parse()

	return *initiate
=======
func FlagHandling() (bool, error) {
	initiate := flag.Bool("init", false, "adds premade data for db. To use it: --init")
	flag.Parse()

	return *initiate, nil
>>>>>>> 6e22c57 (now seeding with a flag.)
}
