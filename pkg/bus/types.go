package bus

const (
	// ObjectManager is the Interface provided by dbus
	ObjectManager = "org.freedesktop.DBus.ObjectManager"
	// Properties is the interface for something that has properties
	Properties = "org.freedesktop.DBus.Properties"
	// Introspectable is implemented by most objects
	Introspectable = "org.freedesktop.DBus.Introspectable"
	// RootPath is the object path of the root
	RootPath = "/"
)

type (

	// ObjectManagerFuncs are functions available. They will be declared through a public
	// package level instance so they're accessed syntactically like constants.
	objectManagerFuncs struct {
		// const names
		GetManagedObjects string
		// Actually, signals. Should be renamed
		InterfacesAdded   string
		InterfacesRemoved string
	}

	propertiesFuncs struct {
		// Actually, signals. Should be renamed
		PropertiesChanged string
	}
	introspectableFuncs struct {
		Introspect string
	}
)

var (
	// ObjectManagerFuncs are the functions provided by ObjectManager.
	ObjectManagerFuncs = objectManagerFuncs{
		GetManagedObjects: ObjectManager + ".GetManagedObjects",
		// InterfacesAdded:   ObjectManager + ".InterfacesAdded",
		// InterfacesRemoved: ObjectManager + ".InterfacesRemoved",
		// Signals are
		InterfacesAdded:   "InterfacesAdded",
		InterfacesRemoved: "InterfacesRemoved",
	}

	// PropertiesFuncs are the signals provided by Properties
	PropertiesFuncs = propertiesFuncs{
		PropertiesChanged: Properties + ".PropertiesChanged",
	}
	// IntrospectableFuncs are the functdions provided on Introspectable
	IntrospectableFuncs = introspectableFuncs{
		Introspect: Introspectable + ".Introspect",
	}
)
