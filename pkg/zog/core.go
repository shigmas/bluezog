package zog

// This is the implementation of the command line functions. So, fmt.Print* is fine, since it's
// meant to be interactive, with the interactive part controlled by the code in cmd.
import (
	"context"
	"fmt"

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
		DeviceCommands(...interface{}) error
		// Close the connection to the bus
		Close(...interface{}) error

		// Test
		Test(...interface{}) error
	}

	// BusImpl is the implementation of Bus. Exposed for... fun?
	BusImpl struct {
		bluez          protocol.Bluez
		defaultAdapter *protocol.Adapter
		cancelFunc     func()
		devices        map[string]*protocol.Device
		deviceRecvCh   protocol.DeviceReceiverCh
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
	BusCommand["discover"] = (Bus).StartDiscovery
	BusCommand["stop"] = (Bus).StopDiscovery
	BusCommand["device"] = (Bus).DeviceCommands
	BusCommand["test"] = (Bus).Test
}

// NewBus creates a new bus
func NewBus(ctx context.Context) Bus {
	fmt.Println("Initializing Bluez")
	bluez, err := protocol.InitializeBluez(ctx)

	if err != nil {
		return nil
	}

	b := BusImpl{
		bluez:        bluez,
		deviceRecvCh: make(protocol.DeviceReceiverCh, 3),
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
		fmt.Println("A: ", a)
		logger.Info("Local Addr: %s", a.Property(protocol.BluezAdapter.AddressProp))
		addr, err := a.FetchProperty(protocol.BluezAdapter.AddressProp)
		if err != nil {
			return fmt.Errorf("Error fetching %s", protocol.BluezAdapter.AddressProp)
		}
		logger.Info("Address: %s", addr)
	}
	return nil
}

// StartDiscovery : The order would be:
// 3. on device, watch properties
func (b *BusImpl) StartDiscovery(...interface{}) error {
	b.devices = make(map[string]*protocol.Device)
	go b.deviceReceiver()
	b.defaultAdapter.StartDiscovery(b.deviceRecvCh)
	return nil
}

// StopDiscovery closes the access to the devices on the default adapter
func (b *BusImpl) StopDiscovery(...interface{}) error {
	b.defaultAdapter.StopDiscovery()
	return nil
}

// ConnectToDevice connects to the specified device
func (b *BusImpl) DeviceCommands(args ...interface{}) error {
	if len(args) < 3 {
		return fmt.Errorf("ConnectToDevice needs an address and command, and any additional arguments")
	}
	address, ok := args[0].(string)
	if !ok {
		return fmt.Errorf("Unable to convert %s to string", args[0])
	}
	dev, ok := b.devices[address]
	if !ok {
		return fmt.Errorf("No device with path [%s]", address)
	}

	command, ok := args[1].(string)
	if !ok {
		return fmt.Errorf("Unable to convert %s to string", args[1])
	}

	switch command {
	case "connect":
		err := dev.Connect()
		if err != nil {
			return fmt.Errorf("Unable to connect to device %s: %s", address, err)
		}
	case "property":
		if len(args) != 3 {
			return fmt.Errorf("property needs the property name as an argument")
		}
		propName, ok := args[2].(string)
		if !ok {
			return fmt.Errorf("Unable to %s to string", args[2])
		}
		prop, err := dev.FetchProperty(propName)
		if err != nil {
			return fmt.Errorf("Failed to get property [%s]: %s", err)
		}
		fmt.Printf("Property %s has value %s\n", propName, prop)
	}

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
	for d := range b.deviceRecvCh {
		fmt.Printf("Received %s\n", d)
		// b.devices[string(d.Path)] = &d
		// prop, err := d.GetProperty(protocol.BluezDevice.AliasProp)
		// if err == nil {
		// 	fmt.Printf("Received: %s (%s)\n", d.Path, prop)
		// }
	}
}