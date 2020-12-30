package protocol

import (
	"context"
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
	//	"github.com/shigmas/bluezog/pkg/logger"
)

type (
	// Adapter is a bluetooth adapter representation
	Adapter struct {
		BaseObject
		discoveryCh  ObjectChangedChan
		cancelDisc   func()
		discoveryMux sync.Mutex
	}
)

func init() {
	typeRegistry[BluezInterface.Adapter] = func(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) Base {
		// need to fix the constructor
		return newAdapter(conn, name, data)
	}

}

func newAdapter(conn *bluezConn, name dbus.ObjectPath, data bus.ObjectMap) *Adapter {
	return &Adapter{
		BaseObject: *newBaseObject(conn, name, BluezInterface.Adapter, data),
	}
}

// StartDiscovery on the adapter
func (a *Adapter) StartDiscovery() (ObjectChangedChan, error) {
	a.discoveryMux.Lock()
	defer a.discoveryMux.Unlock()
	if a.cancelDisc != nil {
		// We can't start discovery if it's already started.
		return nil, fmt.Errorf("Discovery already started")
	}

	_, a.cancelDisc = context.WithCancel(context.Background())
	ch, err := a.conn.AddWatch(a.Path,
		[]InterfaceSignalPair{
			InterfaceSignalPair{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesAdded},
			InterfaceSignalPair{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesRemoved},
		})
	if err != nil {
		a.cancelDisc()
		return nil, err
	}
	a.discoveryCh = ch
	return ch, bus.CallFunction(a.conn.busConn, BluezDest, a.Path, BluezAdapter.StartDiscovery)
}

// StopDiscovery on the adapter. This will disable getting any information from the devices
// connected through this adapter
func (a *Adapter) StopDiscovery() error {
	if a.discoveryCh == nil {
		return fmt.Errorf("Discovery not started")
	}
	// Remove ourselves as watchers
	a.conn.RemoveWatch(a.Path, a.discoveryCh,
		[]InterfaceSignalPair{
			InterfaceSignalPair{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesAdded},
			InterfaceSignalPair{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesRemoved},
		})
	a.discoveryCh = nil
	defer a.cancelDisc()

	return bus.CallFunction(a.conn.busConn, BluezDest, a.Path, BluezAdapter.StopDiscovery)
}

// Connect to the device at the address. The address is of the form "HH:HH:HH:HH:HH:HH".
func (a *Adapter) Connect(address string) error {
	return bus.CallFunctionWithArgs(nil, a.conn.busConn, BluezDest, a.Path, BluezAdapter.Connect, address)

}
