package persistencepg

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/ytake/protoactor-go-persistence-pg/testdata"
)

func TestEnvelope(t *testing.T) {
	t.Run("newEnvelope", func(t *testing.T) {
		evt := &testdata.UserCreated{
			UserID:   ulid.Make().String(),
			UserName: "test",
			Email:    "",
		}
		_, err := newEnvelope(evt)
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}
