// Copyright 2022 Alex Schneider. All rights reserved.

package cron

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

/* ==================================================================================================== */

type randomSource struct{}

// Seed implements the rand.Source interface.
func (s randomSource) Seed(seed int64) {}

// Int63 implements the rand.Source interface.
func (s randomSource) Int63() int64 {
	return int64(s.uint64() & ^uint64(1<<63))
}

func (s randomSource) uint64() (data uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &data)
	if err != nil {
		panic(err)
	}

	return data
}

/* ==================================================================================================== */

// random represents a source of random numbers.
var random *mrand.Rand

func init() {
	var src randomSource

	random = mrand.New(src)
}
