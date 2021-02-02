package protocol

// const strings for this package

const (
	// BluezDest is the destination required for all(?) D-Bus calls
	BluezDest = "org.bluez"
	// BluezRootPath is the root of the bluez objects
	BluezRootPath = "/org/bluez"

	// InterfacesAddedSignalKey is the signal when an interface is added (device. could be an adapter)
	InterfacesAddedSignalKey = "InterfacesAdded"
	// InterfacesRemovedSignalKey is the signal when an interface is removed.
	InterfacesRemovedSignalKey = "InterfacesRemoved"
)

type (
	// InterfaceSignalPair is the paramter type (as a slicle) for the Add/RemoveWatch functions
	InterfaceSignalPair struct {
		// Interface is the interface name providing the signal
		Interface string
		// SignalName is the signal we want to watch
		SignalName string
	}

	// These are internal classes to make it look like scoped constants. They are accessed
	// as BluezInterface and BluezAdapter
	bluezInterface struct {
		Adapter            string
		Device             string
		AgentManager       string
		MediaTransport     string
		GATTService        string
		GATTCharacteristic string
		GATTDescriptor     string
	}

	bluezAdapter struct {
		StartDiscovery string
		StopDiscovery  string
		Connect        string
		AddressProp    string
		AliasProp      string
	}

	bluezDevice struct {
		Connect              string
		Disconnect           string
		ConnectProfile       string
		DisconnectProfile    string
		Pair                 string
		AddressProp          string
		AddressTypeProp      string
		BlockedProp          string
		ConnectedProp        string
		UUIDsProp            string
		AdapterProp          string
		ServiceDataProp      string
		AliasProp            string
		PairedProp           string
		TrustedProp          string
		LegacyPairingProp    string
		RSSIProp             string
		ServicesResolvedProp string
	}

	bluezGATTService struct {
		UUIDProp     string
		PrimaryProp  string
		IncludesProp string
		HandleProp   string
	}

	bluezGATTCharacteristic struct {
		ReadValue   string
		WriteValue  string
		StartNotify string
		StopNotify  string
	}
	bluezGATTDescriptor struct {
		ReadValue  string
		WriteValue string
	}
)

var (
	// BluezInterface is the constants for the base interface
	BluezInterface = bluezInterface{
		Adapter:            BluezDest + ".Adapter1",
		Device:             BluezDest + ".Device1",
		AgentManager:       BluezDest + ".AgentManager1",
		MediaTransport:     BluezDest + ".MediaTransport1",
		GATTService:        BluezDest + ".GattService1",
		GATTCharacteristic: BluezDest + ".GattCharacteristic1",
		GATTDescriptor:     BluezDest + ".GattDescriptor1",
	}

	// BluezAdapter are the constants for the adapter
	BluezAdapter = bluezAdapter{
		StartDiscovery: BluezInterface.Adapter + ".StartDiscovery",
		StopDiscovery:  BluezInterface.Adapter + ".StopDiscovery",
		Connect:        BluezInterface.Adapter + ".Connect",
		// Address:        BluezInterface.Adapter + ".Address",
		// Alias:          BluezInterface.Adapter + ".Alias",
		AddressProp: "Address",
		AliasProp:   "Alias",
	}

	// BluezDevice are the constants in the BluezInterface.Device interface
	BluezDevice = bluezDevice{
		Connect:              BluezInterface.Device + ".Connect",
		Disconnect:           BluezInterface.Device + ".Disconnect",
		ConnectProfile:       BluezInterface.Device + ".ConnectProfile",
		DisconnectProfile:    BluezInterface.Device + ".DisconnectProfile",
		Pair:                 BluezInterface.Device + ".Pair",
		AddressProp:          "Address",
		AddressTypeProp:      "AddressType",
		BlockedProp:          "Blocked",
		ConnectedProp:        "Connected",
		UUIDsProp:            "UUIDs",
		AdapterProp:          "Adapter",
		ServiceDataProp:      "ServiceData",
		AliasProp:            "Alias",
		PairedProp:           "Paired",
		TrustedProp:          "Trusted",
		LegacyPairingProp:    "LegacyPairing",
		RSSIProp:             "RSSI",
		ServicesResolvedProp: "ServicesResolved",
	}

	// BluezGATTService are the constants for the GATT service
	BluezGATTService = bluezGATTService{
		UUIDProp:     "UUID",
		PrimaryProp:  "Primary",
		IncludesProp: "Includes",
		HandleProp:   "Handle",
	}

	// BluezGATTCharacteristic are the constants for the GATT characteristic
	BluezGATTCharacteristic = bluezGATTCharacteristic{
		ReadValue:   BluezInterface.GATTCharacteristic + ".ReadValue",
		WriteValue:  BluezInterface.GATTCharacteristic + ".WriteValue",
		StartNotify: BluezInterface.GATTCharacteristic + ".StartNotify",
		StopNotify:  BluezInterface.GATTCharacteristic + ".StopNotify",
	}

	// BluezGATTDescriptor are the constants for the GATT descriptor
	BluezGATTDescriptor = bluezGATTDescriptor{
		ReadValue:  "ReadValue",
		WriteValue: "WriteValue",
	}
)
