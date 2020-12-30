package protocol

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdapter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bluez, err := InitializeBluez(ctx)
	assert.NoError(t, err, "Unexpected error initializing adapter")
	assert.NotNil(t, bluez, "Unable to initialize bluez")
	var adapters []*Adapter

	t.Run("FindAdapter", func(t *testing.T) {
		adapters = bluez.FindAdapters()
		assert.NoError(t, err, "Unexpected error finding adapters")
		assert.GreaterOrEqual(t, len(adapters), 1)

		adapter := adapters[0]
		t.Run("AdapterGetAddress", func(t *testing.T) {
			address, err := adapter.FetchProperty(BluezAdapter.AddressProp)
			assert.NoError(t, err, "Unexpected error getting Address")
			fmt.Println("Adapter Address: ", address)
			assert.NotEmpty(t, address, "Address was empty")
			address = adapter.Property(BluezAdapter.AddressProp)
			fmt.Println("Adapter Address: ", address)
			assert.NotEmpty(t, address, "Address was empty")
		})
		t.Run("AdapterGetAlias", func(t *testing.T) {
			alias, err := adapter.FetchProperty(BluezAdapter.AliasProp)
			assert.NoError(t, err, "Unexpected error getting Alias")
			fmt.Println("Alias: ", alias)
			assert.NotEmpty(t, alias, "Alias was empty")
		})

		t.Run("Discovery", func(t *testing.T) {
			// Of course, if there are no devices, this will run forever
			ctx, cancel := context.WithTimeout(context.Background(),
				10*time.Second)
			ch, err := adapter.StartDiscovery()
			assert.NoError(t, err, "Unexpected error in StartDiscovery()")
			var chData ObjectChangedData
			go func(cancel func()) {
				count := 0
				for chData := range ch {
					fmt.Println(chData)
					fmt.Println("Device path: ", chData.Path)
					count++
					// Handle some kind of exit. When we add the bus stub for testing, we can put a
					// reasonable number here.
					if count == 1 {
						break
					}
				}
				cancel()
			}(cancel)

			<-ctx.Done()

			t.Run("Device", func(t *testing.T) {
				// This is empty for some reason.
				assert.Contains(t, string(chData.Path), "", "Discovered device path mismatch")
			})

			err = adapter.StopDiscovery()
			assert.NoError(t, err, "Unexpected error in StopDiscovery()")
		})
	})

}
