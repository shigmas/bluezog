package protocol

import (
	//"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
	"github.com/shigmas/bluezog/pkg/logger"
)

type (
	// Device is a bluetooth device associated with an adapter
	Device struct {
		BaseObject
		// When these become real values, this will be a separate file
		properties map[string]dbus.Variant
	}
)

func init() {
	fmt.Printf("Registering device type: %s\n", BluezInterface.Device)
	typeRegistry[BluezInterface.Device] = func(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) Base {
		devData, ok := data[BluezInterface.Device]
		if !ok {
			return nil
		}
		return newDevice(conn, devData, data)
	}

}

func newDevice(conn *bluezConn, deviceDict map[string]dbus.Variant, data bus.ObjectMap) *Device {
	path, err := GetDevicePath(deviceDict)
	if err != nil {
		logger.Error("Unable to get properties from dictionary: %s", err)
		return nil
	}
	d := Device{
		BaseObject: BaseObject{
			conn:       conn,
			Path:       path,
			childType:  BluezInterface.Device,
			properties: data[BluezInterface.Device],
		},
		properties: deviceDict,
	}

	//err = conn.AddWatch(BluezInterface.Device, bus.ObjectManagerFuncs.InterfacesAdded)

	return &d
}

// Update updates the object. This Base interface function is not in BaseObject and needs
// to be implemented
func (d *Device) Update(data []interface{}) {
}

// Connect to the device
func (d *Device) Connect() error {
	return bus.CallFunction(d.conn.busConn, BluezDest, d.Path, BluezDevice.Connect)
}

// GetProperty gets the property by key
func (d *Device) GetProperty(prop string) (interface{}, error) {
	variant, ok := d.properties[prop]
	if !ok {
		return nil, fmt.Errorf("No property %s", prop)
	}

	return variant.Value(), nil
}
