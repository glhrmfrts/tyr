package raid

import (
	"context"

	"github.com/satori/go.uuid"
)

type Etag string

func (e Etag) String() string {
	return string(e)
}

type etagKey int

var etagKeyValue etagKey = 0

func NewEtag() Etag {
	return Etag(uuid.NewV4().String())
}

func NewEtagContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, etagKeyValue, NewEtag())
}

func EtagFromContext(ctx context.Context) (Etag, bool) {
	etag, ok := ctx.Value(etagKeyValue).(Etag)
	return etag, ok
}
