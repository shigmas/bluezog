package base

import (
	"github.com/godbus/dbus/v5"
)

var (
	// DumpData should be set to true to capture sample data for testing.
	DumpData = false
)

// DBbus is an XML protocol. These are the XML types
type (
	// Arg is a function argument
	Arg struct {
		Name      string `xml:"name,attr" json:"name,attr"`
		Type      string `xml:"type,attr" json:"type,attr"`
		Direction string `xml:"direction,attr" json:"direction,attr"`
	}
	// Method is an method available on an interface
	Method struct {
		Name string `xml:"name,attr" json:"name,attr"`
		Args []Arg  `xml:"arg" json:"arg"`
	}

	// Interface is a descriptino of the Methods, Signals, ans Properties available on a remote object
	Interface struct {
		Name    string   `xml:"name,attr" json:"name,attr"`
		Methods []Method `xml:"method" json:"method"`
	}

	// Node can represent an interface and a set of nodes underneath this node in the hierarchy.
	Node struct {
		Name       string      `xml:"name,attr" json:"name,attr"`
		Interfaces []Interface `xml:"interface" json:"interface"`
		Nodes      []Node      `xml:"node" json:"node"`
	}

	// ObjectMap is represents an object. In GetManagedObjects, it's keyed by ObjectPath
	ObjectMap map[string]map[string]dbus.Variant

	// Operations is an interface for mocking. This also allowed the dbus.Conn to be a member
	// in the real implementation, putting all dbus calls in one struct. ObjectPath, dbus.Signal
	// leak out, but those are not opaque.
	Operations interface {
		IntrospectObject(dest string, objPath dbus.ObjectPath) (*Node, error)
		GetObjectProperty(dest string, objPath dbus.ObjectPath, propName string) (interface{}, error)
		GetManagedObjects(dest string, objPath dbus.ObjectPath) (map[dbus.ObjectPath]ObjectMap, error)
		CallFunction(dest string, objPath dbus.ObjectPath, funcName string) error
		CallFunctionWithArgs(
			retVal interface{},
			dest string,
			objPath dbus.ObjectPath,
			funcName string,
			args ...interface{}) error

		RegisterSignalChannel(ch chan<- *dbus.Signal)
		Watch(path dbus.ObjectPath, iface string, method string) error
		UnWatch(path dbus.ObjectPath, iface string, method string) error
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
