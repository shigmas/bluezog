package test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/shigmas/bluezog/pkg/base"
)

type (
	// busMock mocks the bus operations
	busMock struct {
		sigCh       chan<- *dbus.Signal
		sigStopper  map[string]func()
		managedType string
	}
)

var (
	// BusSignalInterval is the time between simulated signals
	BusSignalInterval                 = time.Second
	_                 base.Operations = (*busMock)(nil)
)

// NewBusMock creates a mock bus (base.Operations implementation). The
// parameter is either "simple" or "gatt" for the managed objects file
func NewBusMock(managedType string) base.Operations {
	return &busMock{
		managedType: managedType,
	}
}

// IntrospectObject fetches the XML for Introspection and parses it into a Node hierarchy
func (b *busMock) IntrospectObject(dest string, objPath dbus.ObjectPath) (*base.Node, error) {
	node, err := UnmarshalIntrospect("introspect-794476729")
	return &node, err
}

// GetObjectProperty for the specified object and property name
func (b *busMock) GetObjectProperty(dest string, objPath dbus.ObjectPath, propName string) (interface{}, error) {

	return nil, fmt.Errorf("GetObjectProperty not yet mocked")
}

// GetManagedObjects retrieves the paths of the objects managed by this object
func (b *busMock) GetManagedObjects(dest string, objPath dbus.ObjectPath) (map[dbus.ObjectPath]base.ObjectMap, error) {

	return UnmarshalManagedObjects("managed-" + b.managedType)
}

// CallFunction is exposes the simplest and common way to call a function on the object
func (b *busMock) CallFunction(
	_ context.Context,
	dest string,
	objPath dbus.ObjectPath,
	funcName string) error {
	if strings.HasSuffix(funcName, "StartDiscovery") {
		return nil
	}
	if strings.HasSuffix(funcName, "StopDiscovery") {
		return nil
	}
	return fmt.Errorf("CallFunction(%s) not yet mocked", funcName)
}

// CallFunctionWithArgs is simply CallFunction with arbitrary arguments
func (b *busMock) CallFunctionWithArgs(
	_ context.Context,
	retVal interface{},
	dest string,
	objPath dbus.ObjectPath,
	funcName string,
	args ...interface{}) error {
	return fmt.Errorf("CallFunctionWithArgs not yet mocked")
}

// RegisterSignalChannel passes the signal to DBus.
func (b *busMock) RegisterSignalChannel(ch chan<- *dbus.Signal) {
	b.sigCh = ch
	if b.sigStopper == nil {
		b.sigStopper = make(map[string]func())
	}
}

func getWatchKey(iface, method string) string {
	return iface + "-" + method
}

// Watch is a simplified version of AddMatchsignal
func (b *busMock) Watch(path dbus.ObjectPath, iface string, method string) error {
	if method != "InterfacesAdded" {
		fmt.Printf("No stored signals for %s. Nothing will be found\n", method)
		return nil
	}
	watchKey := getWatchKey(iface, method)
	if _, found := b.sigStopper[watchKey]; found {
		// already watching this. don't do anything
		fmt.Println("Already watching ", watchKey)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.sigStopper[watchKey] = cancel
	// We only handle the one signal right now
	go func(ctx context.Context) {
		sigPaths := []string{"signal-InterfacesAdded-741522808", "signal-InterfacesAdded-859333239", "signal-InterfacesAdded-929583933", "signal-InterfacesAdded-945160042"}
		ticker := time.NewTicker(BusSignalInterval)
		index := 0
		for {
			select {
			case <-ctx.Done():
				break
			case <-ticker.C:
				sig, err := UnmarshalSignal(sigPaths[index])
				if err != nil {
					fmt.Println("Error in Unmarshaling signal to send")
				}
				b.sigCh <- sig
				index++
				if index == len(sigPaths) {
					index = 0
				}
			}
		}
	}(ctx)

	return nil
}

// UnWatch is a simplified version of RemoveMatchsignal
func (b *busMock) UnWatch(path dbus.ObjectPath, iface string, method string) error {
	cancel, ok := b.sigStopper[getWatchKey(iface, method)]
	if !ok {
		fmt.Println("No watcher for ", getWatchKey(iface, method))
		return nil
	}
	fmt.Println("stopping watch for ", getWatchKey(iface, method))
	cancel()
	delete(b.sigStopper, getWatchKey(iface, method))
	return nil
}
