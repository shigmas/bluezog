package protocol

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"

	//	"reflect"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/bus"
	"github.com/shigmas/bluezog/pkg/logger"
	"github.com/shigmas/bluezog/test"
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
		// FindAdapters thcat exist
		FindAdapters() []*Adapter
		GetObjectsByType(oType string) []Base

		IntrospectPath(path string) (*base.Node, error)
		GetManagedObjects(path string) (map[dbus.ObjectPath]base.ObjectMap, error)

		// These return the objects if they the interface named interfaceName or
		// all the objects. Likewise, if property is empty, all the objects. If the
		// property is not empty, we will return all objects with the property. If the
		// value is not nil, we will return the objects that match the property *and*
		// value
		GetObjectsByInterface(interfaceName string) []Base
		FindObjects(pattern string, firstOnly bool) []Base
		//IntrospectObject()

		// AddWatch will watch a path on the signals and will return a channel that we will use
		// to communicate the data to the listener.
		AddWatch(
			path dbus.ObjectPath,
			signalMap []InterfaceSignalPair) (ObjectChangedChan, error)
		// RemoveWatch will remove listener on the path.
		// XXX: Implementation hole: We won't actually remove the watch. But no one will be
		// notified because no one is listening. We should reference count the signals, and stop
		// watching when it reaches zero.
		RemoveWatch(
			path dbus.ObjectPath,
			ch ObjectChangedChan,
			signalMap []InterfaceSignalPair) error
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
		ops base.Operations
		// Objects on this bus. ObjectMap are the interfaces that each
		// object implements.
		root *base.Node
		// Objects known to this connection.
		objectRegistry map[dbus.ObjectPath]Base
		registryMux    sync.RWMutex
		busSignalCh    chan *dbus.Signal
		// Should be a map to slice
		signalWatchers map[dbus.ObjectPath]ObjectChangedChan
		sigWatchersMux sync.RWMutex
	}

	typeConstructorFn func(*bluezConn, dbus.ObjectPath, base.ObjectMap) Base
)

var _ Bluez = (*bluezConn)(nil)

var (
	typeRegistry = make(map[string]typeConstructorFn)
)

// AddressToPath converts the ":" delimited address to the full path.
func AddressToPath(parent string, address string) dbus.ObjectPath {
	return dbus.ObjectPath(path.Join(parent, strings.ReplaceAll(address, ":", "_")))
}

func newObjectChangedData(path dbus.ObjectPath, obj Base, sigName string) ObjectChangedData {
	return ObjectChangedData{
		Path:   path,
		Object: obj,
		Signal: sigName,
	}
}

// InitializeBluez creates a Bluez implementation
func InitializeBluez(ctx context.Context, ops base.Operations) (Bluez, error) {
	node, err := ops.IntrospectObject(BluezDest, BluezRootPath)
	if err != nil {
		return nil, err
	}

	// Initial objects for the registry
	objMap, err := ops.GetManagedObjects(BluezDest, bus.RootPath)
	if err != nil {
		return nil, err
	}

	bluezObj := bluezConn{
		ops:            ops,
		root:           node,
		objectRegistry: make(map[dbus.ObjectPath]Base, len(objMap)),
		busSignalCh:    make(chan *dbus.Signal, 10),
		signalWatchers: make(map[dbus.ObjectPath]ObjectChangedChan, 10),
	}

	// we're locking too long, but no one else has object yet
	bluezObj.registryMux.Lock()
	for path, ifaceMap := range objMap {
		newObj := bluezObj.createObject(path, ifaceMap)
		if newObj == nil {
			fmt.Printf("No interface constructor found: %s: %s\n", path, ifaceMap)
		} else {
			bluezObj.objectRegistry[path] = newObj
		}
	}
	bluezObj.registryMux.Unlock()

	bluezObj.ops.RegisterSignalChannel(bluezObj.busSignalCh)
	go bluezObj.handleSignals(ctx)

	return &bluezObj, nil
}

func (b *bluezConn) createObject(path dbus.ObjectPath, data base.ObjectMap) Base {
	// The ifaceMap is interface: data. We choose the first interface with data.
	var newObj Base
	for iface := range data {
		ctor, ok := typeRegistry[iface]
		//if len(ifaceData) > 0 && ok {
		if ok {
			// no need to pass in the main interface type since that's keyed
			// by the constructor
			newObj = ctor(b, path, data)
			break
		}
	}
	return newObj
}

func (b *bluezConn) AddWatch(
	path dbus.ObjectPath,
	signalMap []InterfaceSignalPair) (ObjectChangedChan, error) {

	for _, pair := range signalMap {
		err := b.ops.Watch(bus.RootPath, pair.Interface, pair.SignalName)
		if err != nil {
			return nil, err
		}
	}

	ch := make(ObjectChangedChan, 2)
	logger.Info("AddWatch %s", path)
	b.sigWatchersMux.Lock()
	defer b.sigWatchersMux.Unlock()
	b.signalWatchers[path] = ch

	return ch, nil
}

func (b *bluezConn) RemoveWatch(
	path dbus.ObjectPath,
	ch ObjectChangedChan,
	signalMap []InterfaceSignalPair) error {
	b.sigWatchersMux.Lock()
	ch, ok := b.signalWatchers[path]
	if !ok {
		return fmt.Errorf("No channel found for %s", path)
	}
	delete(b.signalWatchers, path)
	close(ch)
	logger.Info("RemoveWatch %s", path)
	b.sigWatchersMux.Unlock()
	for _, pair := range signalMap {
		// We should replace the channel in the map value with a slice, and
		// remove the channel from the slice. When the slice is zero, we
		// call the UnWatch. (It means, we need to store the interfaceName).
		// But, this will do for now.
		err := b.ops.UnWatch(bus.RootPath, pair.Interface, pair.SignalName)
		if err != nil {
			return err
		}
	}
	return nil
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

func (b *bluezConn) IntrospectPath(path string) (*base.Node, error) {
	return b.ops.IntrospectObject(BluezDest, dbus.ObjectPath(path))
}

func (b *bluezConn) GetManagedObjects(path string) (map[dbus.ObjectPath]base.ObjectMap, error) {
	return b.ops.GetManagedObjects(BluezDest, dbus.ObjectPath(path))
}

func (b *bluezConn) GetObjectsByType(oType string) []Base {
	objects := make([]Base, 0)
	b.registryMux.RLock()
	defer b.registryMux.RLock()
	for _, v := range b.objectRegistry {
		if v.GetBluezInterface() == oType {
			objects = append(objects, v)
		}
	}
	fmt.Printf("Found %d objects of type %s\n", len(objects), oType)
	return objects
}

func (b *bluezConn) GetObjectsByInterface(interfaceName string) []Base {
	results := make([]Base, 0)
	b.registryMux.RLock()
	defer b.registryMux.RLock()
	for _, obj := range b.objectRegistry {
		ifaces := obj.GetInterfaces()
		for _, i := range ifaces {
			if i == interfaceName {
				results = append(results, obj)
				break
			}
		}
	}

	return results
}

func (b *bluezConn) FindObjects(pattern string, firstOnly bool) []Base {
	// If pattern is a regex, then it can match more than one
	if len(pattern) == 0 {
		fmt.Println("Pattern is empty")
		return nil
	}
	end := string(pattern[len(pattern)-1])
	withoutEnd := string(pattern[:len(pattern)-1])
	results := make([]Base, 0)
	b.registryMux.RLock()
	defer b.registryMux.RLock()
	for path, obj := range b.objectRegistry {
		if end == "*" && strings.HasPrefix(string(path), withoutEnd) {
			results = append(results, obj)
		} else {
			if string(path) == pattern {
				results = append(results, obj)
				break
			}
		}
	}

	return results
}

// while it's just a slice of interfaces, it seems like the data is a slice of:
// - the dbus.ObjectdPath,
// - map of interfaces (string) to properties (map of property names to dbus.Variant), which we've
//   typed to bus.ObjectMap
// So, we'll go with this assumption, and throw and error if it's not, and record when it fails
// our expectations and deal with them later.
func (b *bluezConn) parseSignalBody(signalBody []interface{}) (dbus.ObjectPath, base.ObjectMap, error) {
	var path dbus.ObjectPath
	var props base.ObjectMap
	for _, i := range signalBody {
		p, ok := i.(dbus.ObjectPath)
		if ok {
			path = p
			continue
		}
		m, ok := i.(map[string]map[string]dbus.Variant)
		if ok {
			props = m
			continue
		}
		// If there are more than 2 items in the body, mark it as error so we can figure out what it is
		return "", nil, fmt.Errorf("signal.Body contained unexpected type: %s [%s]", reflect.TypeOf(i), i)
	}

	return path, props, nil
}

// While this reads channel that is passed to dbus, any other gooutine can pass the
// same data.
func (b *bluezConn) handleSignals(ctx context.Context) {
	cancelled := false
	for !cancelled {
		select {
		case sigData := <-b.busSignalCh:
			if base.DumpData {
				_, err := test.MarshalSignal(sigData)
				if err != nil {
					logger.Info("Unable to marshal signal: %s", err)
				}
			}
			path, data, err := b.parseSignalBody(sigData.Body)
			if err != nil {
				logger.Info("Signal Body unhandled: %s: %s", err, sigData.Body)
				continue
			}
			b.registryMux.RLock()
			obj, ok := b.objectRegistry[path]
			b.registryMux.RUnlock()
			if ok {
				//logger.Info("Updating path %s", path)
				obj.Update(data)
			}
			// Doesn't exist. Create the new object
			//logger.Info("Creating new path %s", path)
			obj = b.createObject(path, data)
			b.registryMux.RLock()
			b.objectRegistry[path] = obj
			b.registryMux.RUnlock()
			// Now, forward this to anyone listening on this path.
			sent := false
			b.sigWatchersMux.RLock()
			for p, listener := range b.signalWatchers {
				if listener == nil {
					logger.Info("nil listener")
					continue
				}
				// the Added and Removed are for the adapter, so all of them
				// will get sent do the channel.
				trimmed := strings.TrimPrefix(sigData.Name, bus.ObjectManager+".")
				if trimmed == bus.ObjectManagerFuncs.InterfacesAdded ||
					trimmed == bus.ObjectManagerFuncs.InterfacesRemoved {
					if len(strings.Split(string(p), "/")) == 4 { // /org/bluez/hci0
						listener <- newObjectChangedData(path, obj, sigData.Name)
						sent = true
					}
				} else if path == p {
					fmt.Println("Non newObjectChangedData channel data")
				}
			}
			b.sigWatchersMux.RUnlock()
			if !sent {
				logger.Info("No listeners %s sending %s to %s", sigData.Sender,
					path, sigData.Name)
				continue
			}
		case <-ctx.Done():
			// the conn will close the channel, so don't do this
			//close(b.busSignalCh)
			logger.Info("quitting handler")
			cancelled = true
		}
	}
}
