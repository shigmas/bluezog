package bus

// D-Bus is an OO RPC interface, but the API is really just strings, so these are just
// functions. I found that the D-Bus API and Bluez aren't the greatest combination. The
// separation is good, but conforming to the API was hard for me to grasp
//
// As a wrapper over dbus, This is perhaps too thin of a wrapper, but for testing, it's
// very useful to mock these functions over the real API.

import (
	"encoding/xml"
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/logger"
	"github.com/shigmas/bluezog/test"
)

type (
	// DbusOperations is the Dbus implementation of Operations
	DbusOperations struct {
		conn *dbus.Conn
	}
)

var (
	_ base.Operations = (*DbusOperations)(nil)
)

// NewDbusOperations creates a DbusOperations instance which implements Operations
func NewDbusOperations() base.Operations {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil
	}
	return &DbusOperations{
		conn: conn,
	}
}

// IntrospectObject fetches the XMM for Introspection and parses it into a Node hierarchy
func (d *DbusOperations) IntrospectObject(dest string, objPath dbus.ObjectPath) (*base.Node, error) {
	var s string
	err := d.conn.Object(dest, objPath).Call(IntrospectableFuncs.Introspect, 0).Store(&s)
	if err != nil {
		return nil, err
	}

	var node base.Node
	b := []byte(s)
	err = xml.Unmarshal(b, &node)
	if err != nil {
		return nil, err
	}
	if base.DumpData {
		_, err := test.MarshalIntrospect(&node)
		if err != nil {
			logger.Info("Unable to marshal introspect: %s", err)
		}
	}

	return &node, nil
}

// GetObjectProperty for the specified object and property name
func (d *DbusOperations) GetObjectProperty(dest string, objPath dbus.ObjectPath, propName string) (interface{}, error) {
	val, err := d.conn.Object(dest, objPath).GetProperty(propName)
	if err != nil {
		return nil, err
	}

	return val.Value(), nil
}

// GetManagedObjects retrieves the paths of the objects managed by this object
func (d *DbusOperations) GetManagedObjects(dest string, objPath dbus.ObjectPath) (map[dbus.ObjectPath]base.ObjectMap, error) {
	var s map[dbus.ObjectPath]base.ObjectMap
	err := d.conn.Object(dest, objPath).Call(ObjectManagerFuncs.GetManagedObjects, 0).Store(&s)
	if err != nil {
		logger.Debug("%s error: %s\n", ObjectManagerFuncs.GetManagedObjects, err)
		return nil, err
	}

	if base.DumpData {
		_, err := test.MarshalManagedObjects(s)
		if err != nil {
			logger.Info("Unable to marshal ManagedObjects: %s", err)
		}
	}

	return s, nil
}

// CallFunction is exposes the simplest and common way to call a function on the object
// The other functions should probably call this one with the hardcoded name. But, they
// actually know what they should receive. This just receives nothing. There should
// be a function with expected return values.
// Should have a func struct with the conn, dest, path.
func (d *DbusOperations) CallFunction(dest string, objPath dbus.ObjectPath, funcName string) error {
	logger.Debug("%s: Call parameters %s, %s", funcName, dest, string(objPath))
	err := d.conn.Object(dest, objPath).Call(funcName, 0).Store()
	return err
}

// CallFunctionWithArgs is simply CallFunction with arbitrary arguments
func (d *DbusOperations) CallFunctionWithArgs(
	retVal interface{},
	dest string,
	objPath dbus.ObjectPath,
	funcName string,
	args ...interface{}) error {
	logger.Debug("%s: CallWithArgs parameters %s, %s", funcName, dest, string(objPath))
	var flags dbus.Flags
	if retVal == nil {
		fmt.Println("retVal is nil")
		flags = dbus.FlagNoReplyExpected
	}
	err := d.conn.Object(dest, objPath).Call(funcName, flags, args...).Store(retVal)
	return err
}

// RegisterSignalChannel passes the signal to DBus
func (d *DbusOperations) RegisterSignalChannel(ch chan<- *dbus.Signal) {
	d.conn.Signal(ch)
}

// Watch is a simplified version of AddMatchsignal
func (d *DbusOperations) Watch(path dbus.ObjectPath, iface string, method string) error {
	return d.conn.AddMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember(method))
}

// UnWatch is a simplified version of RemoveMatchsignal
func (d *DbusOperations) UnWatch(path dbus.ObjectPath, iface string, method string) error {
	return d.conn.RemoveMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember(method))
}
