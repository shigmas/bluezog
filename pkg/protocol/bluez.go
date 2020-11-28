package protocol

import (
	"context"
	"fmt"
	"path"
	"strings"

	//	"reflect"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/bus"
	"github.com/shigmas/bluezog/pkg/logger"
)

var (
	// ChannelBufferSize specifies the efault sizes of the ChannelBuffer for all channels
	ChannelBufferSize = 8
)

type (
	// XXX - normalize the information that we get from the bus.
	// 1. XML data from introspect: Nodes are what are Unmarshalled from XML.
	// 2. Data from GetManagedObjects. This is a map: ObjectPath -> bus.ObjectMap (aka map of
	//    string to values). The map is interface to a map of properties. e.g.
	//    - Introspectable: no props (maybe it needs to be introspected?)
	//    - Device1: <path, address, etc
	//    - DBus.Properties: empty
	// 3. Data from signals: This is a dbus.Signal, which provides the
	//    - Sender:
	//    - Path: The path
	//    - Name:
	//    - Body: contents, which is just a slice of interfaces.
	// Objects can be created from 2 or 3. But, 2 seems to provide the more complete
	// information, and 3 seems to be the updates.

	// ObjectChangedData is the data that we write to any listeners of signals that we
	// get from dbus. This data is usually triggered through a dbus function, such as
	// discovery (called through the adapter) or on a device.
	ObjectChangedData struct {
		// Path is the dbus.ObjectPath
		Path dbus.ObjectPath
		// Object is an implementation of the Base interface
		Object Base
		// Signal is the name of the signal that triggered this
		Signal string
		// Maybe we might need the original data. Like an updated property or something
	}

	// ObjectChangedChan receives data from a signal watcher
	ObjectChangedChan chan ObjectChangedData

	// Bluez is the root object. The interface is minimal because most interaction will be
	// done through the adapter. (The other objects will interact directly with the
	// implementation of this interface.)
	Bluez interface {
		// FindAdapters that exist
		FindAdapters() []*Adapter
		// AddWatch will watch a path on the signals and will return a channel that we will use
		// to communicate the data to the listener.
		AddWatch(
			path dbus.ObjectPath,
			interfaceName string,
			signalNames []string) (ObjectChangedChan, error)
		// RemoveWatch will remove listener on the path.
		// XXX: Implementation hole: We won't actually remove the watch. But no one will be
		// notified because no one is listening. We should reference count the signals, and stop
		// watching when it reaches zero.
		RemoveWatch(
			path dbus.ObjectPath,
			ch ObjectChangedChan,
			interfaceName string,
			signalNames []string) error
	}

	// Three purposes:
	// - Hold the connection to DBus for bluez.
	//   - We will send requests through our minimal DBus wrapper
	//   - We will receive signals through the channel. We provide an Add/Remove interface for
	//     other objects and users to modify which signals we are interested in.
	// - Be our own client for the Adapter type. We will have a channel which receives the
	//   data from our signal receiver channel for the Adapter type.
	// - Keep an object registry of existing objects. But, maybe this should be a tree and we
	//   just hold the root node. I think this is more dbus-ish anyway.
	// This is an implementation of the Bluez interface.
	bluezConn struct {
		busConn *dbus.Conn
		// Objects on this bus. ObjectMap are the interfaces that each
		// object implements.
		root *bus.Node
		// Objects known to this connection.
		objectRegistry map[dbus.ObjectPath]Base
		busSignalCh    chan *dbus.Signal
		// Should be a map to slice
		signalWatchers map[dbus.ObjectPath]ObjectChangedChan
	}

	typeConstructorFn func(*bluezConn, dbus.ObjectPath, bus.ObjectMap) Base
)

var _ Bluez = (*bluezConn)(nil)

var (
	typeRegistry = make(map[string]typeConstructorFn)
)

// AddressToPath converts the ":" delimited address to the full path.
func AddressToPath(parent string, address string) dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(parent, strings.ReplaceAll(address, ":", "_")))
}

// InitializeBluez creates a Bluez implementation
func InitializeBluez(ctx context.Context) (Bluez, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	node, err := bus.GetObject(conn, BluezDest, BluezRootPath)
	if err != nil {
		return nil, err
	}

	// Initial objects for the registry
	objMap, err := bus.GetManagedObjects(conn, BluezDest, bus.RootPath)
	if err != nil {
		return nil, err
	}

	bluezObj := bluezConn{
		busConn:        conn,
		root:           node,
		objectRegistry: make(map[dbus.ObjectPath]Base, len(objMap)),
		busSignalCh:    make(chan *dbus.Signal, 10),
		signalWatchers: make(map[dbus.ObjectPath]ObjectChangedChan, 10),
	}

	for path, ifaceMap := range objMap {
		// The ifaceMap is interface: data. We choose the first interface with data.
		var newObj Base
		for iface, data := range ifaceMap {
			ctor, ok := typeRegistry[iface]
			if len(data) > 0 && ok {
				// no need to pass in the main interface type since that's keyed
				// by the constructor
				newObj = ctor(&bluezObj, path, ifaceMap)
				break
			}
		}
		if newObj == nil {
			fmt.Printf("No interface constructor found: %s: %s\n", path, ifaceMap)
		} else {
			bluezObj.objectRegistry[path] = newObj
		}

	}

	bluezObj.busConn.Signal(bluezObj.busSignalCh)
	go bluezObj.handleSignals(ctx)

	return &bluezObj, nil
}

// XXX - this should just be AddWatch and takes a signal. We only pass back
// channels
// AddWatcher takes a channel and the string for the signal that we are watching for.
// So far, we only listen on org.freedesktop.DBus.ObjectManager signals.
// For expanding, we parse sig into object and signal.
func (b *bluezConn) AddWatch(
	path dbus.ObjectPath,
	interfaceName string,
	signalNames []string) (ObjectChangedChan, error) {

	for _, s := range signalNames {
		err := bus.Watch(b.busConn, bus.RootPath, interfaceName, s)
		if err != nil {
			return nil, err
		}
	}

	ch := make(ObjectChangedChan)
	b.signalWatchers[path] = ch

	return ch, nil
}

func (b *bluezConn) RemoveWatch(
	path dbus.ObjectPath,
	ch ObjectChangedChan,
	interfaceName string,
	signalNames []string) error {
	ch, ok := b.signalWatchers[path]
	if !ok {
		return fmt.Errorf("No channel found for %s", path)
	}
	delete(b.signalWatchers, path)
	close(ch)
	for _, s := range signalNames {
		err := bus.UnWatch(b.busConn, bus.RootPath, interfaceName, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *bluezConn) GetObjectsByType(oType string) []Base {
	objects := make([]Base, 0)
	for _, v := range b.objectRegistry {
		if v.GetBluezInterface() == oType {
			objects = append(objects, v)
		}
	}
	fmt.Printf("Found %d objects of type %s\n", len(objects), oType)
	return objects
}

func (b *bluezConn) FindAdapters() []*Adapter {
	objects := b.GetObjectsByType(BluezInterface.Adapter)
	adapters := make([]*Adapter, len(objects))
	for i, o := range objects {
		a, ok := o.(*Adapter)
		if !ok {
			// This is an internal consistency problem. i.e. a bug
			fmt.Printf("Object registered as Adapter, but could not cast as adapter\n")
		} else {
			adapters[i] = a
		}
	}
	return adapters
}

// While this reads channel that is passed to dbus, any other gooutine can pass the
// same data.
func (b *bluezConn) handleSignals(ctx context.Context) {
	cancelled := false
	for !cancelled {
		select {
		case sigData := <-b.busSignalCh:
			logger.Info("Received: %s and %s", sigData.Name, sigData.Path)
			obj, ok := b.objectRegistry[sigData.Path]
			if ok {
				obj.Update(sigData.Body)
			}
			// Actually, if it doesn't exist, we need to get the managed objects?
		case <-ctx.Done():
			// the conn will close the channel, so don't do this
			//close(b.busSignalCh)
			logger.Info("quitting handler")
			cancelled = true
		}
	}
}
