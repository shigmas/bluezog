package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/godbus/dbus/v5"

	"github.com/shigmas/bluezog/pkg/base"
	"github.com/shigmas/bluezog/pkg/logger"
)

func writeBytes(b []byte, prefix string) (string, error) {
	f, err := ioutil.TempFile("./testdata", prefix)
	if err != nil {
		return "", err
	}

	_, err = f.Write(b)
	return f.Name(), err
}

func readBytes(data interface{}, n string) error {
	path := filepath.Join("../..", "testdata", n)
	fmt.Println("Opening ", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}

// MarshalRaw writes the raw bytes and returns the file name or error
func MarshalRaw(b []byte, prefix string) (string, error) {
	return writeBytes(b, fmt.Sprintf("raw-%s-", prefix))
}

// MarshalIntrospect writes the introspect data and returns the file name or error
func MarshalIntrospect(n *base.Node) (string, error) {
	introBytes, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	return writeBytes(introBytes, "introspect-")
}

// UnmarshalIntrospect reads the introspect data and returns the data or error
func UnmarshalIntrospect(fname string) (base.Node, error) {
	var introspectData base.Node
	err := readBytes(&introspectData, fname)
	return introspectData, err
}

// MarshalManagedObjects writes the managed object data and returns the file name or error
func MarshalManagedObjects(s map[dbus.ObjectPath]base.ObjectMap) (string, error) {
	mged, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return writeBytes(mged, "managed-")
}

// UnmarshalManagedObjects reads the managed object data and returns the object data or error
func UnmarshalManagedObjects(fname string) (map[dbus.ObjectPath]base.ObjectMap, error) {
	var s map[dbus.ObjectPath]base.ObjectMap
	err := readBytes(&s, fname)
	return s, err
}

// MarshalSignal writes the signal data and returns the file name or error
func MarshalSignal(signal *dbus.Signal) (string, error) {
	sigBytes, err := json.Marshal(signal)
	if err != nil {
		return "", err
	}

	return writeBytes(sigBytes, "signal-")
}

// UnmarshalSignal reads the signal data and returns the signal or error
func UnmarshalSignal(fname string) (*dbus.Signal, error) {
	var signal dbus.Signal
	err := readBytes(&signal, fname)
	ifaceArray := make([]interface{}, len(signal.Body))
	for index, val := range signal.Body {
		if str, ok := val.(string); ok {
			ifaceArray[index] = dbus.ObjectPath(str)
		} else if ifaceMap, ok := val.(map[string]interface{}); ok {
			trueMap := make(map[string]map[string]dbus.Variant)
			// at this point, all the data is just empty interfaces. That might be
			// the result of shallow marshalling, but, maybe that is as far as we can test
			for k, v := range ifaceMap {
				if fakeVarMap, ok := v.(map[string]interface{}); ok {
					// It's a real value, so we need to drill down even more
					realVarMap := make(map[string]dbus.Variant)
					for subk, hiddenVar := range fakeVarMap {
						realVarMap[subk] = dbus.MakeVariant(hiddenVar)
					}
					trueMap[k] = realVarMap
				} else {
					logger.Error("Unable unmarshal signal body, sub sub map")
					trueMap[k] = map[string]dbus.Variant{}
				}

			}
			ifaceArray[index] = trueMap
		} else {
			fmt.Printf("Unhandled Unmarshal signal data: %s\n", val)
		}
	}
	signal.Body = ifaceArray

	return &signal, err
}
