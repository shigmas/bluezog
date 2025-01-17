package protocol

import (
	"context"
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/bus"
)

type (
	// GattCharacteristic is a bluetooth device associated with a GATT Characteristic.
	GattCharacteristic struct {
		BaseObject
		notifyMux sync.Mutex
		notifyCh  ObjectChangedChan
	}
)

func init() {
	typeRegistry[BluezInterface.GATTCharacteristic] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newGattCharacteristic(conn, name, data)

	}

}

func newGattCharacteristic(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *GattCharacteristic {
	return &GattCharacteristic{
		BaseObject: *newBaseObject(conn, name, BluezInterface.GATTCharacteristic, data),
	}
}

// ReadValue reads the value from the characteristic. The function takes a dict,
// but since the client only takes the offset, that's all we provide here.
func (gc *GattCharacteristic) ReadValue(ctx context.Context, offset uint16) ([]byte, error) {
	// gatt /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C/service0026/char0035
	// command gatt returned error [Method "ReadValue" with signature "a{sv}" on interface "(null)" doesn't exist
	// ]
	// zogctl> gatt /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C/service0026/char0031
	// command gatt returned error [Method "ReadValue" with signature "a{sv}" on interface "(null)" doesn't exist
	// ]

	args := map[string]interface{}{
		"offset": offset,
	}
	// This works for the ALP's sensor (SNM00). Hard to add this as a CLI, so hardcoding it for now,
	// but is one of the last things that should be hardcoded.
	val := make([]byte, 4)
	err := gc.bluez.ops.CallFunctionWithArgs(ctx, val, BluezDest, gc.Path,
		BluezGATTCharacteristic.ReadValue, args)

	return val, err
}

// StartNotify will start receiving notifications for this characteristic
func (gc *GattCharacteristic) StartNotify() error {
	gc.notifyMux.Lock()
	defer gc.notifyMux.Unlock()
	if gc.notifyCh != nil {
		return fmt.Errorf("Discovery already started")
	}

	ch, err := gc.bluez.AddWatch(gc.Path,
		[]InterfaceSignalPair{
			{bus.Properties,
				bus.PropertiesFuncs.PropertiesChanged},
		})
	if err != nil {
		return err
	}
	gc.notifyCh = ch

	return gc.bluez.ops.CallFunction(context.Background(), BluezDest, gc.Path,
		BluezGATTCharacteristic.StartNotify)

}

// StopNotify will stop receiving notifications for this characteristic
func (gc *GattCharacteristic) StopNotify() error {
	return gc.bluez.ops.CallFunction(context.Background(), BluezDest, gc.Path,
		BluezGATTCharacteristic.StopNotify)
}
