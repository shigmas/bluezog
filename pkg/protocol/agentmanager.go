package protocol

import (
	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// AgentManager is a bluetooth device associated with an adapter
	AgentManager struct {
		BaseObject
	}
)

func init() {
	typeRegistry[BluezInterface.AgentManager] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		return newAgentManager(conn, name, data)

	}

}

func newAgentManager(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *AgentManager {
	return &AgentManager{
		BaseObject: *newBaseObject(conn, name, BluezInterface.AgentManager, data),
	}
}
