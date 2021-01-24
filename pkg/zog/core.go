package zog

// This is the implementation of the command line functions. So, fmt.Print* is fine, since it's
// meant to be interactive, with the interactive part controlled by the code in cmd.
import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/logger"
	"github.com/shigmas/bluezog/pkg/protocol"
)

type (
	// Bus is the API for the Bluez DBus
	Bus interface {
		// GetInterface starts the search for adapters, setting the default adapter
		GetInterface(...interface{}) error
		// StartDiscovery starts device discovery
		StartDiscovery(...interface{}) error
		// StopDiscovery stops discovery
		StopDiscovery(...interface{}) error
		// DeviceCommands provides API to the device
		ObjectCommands(...interface{}) error
		// Gatt handles GATT api requests
		Gatt(...interface{}) error
		// Close the connection to the bus
		Close(...interface{}) error
		// List objects. Can pass a property that we're looking for. Only objects that have that
		// property will be listed, with that property
		List(...interface{}) error
		// Filter is kind of like List, but takes more arguments?
		Filter(...interface{}) error
		// Test
		Test(...interface{}) error
	}

	// BusImpl is the implementation of Bus. Exposed for... fun?
	BusImpl struct {
		bluez          protocol.Bluez
		defaultAdapter *protocol.Adapter
		cancelFunc     func()
		deviceRecvCh   protocol.ObjectChangedChan
		rwMux          sync.RWMutex
	}
	// BusFunc declares the command interface to the shell
	BusFunc func(Bus, ...interface{}) error
)

var (
	// BusCommand is the map for commands to the functions.
	BusCommand map[string]BusFunc = make(map[string]BusFunc)
)

func init() {
	// These should be command sets. And some commands in the sets will set the command set.
	BusCommand["close"] = (Bus).Close
	BusCommand["adapter"] = (Bus).GetInterface
	BusCommand["start"] = (Bus).StartDiscovery
	BusCommand["stop"] = (Bus).StopDiscovery
	BusCommand["object"] = (Bus).ObjectCommands
	BusCommand["list"] = (Bus).List
	BusCommand["filter"] = (Bus).Filter
	BusCommand["gatt"] = (Bus).Gatt
	BusCommand["test"] = (Bus).Test
}

// NewBus creates a new bus
func NewBus(ctx context.Context, ops base.Operations) Bus {
	fmt.Println("Initializing Bluez")
	//base.DumpData = true
	bluez, err := protocol.InitializeBluez(ctx, ops)

	if err != nil {
		return nil
	}

	b := BusImpl{
		bluez:        bluez,
		deviceRecvCh: make(protocol.ObjectChangedChan, 3),
	}

	return &b
}

// GetInterface searches the bus for adapters
func (b *BusImpl) GetInterface(...interface{}) error {
	adapters := b.bluez.FindAdapters()
	if len(adapters) == 0 {
		return fmt.Errorf("Failed to FindAdapters")
	}

	for _, a := range adapters {
		if b.defaultAdapter == nil {
			b.defaultAdapter = a
		}
		addr, err := a.FetchProperty(protocol.BluezAdapter.AddressProp)
		if err != nil {
			return fmt.Errorf("Error fetching %s", protocol.BluezAdapter.AddressProp)
		}
		logger.Info("Address: %s", addr)
	}
	return nil
}

// StartDiscovery : The order would be:
// 1. get the adapter
// 2. start discovery
// 3. stop (when done)
func (b *BusImpl) StartDiscovery(...interface{}) error {
	go b.deviceReceiver()
	var err error
	b.rwMux.Lock()
	b.deviceRecvCh, err = b.defaultAdapter.StartDiscovery()
	b.rwMux.Unlock()
	if err != nil {
		return err
	}
	return nil
}

// StopDiscovery closes the access to the devices on the default adapter
func (b *BusImpl) StopDiscovery(...interface{}) error {
	b.defaultAdapter.StopDiscovery()
	return nil
}

func printNode(n *base.Node) string {
	str := fmt.Sprintf("%s\n", n.Name)
	str += fmt.Sprintf("Sub Nodes:\n")
	for _, sub := range n.Nodes {
		str += fmt.Sprintf("\t%s\n", sub.Name)
	}
	str += fmt.Sprintf("Interfaces:\n")
	for _, sub := range n.Interfaces {
		str += fmt.Sprintf("\t%s\n", sub.Name)
		for _, m := range sub.Methods {
			str += fmt.Sprintf("\t\tMethod: %s(%s)\n", m.Name, m.Args)
		}
		for _, s := range sub.Signals {
			str += fmt.Sprintf("\t\tSignal: %s(%s)\n", s.Name, s.Args)
		}
	}

	return str
}

// ObjectCommands provide the API to send commands to devices
func (b *BusImpl) ObjectCommands(args ...interface{}) error {
	// device /org/bluez/hci0/dev_FE_CD_66_43_D8_9E connect 00001800-0000-1000-8000-00805f9b34fb 00001801-0000-1000-8000-00805f9b34fb
	if len(args) < 2 {
		return fmt.Errorf("ConnectToDevice needs an address and a command, and any additional arguments. only %d args", len(args))
	}
	addressArg, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("Unable to convert %s to string", args[0])
	}
	command, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("Unable to convert %s to string", args[2])
	}

	objs := b.bluez.FindObjects(addressArg, true)
	if len(objs) == 0 {
		return fmt.Errorf("No devices in registry with address %s", addressArg)
	}
	base := objs[0]
	device, ok := base.(*protocol.Device)
	if !ok && (command == "connect" || command == "disconnect") {
		return fmt.Errorf("Base is not a Device")
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	switch command {
	case "dump":
		fmt.Printf("Path: %s\n", base.GetPath())
		props := base.AllProperties()
		for k, variant := range props {
			fmt.Printf("%s: %s\n", k, variant.Value())
		}
	case "introspect":
		node, err := b.bluez.IntrospectPath(addressArg)
		if err != nil {
			return err
		}
		fmt.Printf(printNode(node))
	case "children":
		managed, err := b.bluez.GetManagedObjects(addressArg)
		if err != nil {
			return err
		}
		fmt.Printf("children: %s\n", managed)

	case "connect":
		// device /org/bluez/hci0/dev_D7_57_C6_C2_0B_FA connect
		var err error
		if len(args) == 3 {
			uuid, ok := args[2].(string)
			if !ok {
				return fmt.Errorf("Unable to convert %s to string", args[2])
			}
			err = device.ConnectProfile(ctx, uuid)
			fmt.Println("ConnectProfile")
		} else {
			err = device.Connect(ctx)
		}
		if err != nil {
			return fmt.Errorf("Unable to connect to device %s: %s", addressArg, err)
		}
	case "disconnect":
		// device /org/bluez/hci0/dev_D7_57_C6_C2_0B_FA disconnect
		var err error
		if len(args) == 3 {
			uuid, ok := args[2].(string)
			if !ok {
				return fmt.Errorf("Unable to convert %s to string", args[2])
			}
			err = device.DisconnectProfile(ctx, uuid)
			fmt.Println("ConnectProfile")
		} else {
			err = device.Disconnect(ctx)
		}
		if err != nil {
			return fmt.Errorf("Unable to connect to device %s: %s", addressArg, err)
		}
	case "property":
		if len(args) != 3 {
			return fmt.Errorf("property needs the property name as an argument")
		}
		propName, ok := args[2].(string)
		if !ok {
			return fmt.Errorf("Unable to %s to string", args[2])
		}
		prop, err := device.FetchProperty(propName)
		if err != nil {
			return fmt.Errorf("Failed to get property [%s]: %s", propName, err)
		}
		fmt.Printf("Property %s has value %s\n", propName, prop)
	}

	return nil
}

// Gatt provides access to the GATT functionality
func (b *BusImpl) Gatt(args ...interface{}) error {
	if len(args) != 1 {
		return fmt.Errorf("gatt needs an address")
	}
	addressArg, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("Unable to convert %s to string", args[0])
	}
	parts := strings.Split(addressArg, "/")
	objs := b.bluez.FindObjects(addressArg, true)
	if len(objs) == 0 {
		return fmt.Errorf("No devices in registry with address %s", addressArg)
	}
	base := objs[0]

	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	defer cancel()

	// A GATT paths
	// Service:
	// /org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f
	// Characteristic:
	// /org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f/char0022
	// and a Descriptor:
	// /org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service001f/char0022/desc0024
	if len(parts) == 6 { // the empty space before the / is included
		service, ok := base.(*protocol.GattService)
		if !ok {
			return fmt.Errorf("Path appeared to a GATT Service, but not convertible")
		}
		fmt.Println(service)
	} else if len(parts) == 7 {
		characteristic, ok := base.(*protocol.GattCharacteristic)
		if !ok {
			return fmt.Errorf("Path appeared to a GATT Characteristic, but not convertible")
		}
		val, err := characteristic.ReadValue(ctx, 0)
		if err != nil {
			return err
		}
		fmt.Printf("Char: %s\n", val)
	} else if len(parts) == 8 {
		return fmt.Errorf("Path appeared to a GATT Descriptor, but not implemented")
	} else {
		return fmt.Errorf("Not a GATT path")
	}

	return nil
}

// List objects by interface, and, optionally, if they have the specified property.
func (b *BusImpl) List(args ...interface{}) error {
	var ok bool
	if len(args) < 2 {
		return fmt.Errorf("[path|all] <interface name> (property))")
	}

	onlyPath := false
	if pathMode, ok := args[0].(string); ok {
		if pathMode == "path" {
			onlyPath = true
		}
	}

	name := ""
	if name, ok = args[1].(string); !ok {
		return fmt.Errorf("interface or property argument was not a string")
	}

	propName := ""
	if len(args) >= 3 {
		if propName, ok = args[2].(string); !ok {
			return fmt.Errorf("interface or property argument was not a string")
		}
	}

	objects := b.bluez.GetObjectsByInterface(name)

	fmt.Printf("Found %d objects\n", len(objects))
	if propName != "" {
		fmt.Printf("Filtering for only objects containing %s\n", propName)
	}

	for _, o := range objects {
		if propName != "" {
			prop, err := o.FetchProperty(propName)
			if err != nil || prop == nil {
				continue
			}
		}
		fmt.Printf("Path: %s\n", o.GetPath())
		if propName != "" {
			propVal := o.Property(propName)
			fmt.Printf("%s: %s\n", propName, propVal)
		} else if !onlyPath {
			props := o.AllProperties()
			for k, variant := range props {
				fmt.Printf("%s: %s\n", k, variant.Value())
			}
		}
	}
	return nil
}

// Filter doesn't do anything
func (b *BusImpl) Filter(args ...interface{}) error {
	return nil
}

// Close the connection. Or not
func (b *BusImpl) Close(...interface{}) error {
	//b.conn.Close()
	return nil
}

// Test tests
func (b *BusImpl) Test(args ...interface{}) error {
	numArgs := 0
	for a := range args {
		fmt.Println("arg: ", args[a])
		numArgs++
	}
	fmt.Println("Num args: ", numArgs)
	return nil
}

func (b *BusImpl) deviceReceiver() {
	fmt.Printf("Waiting for devices...\n")
	b.rwMux.RLock()
	for d := range b.deviceRecvCh {
		objs := b.bluez.FindObjects(string(d.Path), true)
		mfgData := "default"
		if len(objs) >= 0 {
			base := objs[0]
			device, ok := base.(*protocol.Device)
			if !ok {
				mfgData = "not device"
			} else {
				prop, err := device.FetchProperty("ManufacturerData")
				if err == nil {
					var ok bool
					mfgDataMap, ok := prop.(map[uint16]interface{})
					if ok {
						mfgData = fmt.Sprintf("%d: %v", len(mfgDataMap), mfgDataMap)
					} else {
						// This may be empty for devices that don't provide this property
						mfgData = reflect.TypeOf(prop).Name()
					}
				} else {
					fmt.Println("Error: ", err)
					mfgData = err.Error()
				}
			}
		} else {
			mfgData = "Not in Cache"
		}

		fmt.Printf("Received %s (ManufacturerData: %s)\n", d.Path, mfgData)
	}
	b.rwMux.RUnlock()
	fmt.Printf("Leaving deviceReceiver\n")
}
