package raid

import (
	"github.com/satori/go.uuid"
)

func Etag() string {
	return uuid.NewV4().String()
}