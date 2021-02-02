package protocol

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// Base might be the only public 'object'.
	Base interface {
		GetPath() dbus.ObjectPath
		// checks the in memory dictionary for the property. It should correspond to the
		// implementation's property. Concrete implementations can use this to provide
		// typed accessors.
		Property(propName string) interface{}
		AllProperties() map[string]dbus.Variant
		// Typically, calls through to dbus to get the property
		FetchProperty(propName string) (interface{}, error)
		// In GetManagedObjects, the data has several interfaces, but only one is populated (so far).
		// That interface is the bluez interface (as opposed to the generic dbus interfaces).
		GetBluezInterface() string
		GetInterfaces() []string
		// Update is called from the main signal handler for updates to the objects in the registry
		Update(data base.ObjectMap) error
	}

	Connectable interface {
		Connect(ctx context.Context) error
		ConnectProfile(ctx context.Context, uuid string) error
		Disconnect(ctx context.Context) error
		DisconnectProfile(ctx context.Context, uuid string) error
	}

	// BaseObject has some objects and implementation so children don't need to implement the
	// Base interface
	BaseObject struct {
		bluez *bluezConn
		// Path is the debus object path that all objects will have
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

// The 'derived' objects are copying the same boilerplate for now, but as they get specialized, I think that
// tiny big of copied code can be customized.
func newBaseObject(
	conn *bluezConn,
	name dbus.ObjectPath,
	mainInterface string,
	data base.ObjectMap) *BaseObject {
	props, ok := data[mainInterface]
	if !ok {
		return nil
	}

	iFaces := make([]string, 0)
	for i := range data {
		iFaces = append(iFaces, i)
	}
	return &BaseObject{
		bluez:      conn,
		Path:       name,
		childType:  mainInterface,
		interfaces: iFaces,
		properties: props,
	}
}

// Update is called from the main signal handler for updates to the objects in the registry
func (b *BaseObject) Update(data base.ObjectMap) error {
	props, ok := data[BluezInterface.Adapter]
	if !ok {
		return fmt.Errorf("Data did not contain properties")
	}
	b.properties = props

	return nil
}

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

// AllProperties returns all (cached) properties
func (b *BaseObject) AllProperties() map[string]dbus.Variant {
	return b.properties
}

// FetchProperty for a type. Uses the childType member
func (b *BaseObject) FetchProperty(propName string) (interface{}, error) {
	propPath := fmt.Sprintf("%s.%s", b.childType, propName)
	return b.bluez.ops.GetObjectProperty(BluezDest, b.Path, propPath)
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
