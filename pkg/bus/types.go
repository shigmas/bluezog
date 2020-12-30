package bus

import "github.com/godbus/dbus/v5"

// DBbus is an XML protocol. These are the XML types
type (
	// Arg is a function argument
	Arg struct {
		Name      string `xml:"name,attr"`
		Type      string `xml:"type,attr"`
		Direction string `xml:"direction,attr"`
	}
	// Method is an method available on an interface
	Method struct {
		Name string `xml:"name,attr"`
		Args []Arg  `xml:"arg"`
	}

	// Interface is a descriptino of the Methods, Signals, ans Properties available on a remote object
	Interface struct {
		Name    string   `xml:"name,attr"`
		Methods []Method `xml:"method"`
	}

	// Node can represent an interface and a set of nodes underneath this node in the hierarchy.
	Node struct {
		Name       string      `xml:"name,attr"`
		Interfaces []Interface `xml:"interface"`
		Nodes      []Node      `xml:"node"`
	}

	// ObjectMap is represents an object. In GetManagedObjects, it's keyed by ObjectPath
	ObjectMap map[string]map[string]dbus.Variant
)

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
		InterfacesAdded   string
		InterfacesRemoved string
	}

	propertiesFuncs struct {
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

	PropertiesFuncs = propertiesFuncs{
		PropertiesChanged: Properties + ".PropertiesChanged",
	}
	// IntrospectableFuncs are the functdions provided on Introspectable
	IntrospectableFuncs = introspectableFuncs{
		Introspect: Introspectable + ".Introspect",
	}
)

// Obviously, not types, but they provide the Stringify interface
func (a *Arg) String() string {
	s := "Arg: " + a.Name
	s += ", Type: " + a.Type
	s += ", Direction: " + a.Direction
	s += "\n"
	return s
}

func (m *Method) String() string {
	s := "Node: " + m.Name
	for _, a := range m.Args {
		s += a.String()
	}
	s += "\n"
	return s
}

func (i *Interface) String() string {
	s := "\nInterface: " + i.Name
	// for _, m := range i.Methods {
	// 	s += m.String()
	// }
	s += "\n"
	return s
}

func (n *Node) String() string {
	s := "Node: " + n.Name + "\n"
	for _, n := range n.Nodes {
		s += n.String()
	}
	for _, i := range n.Interfaces {
		s += i.String()
	}
	s += "\n"
	return s
}
