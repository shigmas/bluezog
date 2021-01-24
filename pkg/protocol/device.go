package protocol

import (
	//"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/bus"
)

type (
	// Device is a bluetooth device associated with an adapter
	Device struct {
		BaseObject
		discoveryCh ObjectChangedChan
	}
)

func init() {
	typeRegistry[BluezInterface.Device] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newDevice(conn, name, data)
	}

}

func newDevice(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *Device {
	d := Device{
		BaseObject: *newBaseObject(conn, name, BluezInterface.Device, data),
	}

	//err = conn.AddWatch(BluezInterface.Device, bus.ObjectManagerFuncs.InterfacesAdded)

	return &d
}

// Connect to the device
func (d *Device) Connect() error {
	err := d.bluez.ops.CallFunction(BluezDest, d.Path, BluezDevice.Connect)
	if err != nil {
		return err
	}
	d.discoveryCh, err = d.bluez.AddWatch(d.Path,
		[]InterfaceSignalPair{
			InterfaceSignalPair{bus.Properties,
				bus.PropertiesFuncs.PropertiesChanged},
		})

	return err
}

// Disconnect from the device
func (d *Device) Disconnect() error {
	err := d.bluez.ops.CallFunction(BluezDest, d.Path, BluezDevice.Disconnect)
	if err != nil {
		return err
	}
	d.discoveryCh, err = d.bluez.AddWatch(d.Path,
		[]InterfaceSignalPair{
			InterfaceSignalPair{bus.Properties,
				bus.PropertiesFuncs.PropertiesChanged},
		})

	return err
}

// ConnectProfile connects to the device for the specificed UUID
func (d *Device) ConnectProfile(uuid string) error {
	// The specs don't have a return value, but we get one. So, let's see what this is.
	var ret int
	err := d.bluez.ops.CallFunctionWithArgs(&ret, BluezDest, d.Path, BluezDevice.ConnectProfile, uuid)
	// d.discoveryCh, err = d.bluez.AddWatch(d.Path,
	// 	[]InterfaceSignalPair{
	// 		InterfaceSignalPair{bus.Properties,
	// 			bus.PropertiesFuncs.PropertiesChanged},
	// 	})

	return err
}

// DisconnectProfile disconnects from the device for the specificed UUID
func (d *Device) DisconnectProfile(uuid string) error {
	return d.bluez.ops.CallFunctionWithArgs(nil, BluezDest, d.Path, BluezDevice.DisconnectProfile, uuid)
}

// GetProperty gets the property by key
func (d *Device) GetProperty(prop string) (interface{}, error) {
	variant, ok := d.properties[prop]
	if !ok {
		return nil, fmt.Errorf("No property %s", prop)
	}

	return variant.Value(), nil
}
