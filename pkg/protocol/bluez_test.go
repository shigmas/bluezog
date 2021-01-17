package protocol

import (
	"context"
	"strings"
	"testing"

	"github.com/shigmas/bluezog/test"
	"github.com/stretchr/testify/assert"
)

func createBluez(t *testing.T, managedType string) (Bluez, func()) {
	//ops := bus.NewDbusOperations()
	ctx, cancel := context.WithCancel(context.Background())
	bluez, err := InitializeBluez(ctx, test.NewBusMock(managedType))
	assert.NoError(t, err, "Unexpected error initializing adapter")
	assert.NotNil(t, bluez, "Unable to initialize bluez")

	return bluez, cancel
}

// Since some tests need data from others, most tests are here. The ones that can be
// isolated are in other test files
func TestBluez(t *testing.T) {
	bluez, cancel := createBluez(t, "simple")
	defer cancel()

	t.Run("FindAdapter", func(t *testing.T) {
		adapters := bluez.FindAdapters()
		assert.GreaterOrEqual(t, len(adapters), 1)
	})

}

func TestBluezReceiveSignalStartStop(t *testing.T) {
	bluez, cancel := createBluez(t, "simple")
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
	bluez, cancel := createBluez(t, "simple")
	defer cancel()

	t.Run("FindAdapter", func(t *testing.T) {
		ch, err := bluez.AddWatch("/foo/bar", []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error AddWatch")
		assert.NotNil(t, ch, "Channel is nil")
		err = bluez.RemoveWatch("/foo/bar", ch, []InterfaceSignalPair{InterfaceSignalPair{"org.bluez.interface", "PropertiesChanged"}})
		assert.NoError(t, err, "Unexpected error RemoveWatch")
	})
}

func verifyPath(t *testing.T, obj Base, expectedElements int) {
	assert.Equal(t, expectedElements, len(strings.Split(string(obj.GetPath()), "/")),
		"Expected path to have %d elements", expectedElements)
}

func TestBluezGattObjects(t *testing.T) {
	bluez, cancel := createBluez(t, "gatt")
	defer cancel()

	services := bluez.GetObjectsByInterface(BluezInterface.GATTService)
	assert.NotEmpty(t, services, "Expected to find services")
	verifyPath(t, services[0], 6)
	characteristics := bluez.GetObjectsByInterface(BluezInterface.GATTCharacteristic)
	assert.NotEmpty(t, characteristics, "Expected to find characteristics")
	verifyPath(t, characteristics[0], 7)
	descriptors := bluez.GetObjectsByInterface(BluezInterface.GATTDescriptor)
	assert.NotEmpty(t, descriptors, "Expected to find descriptors")
	verifyPath(t, descriptors[0], 8)
}
