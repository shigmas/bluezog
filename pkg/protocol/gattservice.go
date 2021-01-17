package protocol

import (
	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// GattService is a bluetooth device associated with a GATT Service. This is
	// the parent of the GattCharacteristic in the hierarchy
	GattService struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.GATTService] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newGattService(conn, name, data)

	}

}

func newGattService(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *GattService {
	return &GattService{
		BaseObject: *newBaseObject(conn, name, BluezInterface.GATTService, data),
	}
}
