package bus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectManager(t *testing.T) {
	assert.Equal(t, ObjectManager, "org.freedesktop.DBus.ObjectManager", "they should be equal")
	assert.Equal(t, ObjectManagerFuncs.GetManagedObjects, "org.freedesktop.DBus.ObjectManager.GetManagedObjects",
		"they should be equal")
}
