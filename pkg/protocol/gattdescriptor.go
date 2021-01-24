package protocol

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// GattDescriptor is a bluetooth device associated with a GATT Descriptor.
	GattDescriptor struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.GATTDescriptor] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newGattDescriptor(conn, name, data)

	}

}

func newGattDescriptor(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *GattDescriptor {
	return &GattDescriptor{
		BaseObject: *newBaseObject(conn, name, BluezInterface.GATTDescriptor, data),
	}
}

// ReadValue reads the value from the descriptor. The function takes a dict,
// but since the client only takes the offset, that's all we provide here.
func (gc *GattDescriptor) ReadValue(ctx context.Context, offset uint16) ([]byte, error) {
	args := map[string]interface{}{
		"offset": offset,
	}
	var val []byte
	err := gc.bluez.ops.CallFunctionWithArgs(ctx, val, BluezDest, gc.Path, BluezGATTDescriptor.ReadValue, args)

	return val, err
}
