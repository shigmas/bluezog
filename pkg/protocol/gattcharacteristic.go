package protocol

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// GattCharacteristic is a bluetooth device associated with a GATT Characteristic.
	GattCharacteristic struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.GATTCharacteristic] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newGattCharacteristic(conn, name, data)

	}

}

func newGattCharacteristic(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *GattCharacteristic {
	fmt.Println("Creating ", BluezInterface.GATTCharacteristic)
	return &GattCharacteristic{
		BaseObject: *newBaseObject(conn, name, BluezInterface.GATTCharacteristic, data),
	}
}

// ReadValue reads the value from the characteristic. The function takes a dict,
// but since the client only takes the offset, that's all we provide here.
func (gc *GattCharacteristic) ReadValue(offset uint16) ([]byte, error) {
	args := map[string]interface{}{
		"offset": offset,
	}
	var val []byte
	err := gc.bluez.ops.CallFunctionWithArgs(val, BluezDest, gc.Path, BluezGATTCharacteristic.ReadValue, args)

	return val, err
}
