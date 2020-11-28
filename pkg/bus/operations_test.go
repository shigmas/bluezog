package bus

import (
	"context"
	"fmt"
	"testing"

	"github.com/godbus/dbus/v5"

	"github.com/stretchr/testify/assert"
)

func TestObject(t *testing.T) {
	conn, err := dbus.SystemBus()
	assert.NoError(t, err, "Unable to connect to system d-bus")

	badDest := "org.noservice"
	noPath := dbus.ObjectPath("/foo/bar")
	objDest := "org.bluez"
	objPath := dbus.ObjectPath("/org/bluez")

	t.Run("GetObject", func(t *testing.T) {
		t.Run("Failure", func(t *testing.T) {
			node, err := GetObject(conn, badDest, noPath)
			assert.Error(t, err, "Expected error for service %s and path %s: err: %s",
				badDest, noPath, err)
			assert.Nil(t, node)
		})
		t.Run("Success", func(t *testing.T) {
			node, err := GetObject(conn, objDest, objPath)
			assert.NoError(t, err, "Error for service %s and path %s: err: %s",
				objDest, objPath, err)
			assert.NotNil(t, node)
		})
	})

	t.Run("CallFunctions", func(t *testing.T) {
		adapterPath := dbus.ObjectPath("/org/bluez/hci0")
		badFunc := "org.bluez.Adapter1.NoSuchFunction"
		startFunc := "org.bluez.Adapter1.StartDiscovery"
		stopFunc := "org.bluez.Adapter1.StopDiscovery"
		// Should really see what functions should be called for different numbers of
		// arguments.
		t.Run("Failure", func(t *testing.T) {
			err := CallFunction(conn, objDest, adapterPath, badFunc)
			assert.Error(t, err, "Expected Error for %s and function %s",
				adapterPath, badFunc)

			// Should still fail because we haven't started discovery
			err = CallFunction(conn, objDest, adapterPath, stopFunc)
			assert.Error(t, err, "Expected Error for %s and function %s",
				adapterPath, badFunc)
		})

		t.Run("Success", func(t *testing.T) {
			err = CallFunction(conn, objDest, adapterPath, startFunc)
			assert.NoError(t, err, "Unexpected Error for %s", startFunc)
			err = CallFunction(conn, objDest, adapterPath, stopFunc)
			assert.NoError(t, err, "Unexpected Error for %s", stopFunc)
		})

	})

	t.Run("GetObjectProperty", func(t *testing.T) {
		t.Run("Failure", func(t *testing.T) {
			adapterPath := dbus.ObjectPath("/org/bluez/hcia")
			propPath := "org.bluez.Adapter1.Telephone"
			// err is not set for a property that isn't there
			prop, _ := GetObjectProperty(conn, objDest, adapterPath, propPath)
			// assert.Error(t, err, "Error for service %s and path %s: err: %s",
			// 	objDest, objPath, err)
			assert.Nil(t, prop)
		})
		t.Run("Success", func(t *testing.T) {
			adapterPath := dbus.ObjectPath("/org/bluez/hci0")
			propPath := "org.bluez.Adapter1.Address"
			prop, err := GetObjectProperty(conn, objDest, adapterPath, propPath)
			assert.NoError(t, err, "Error for service %s and path %s: err: %s",
				objDest, objPath, err)
			assert.NotNil(t, prop)
		})
	})

	t.Run("Watch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// We might get some info on the channel, but no guarantee. More importantly, we want to
		// test the expected
		signalCh := make(chan *dbus.Signal, 10)
		conn.Signal(signalCh)
		go func() {
			cancelled := false
			for !cancelled {
				select {
				case v := <-signalCh:
					fmt.Println("Received: ", v)
				case <-ctx.Done():
					// the conn will close the channel, so don't do this
					//close(b.busSignalCh)
					fmt.Println("quitting handler")
					cancelled = true
				}
			}
		}()
		err := Watch(conn, RootPath, ObjectManager, ObjectManagerFuncs.InterfacesAdded)
		assert.NoError(t, err, "Unexpected Error in Watch")
	})
}