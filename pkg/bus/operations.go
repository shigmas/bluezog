package bus

// D-Bus is an OO RPC interface, but the API is really just strings, so these are just
// functions. I found that the D-Bus API and Bluez aren't the greatest combination. The
// separation is good, but conforming to the API was hard for me to grasp
//
// As a wrapper over dbus, This is perhaps too thin of a wrapper, but for testing, it's
// very useful to mock these functions over the real API.

import (
	"encoding/xml"

	"github.com/godbus/dbus/v5"

	"github.com/shigmas/bluezog/pkg/logger"
)

// GetObject fetches the introspection information for the object
func GetObject(conn *dbus.Conn, dest string, objPath dbus.ObjectPath) (*Node, error) {
	var s string
	err := conn.Object(dest, objPath).Call(IntrospectableFuncs.Introspect, 0).Store(&s)
	if err != nil {
		return nil, err
	}

	var node Node
	b := []byte(s)
	err = xml.Unmarshal(b, &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetObjectProperty for the specified object and property name
func GetObjectProperty(conn *dbus.Conn, dest string, objPath dbus.ObjectPath,
	propName string) (interface{}, error) {
	val, err := conn.Object(dest, objPath).GetProperty(propName)
	if err != nil {
		return nil, err
	}

	return val.Value(), nil
}

// GetManagedObjects retrieves the paths of the objects managed by this object
func GetManagedObjects(conn *dbus.Conn, dest string, objPath dbus.ObjectPath) (map[dbus.ObjectPath]ObjectMap, error) {
	var s map[dbus.ObjectPath]ObjectMap
	err := conn.Object(dest, objPath).Call(ObjectManagerFuncs.GetManagedObjects, 0).Store(&s)
	if err != nil {
		logger.Debug("%s error: %s\n", ObjectManagerFuncs.GetManagedObjects, err)
		return nil, err
	}

	return s, nil
}

// CallFunction is exposes the simplest and common way to call a function on the object
// The other functions should probably call this one with the hardcoded name. But, they
// actually know what they should receive. This just receives nothing. There should
// be a function with expected return values.
// Should have a func struct with the conn, dest, path.
func CallFunction(conn *dbus.Conn, dest string, objPath dbus.ObjectPath, funcName string) error {
	logger.Debug("%s: Call parameters %s, %s", funcName, dest, string(objPath))
	err := conn.Object(dest, objPath).Call(funcName, 0).Store()
	return err
}

// CallFunctionWithArgs is simply CallFunction with arbitrary arguments
func CallFunctionWithArgs(conn *dbus.Conn, dest string, objPath dbus.ObjectPath, funcName string,
	args ...interface{}) error {
	err := conn.Object(dest, objPath).Call(funcName, 0, args...).Store()
	return err
}

// Watch is a simplified version of AddMatchsignal
func Watch(conn *dbus.Conn, path dbus.ObjectPath, iface string, method string) error {
	return conn.AddMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember(method))
}

// Watch is a simplified version of AddMatchsignal
func UnWatch(conn *dbus.Conn, path dbus.ObjectPath, iface string, method string) error {
	return conn.RemoveMatchSignal(
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember(method))
}
