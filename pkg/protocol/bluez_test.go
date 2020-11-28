package protocol

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Since some tests need data from others, most tests are here. The ones that can be
// isolated are in other test files
func TestBluez(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bluez, err := InitializeBluez(ctx)
	assert.NoError(t, err, "Unexpected error initializing adapter")
	assert.NotNil(t, bluez, "Unable to initialize bluez")

	t.Run("FindAdapter", func(t *testing.T) {
		adapters := bluez.FindAdapters()
		assert.NoError(t, err, "Unexpected error finding adapters")
		assert.GreaterOrEqual(t, len(adapters), 1)
	})

}
