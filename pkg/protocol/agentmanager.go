package protocol

import (
	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
)

type (
	// AgentManager is a bluetooth device associated with an adapter
	AgentManager struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.AgentManager] = func(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) Base {
		return newAgentManager(conn, name, data)

	}

}

func newAgentManager(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) *AgentManager {
	return &AgentManager{
		BaseObject: *newBaseObject(conn, name, BluezInterface.AgentManager, data),
	}
}
