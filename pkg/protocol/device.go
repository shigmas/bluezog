package protocol

import (
	//"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
)

type (
	// Device is a bluetooth device associated with an adapter
	Device struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.Device] = func(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) Base {
		return newDevice(conn, name, data)
	}

}

func newDevice(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) *Device {
	d := Device{
		BaseObject: *newBaseObject(conn, name, BluezInterface.Device, data),
	}

	//err = conn.AddWatch(BluezInterface.Device, bus.ObjectManagerFuncs.InterfacesAdded)

	return &d
}

// Connect to the device
func (d *Device) Connect() error {
	return bus.CallFunction(d.conn.busConn, BluezDest, d.Path, BluezDevice.Connect)
}

func (d *Device) ConnectProfile(uuid string) error {
	return bus.CallFunctionWithArgs(d.conn.busConn, BluezDest, d.Path, BluezDevice.ConnectProfile, uuid)
}

// GetProperty gets the property by key
func (d *Device) GetProperty(prop string) (interface{}, error) {
	variant, ok := d.properties[prop]
	if !ok {
		return nil, fmt.Errorf("No property %s", prop)
	}

	return variant.Value(), nil
}
