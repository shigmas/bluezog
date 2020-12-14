package protocol

import (
	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
)

type (
	// MediaTransport is a bluetooth device associated with an adapter
	MediaTransport struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.MediaTransport] = func(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) Base {
		return newMediaTransport(conn, name, data)

	}

}

func newMediaTransport(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) *MediaTransport {
	return &MediaTransport{
		BaseObject: *newBaseObject(conn, name, BluezInterface.MediaTransport, data),
	}
}
