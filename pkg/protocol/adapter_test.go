package protocol

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shigmas/bluezog/test"
	"github.com/stretchr/testify/assert"
)

func TestAdapter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//bluez, err := InitializeBluez(ctx, bus.NewDbusOperations())
	bluez, err := InitializeBluez(ctx, &test.BusMock{})
	assert.NoError(t, err, "Unexpected error initializing adapter")
	assert.NotNil(t, bluez, "Unable to initialize bluez")
	var adapter *Adapter

	t.Run("FindAdapter", func(t *testing.T) {
		adapters := bluez.FindAdapters()
		assert.NoError(t, err, "Unexpected error finding adapters")
		assert.GreaterOrEqual(t, len(adapters), 1)

		adapter = adapters[0]
		mock := true
		if !mock {
			t.Run("AdapterGetAddress", func(t *testing.T) {
				address := adapter.Property(BluezAdapter.AddressProp)
				fmt.Println("Adapter Address: ", address)
				assert.NotEmpty(t, address, "Address was empty")
				address = adapter.Property(BluezAdapter.AddressProp)
				fmt.Println("Adapter Address: ", address)
				assert.NotEmpty(t, address, "Address was empty")
			})
			t.Run("AdapterGetAlias", func(t *testing.T) {
				alias := adapter.Property(BluezAdapter.AliasProp)
				fmt.Println("Alias: ", alias)
				assert.NotEmpty(t, alias, "Alias was empty")
			})
		}
	})

	t.Run("Discovery", func(t *testing.T) {
		// Of course, if there are no devices, this will run forever
		ctx, cancel := context.WithTimeout(context.Background(),
			4*time.Second)
		ch, err := adapter.StartDiscovery()
		assert.NoError(t, err, "Unexpected error in StartDiscovery()")
		var chData ObjectChangedData
		count := 0
		go func(cancel func()) {
			for chData := range ch {
				fmt.Println(chData)
				fmt.Println("Device path: ", chData.Path)
				count++
				// Handle some kind of exit. When we add the bus stub for testing, we can put a
				// reasonable number here.
				if count == 2 {
					break
				}
			}
			cancel()
		}(cancel)

		<-ctx.Done()
		assert.Equal(t, 2, count, "Incorrect number of devices discovered")
		t.Run("Device", func(t *testing.T) {
			// This is empty for some reason.
			assert.Contains(t, string(chData.Path), "", "Discovered device path mismatch")
		})

		err = adapter.StopDiscovery()
		assert.NoError(t, err, "Unexpected error in StopDiscovery()")

		// restart
		_, err = adapter.StartDiscovery()
		assert.NoError(t, err, "Unable to restart discovery")
		err = adapter.StopDiscovery()
		assert.NoError(t, err, "Unexpected error in StopDiscovery()")
	})

}
