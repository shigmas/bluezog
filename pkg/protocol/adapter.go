package protocol

import (
	//"context"
	"context"
	"fmt"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/bus"
	//	"github.com/shigmas/bluezog/pkg/logger"
)

type (
	// Adapter is a bluetooth adapter representation
	Adapter struct {
		BaseObject
		// This channel is currently passed back to the client, which should have
		// a goroutine to read it, receiving the new Object. But, for
		// experimentation/figuring things out, we pass this back in
		// StartDiscovery.
		discoveryCh  ObjectChangedChan
		cancelDisc   func()
		discoveryMux sync.Mutex
	}
)

func init() {
	typeRegistry[BluezInterface.Adapter] = func(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) Base {
		// need to fix the constructor
		return newAdapter(conn, name, data)
	}

}

func newAdapter(conn *bluezConn, name dbus.ObjectPath, data base.ObjectMap) *Adapter {
	return &Adapter{
		BaseObject: *newBaseObject(conn, name, BluezInterface.Adapter, data),
	}
}

// StartDiscovery on the adapter
func (a *Adapter) StartDiscovery() (ObjectChangedChan, error) {
	a.discoveryMux.Lock()
	defer a.discoveryMux.Unlock()
	if a.discoveryCh != nil {
		// We can't start discovery if it's already started.
		return nil, fmt.Errorf("Discovery already started")
	}

	//_, a.cancelDisc = context.WithCancel(context.Background())
	ch, err := a.bluez.AddWatch(a.Path,
		[]InterfaceSignalPair{
			{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesAdded},
			{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesRemoved},
		})
	if err != nil {
		//a.cancelDisc()
		return nil, err
	}
	a.discoveryCh = ch
	return ch, a.bluez.ops.CallFunction(context.Background(),
		BluezDest, a.Path, BluezAdapter.StartDiscovery)
}

// StopDiscovery on the adapter. This will disable getting any information from the devices
// connected through this adapter
func (a *Adapter) StopDiscovery() error {
	if a.discoveryCh == nil {
		return fmt.Errorf("Discovery not started")
	}
	// Remove ourselves as watchers. AddWatch created the channel, so it will
	// close the channel.
	a.bluez.RemoveWatch(a.Path, a.discoveryCh,
		[]InterfaceSignalPair{
			{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesAdded},
			{bus.ObjectManager,
				bus.ObjectManagerFuncs.InterfacesRemoved},
		})
	a.discoveryCh = nil
	//defer a.cancelDisc()

	return a.bluez.ops.CallFunction(context.Background(),
		BluezDest, a.Path, BluezAdapter.StopDiscovery)
}
