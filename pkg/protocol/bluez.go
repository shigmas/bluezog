package protocol

import (
	"context"
	"fmt"
	"path"
	"reflect"
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
		GetObjectsByType(oType string) []Base
		GetObjectsByInterface(interfaceName string) []Base
		FindObjects(pattern string, firstOnly bool) []Base

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

func newObjectChangedData(path dbus.ObjectPath, obj Base, sigName string) ObjectChangedData {
	return ObjectChangedData{
		Path:   path,
		Object: obj,
		Signal: sigName,
	}
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
		newObj := bluezObj.createObject(path, ifaceMap)
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

func (b *bluezConn) createObject(path dbus.ObjectPath, data bus.ObjectMap) Base {
	// The ifaceMap is interface: data. We choose the first interface with data.
	var newObj Base
	for iface := range data {
		ctor, ok := typeRegistry[iface]
		//if len(ifaceData) > 0 && ok {
		if ok {
			// no need to pass in the main interface type since that's keyed
			// by the constructor
			logger.Info("Creating object at path %s", path)
			newObj = ctor(b, path, data)
			break
		}
	}
	return newObj
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
		// We should replace the channel in the map value with a slice, and
		// remove the channel from the slice. When the slice is zero, we
		// call the UnWatch. (It means, we need to store the interfaceName).
		// But, this will do for now.
		err := bus.UnWatch(b.busConn, bus.RootPath, interfaceName, s)
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

func (b *bluezConn) GetObjectsByInterface(interfaceName string) []Base {
	results := make([]Base, 0)
	for _, obj := range b.objectRegistry {
		if strings.HasPrefix(interfaceName, BluezDest) {
			if obj.GetBluezInterface() == interfaceName {
				results = append(results, obj)
			}
		} else {
			ifaces := obj.GetInterfaces()
			for _, i := range ifaces {
				if i == interfaceName {
					results = append(results, obj)
					break
				}
			}
		}
	}

	return results
}

func (b *bluezConn) FindObjects(pattern string, firstOnly bool) []Base {
	// If pattern is a regex, then it can match more than one
	end := string(pattern[len(pattern)-1])
	withoutEnd := string(pattern[:len(pattern)-1])
	results := make([]Base, 0)
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
func (b *bluezConn) parseSignalBody(signalBody []interface{}) (dbus.ObjectPath, bus.ObjectMap, error) {
	var path dbus.ObjectPath
	var props bus.ObjectMap
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
		return "", nil, fmt.Errorf("signal.Body contained unexpected type: %s", reflect.TypeOf(i))
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
			logger.Info("Received: %s", sigData.Path)
			path, data, err := b.parseSignalBody(sigData.Body)
			if err != nil {
				logger.Info("Signal Body unhandled: %s", err)
				continue
			}
			obj, ok := b.objectRegistry[path]
			if ok {
				logger.Info("Updating path %s", path)
				obj.Update(data)
			}
			// Doesn't exist. Create the new object
			logger.Info("Creating new path %s", path)
			obj = b.createObject(path, data)
			b.objectRegistry[path] = obj

			// Now, forward this to anyone listening on this path.
			// XXX - Since we allow patterns (well, just *), we need to iterate, which isn't so scalable.
			sent := false
			for p, listener := range b.signalWatchers {
				end := string(p[len(p)-1])
				withoutEnd := string(p[:len(p)-1])
				if (end == "*" && strings.HasPrefix(string(path), withoutEnd)) || (p == path) {
					listener <- newObjectChangedData(path, obj, sigData.Name)
					logger.Info("Sent change on %s", path)
					sent = true
				}
			}
			if !sent {
				logger.Info("No listeners for %s", path)
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
