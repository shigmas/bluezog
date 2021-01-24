package zog

import (
	"context"
	//	"github.com/shigmas/bluezog/pkg/protocol"

	"testing"

	"github.com/shigmas/bluezog/pkg/bus"
	"github.com/shigmas/bluezog/test"
	"github.com/stretchr/testify/assert"
)

func TestBus(t *testing.T) {
	ctx := context.Background()
	ops := bus.DbusOperations{}

	bus := NewBus(ctx, &ops)
	err := bus.List("property", "Name")
	assert.NoError(t, err, "Unexpected error: ", err)
}

func TestGattPath(t *testing.T) {
	bus := NewBus(context.Background(), test.NewBusMock("gatt"))
	assert.NoError(t, bus.GetInterface(), "Unexpected error setting adapter")
	assert.NoError(t, bus.StartDiscovery(), "Unexpected error starting discovery")

	t.Run("TestBadPaths", func(t *testing.T) {
		devicePath := "/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4"
		descriptorAndMore := "/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f/char0022/desc0024/foo"

		assert.Error(t, bus.Gatt(devicePath), "Expected error on path")
		assert.Error(t, bus.Gatt(descriptorAndMore), "Expected error on path")
	})

	t.Run("TestGattPaths", func(t *testing.T) {
		t.Run("TestService", func(t *testing.T) {
			servicePath := "/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f"
			//assert.NoError(t, bus.Gatt(servicePath), "Unexpected error with service")
			assert.Error(t, bus.Gatt(servicePath), "Unexpected error with service")
		})
		t.Run("TestCharacteristic", func(t *testing.T) {
			characteristicPath := "/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f/char0022"
			//assert.NoError(t, bus.Gatt(characteristicPath), "Unexpected error with service")
			assert.Error(t, bus.Gatt(characteristicPath), "Unexpected error with service")
		})
		t.Run("TestDescriptor", func(t *testing.T) {
			descriptorPath := "/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f/char0022/desc0024"
			//assert.NoError(t, bus.Gatt(descriptorPath), "Unexpected error with service")
			assert.Error(t, bus.Gatt(descriptorPath), "Unexpected error with service")
		})
	})
}
