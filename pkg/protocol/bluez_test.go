package protocol

import (
	"context"
	"testing"

	"github.com/shigmas/bluezog/test"
	"github.com/stretchr/testify/assert"
)

func createBluez(t *testing.T) (Bluez, func()) {
	//ops := bus.NewDbusOperations()
	ctx, cancel := context.WithCancel(context.Background())
	bluez, err := InitializeBluez(ctx, &test.BusMock{})
	assert.NoError(t, err, "Unexpected error initializing adapter")
	assert.NotNil(t, bluez, "Unable to initialize bluez")

	return bluez, cancel
}

// Since some tests need data from others, most tests are here. The ones that can be
// isolated are in other test files
func TestBluez(t *testing.T) {
	bluez, cancel := createBluez(t)
	defer cancel()

	t.Run("FindAdapter", func(t *testing.T) {
		adapters := bluez.FindAdapters()
		assert.GreaterOrEqual(t, len(adapters), 1)
	})

}

func TestBluezReceiveSignalStartStop(t *testing.T) {
	bluez, cancel := createBluez(t)
	defer cancel()

	t.Run("FindAdapter", func(t *testing.T) {
		ch, err := bluez.AddWatch("/foo/bar", []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error AddWatch")
		assert.NotNil(t, ch, "Channel is nil")
		err = bluez.RemoveWatch("/foo/bar", ch, []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error RemoveWatch")
	})
}

func TestBluezStartStop(t *testing.T) {
	bluez, cancel := createBluez(t)
	defer cancel()

	t.Run("FindAdapter", func(t *testing.T) {
		ch, err := bluez.AddWatch("/foo/bar", []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error AddWatch")
		assert.NotNil(t, ch, "Channel is nil")
		err = bluez.RemoveWatch("/foo/bar", ch, []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error RemoveWatch")
	})
}
