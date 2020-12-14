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
	// These are internal classes to make it look like scoped constants. They are accessed
	// as BluezInterface and BluezAdapter
	bluezInterface struct {
		Adapter        string
		Device         string
		AgentManager   string
		MediaTransport string
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
		ConnectProfile       string
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
)

var (
	// BluezInterface is the constants for the base interface
	BluezInterface = bluezInterface{
		Adapter:        BluezDest + ".Adapter1",
		Device:         BluezDest + ".Device1",
		AgentManager:   BluezDest + ".AgentManager1",
		MediaTransport: BluezDest + ".MediaTransport1",
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
		ConnectProfile:       BluezInterface.Device + ".ConnectProfile",
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
)