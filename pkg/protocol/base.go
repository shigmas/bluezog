package protocol

import (
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
)

const (
	// BluezDest is the destination required for all(?) D-Bus calls
	BluezDest = "org.bluez"
	// BluezRootPath is the root of the bluez objects
	BluezRootPath = "/org/bluez"

	// InterfacesAddedSignalKey is the signal when an interface is added (device. could be an adapter)
	InterfacesAddedSignalKey = "InterfacesAdded"
	// InterfacesRemovedSignalKey is the signal when an interface is removed.
	InterfacesRemovedSignalKey = "InterfacesRemoved"
)

type (
	// These are internal classes to make it look like scoped constants. They are accessed
	// as BluezInterface and BluezAdapter
	bluezInterface struct {
		Adapter string
		Device  string
	}

	bluezAdapter struct {
		StartDiscovery string
		StopDiscovery  string
		Connect        string
		AddressProp    string
		AliasProp      string
	}

	bluezDevice struct {
		Connect              string
		AddressProp          string
		AddressTypeProp      string
		BlockedProp          string
		ConnectedProp        string
		UUIDsProp            string
		AdapterProp          string
		ServiceDataProp      string
		AliasProp            string
		PairedProp           string
		TrustedProp          string
		LegacyPairingProp    string
		RSSIProp             string
		ServicesResolvedProp string
	}

	// Base might be the only public 'object'.
	Base interface {
		GetPath() dbus.ObjectPath
		// checks the in memory dictionary for the property. It should correspond to the
		// implementation's property. Concrete implementations can use this to provide
		// typed accessors.
		Property(propName string) interface{}
		// Typically, calls through to dbus to get the property
		FetchProperty(propName string) (interface{}, error)
		// In GetManagedObjects, the data has several interfaces, but only one is populated (so far).
		// That interface is the bluez interface (as opposed to the generic dbus interfaces).
		GetBluezInterface() string
		GetInterfaces() []string
		// Update is called from the main signal handler for updates to the objects in the registry
		Update(data []interface{})
	}

	// BaseObject has some objects and implementation so children don't need to implement the
	// Base interface
	BaseObject struct {
		// Let's see if we can remove this
		conn *bluezConn
		// Path is the debus object path that all objects will ahve
		Path dbus.ObjectPath
		// string in the constants
		childType string
		// The interfaces implemented by this object. Possibly, we could have an Golang interface for
		// all of the interfaces, but that will be left to the users of this library. We will only
		// provide an implementation for one
		interfaces []string
		properties map[string]dbus.Variant
	}
)

var (
	// BluezInterface is the constants for the base interface
	BluezInterface = bluezInterface{
		Adapter: BluezDest + ".Adapter1",
		Device:  BluezDest + ".Device1",
	}

	// BluezAdapter are the constants for the adapter
	BluezAdapter = bluezAdapter{
		StartDiscovery: BluezInterface.Adapter + ".StartDiscovery",
		StopDiscovery:  BluezInterface.Adapter + ".StopDiscovery",
		Connect:        BluezInterface.Adapter + ".Connect",
		// Address:        BluezInterface.Adapter + ".Address",
		// Alias:          BluezInterface.Adapter + ".Alias",
		AddressProp: "Address",
		AliasProp:   "Alias",
	}

	// BluezDevice are the constants in the BluezInterface.Device interface
	BluezDevice = bluezDevice{
		Connect:              BluezInterface.Device + ".Connect",
		AddressProp:          "Address",
		AddressTypeProp:      "AddressType",
		BlockedProp:          "Blocked",
		ConnectedProp:        "Connected",
		UUIDsProp:            "UUIDs",
		AdapterProp:          "Adapter",
		ServiceDataProp:      "ServiceData",
		AliasProp:            "Alias",
		PairedProp:           "Paired",
		TrustedProp:          "Trusted",
		LegacyPairingProp:    "LegacyPairing",
		RSSIProp:             "RSSI",
		ServicesResolvedProp: "ServicesResolved",
	}
)

// GetPath returns the unique path for this object
func (b *BaseObject) GetPath() dbus.ObjectPath {
	return b.Path
}

// Property returns the property value variant as an interface. nil if it wasn't
// found locally
func (b *BaseObject) Property(propName string) interface{} {
	prop, ok := b.properties[propName]
	if !ok {
		return nil
	}

	return prop.Value()
}

// FetchProperty for a type. Uses the childType member
func (b *BaseObject) FetchProperty(propName string) (interface{}, error) {
	propPath := fmt.Sprintf("%s.%s", b.childType, propName)
	return bus.GetObjectProperty(b.conn.busConn, BluezDest, b.Path,
		propPath)
}

// GetBluezInterface returns the bluez type that this object was created as. (
// Always (or at least, usually), there is only one applicable bluez type
func (b *BaseObject) GetBluezInterface() string {
	return b.childType
}

// GetInterfaces retries the interfaces that this object provides
func (b *BaseObject) GetInterfaces() []string {
	return b.interfaces
}

// GetDevicePath will build the dbus.ObjectPath from the data.
func GetDevicePath(propDict map[string]dbus.Variant) (dbus.ObjectPath, error) {
	adapterVar, ok := propDict[BluezDevice.AdapterProp]
	if !ok {
		return "", fmt.Errorf("No property %s", BluezDevice.AdapterProp)
	}
	addrVar, ok := propDict[BluezDevice.AddressProp]
	if !ok {
		return "", fmt.Errorf("No property %s", BluezDevice.AddressProp)
	}

	return AddressToPath(string(adapterVar.Value().(dbus.ObjectPath)), addrVar.Value().(string)), nil
}
