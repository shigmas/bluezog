An attempt at a bluez-like interface to access bluetooth devices over DBus in Go.

There are already a few out there, but they are too full featured. Features are nice, but not when the framework
has bugs which are impossible to work around. Bluez is a simple interface, so it shouldn't be bogged down with
too much extra stuff if you just want to a way to access devices in Go.

So, basically, this is just wraps the Golang DBus API (godbus) with some bluetooth-ish commands

There is one command that is meant to be something like bluetoothctl.

* Dependencies
 - godbus: A nice golang interface to D-Bus
 - testify: I like this for testing
 - readline: This is for the command line interface. Ideally, it would be a selective dependency, or the CLI tool could be a separate module. But, this can really make module fetching messy.
 
* Testing notes:
 - > device /org/bluez/hci0/dev_FF_F2_DF_D8_10_D4 connect
   This works, but it seems like it's not getting the alert when it is initially found. But it's in the cache. This is one of my ble beacons. No UUID shows up.
 - cached devices are in /var/lib/bluetooth, under the adapter. 
 
Omron USB ?:
/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4
service:
/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service000e
characteristic:
/org/bluez/hci0/dev_FF_F2_DF_D8_10_D4/service000e/char0019

Omron Bag:
Path: /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C
LegacyPairing: %!s(bool=false)
Connected: %!s(bool=false)
Address: D1:40:FD:DE:C6:1C
Alias: EnvSensor-BL01
Blocked: %!s(bool=false)
Adapter: /org/bluez/hci0
AddressType: random
Paired: %!s(bool=false)
Trusted: %!s(bool=false)
UUIDs: [00001800-0000-1000-8000-00805f9b34fb 00001801-0000-1000-8000-00805f9b34fb 0000180a-0000-1000-8000-00805f9b34fb 0c4c3000-7700-46f4-aa96-d5e974e32a54 0c4c3010-7700-46f4-aa96-d5e974e32a54 0c4c3030-7700-46f4-aa96-d5e974e32a54 0c4c3040-7700-46f4-aa96-d5e974e32a54]
ServicesResolved: %!s(bool=false)
Name: EnvSensor-BL01

Usage:
 * GATT: Depending on the hardware, which might be by the name of the object, or the Manufacturer Data, you need the UUID of the GATT Service, Characteristic, or Descriptor. The FindObjects method on protocol.Bluez is perhaps the primary way to get the device. You will cast it to a protocol.Device.

commands:
list path org.bluez.Device1
object /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C dump
list path org.bluez.GattCharacteristic1
# This one seems to be valid (but crashed because the in param was nil?)
object /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C connect 0c4c3000-7700-46f4-aa96-d5e974e32a54
# calling this without connect will crash
gatt /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C/service0026/char002d

adapter
start
object /org/bluez/hci0/dev_D1_40_FD_DE_C6_1C connect 0c4c3000-7700-46f4-aa96-d5e974e32a54
stop

I think the ALPS is:
 SNM00 ( /org/bluez/hci0/dev_48_F0_7B_78_45_5E)

The UUID isn't in the main advertising data, so you can connect to it without a profile
object /org/bluez/hci0/dev_48_F0_7B_78_45_5E connect

Then the GATT objects will be added
