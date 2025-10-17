package seededuuid

import (
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

var namespace = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
var intSeed uint64 = 69420
var SeededRand = rand.New(rand.NewSource(intSeed))

func SeededUUID() uuid.UUID {
	oneTimeSeedStr := strconv.Itoa(SeededRand.Int())
	return uuid.NewMD5(namespace, []byte(oneTimeSeedStr))
}
