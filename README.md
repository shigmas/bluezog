An attempt at a bluez-like interface to access bluetooth devices over DBus in Go.

There are already a few out there, but they are too full featured. Features are nice, but not when the framework
has bugs which are impossible to work around. Bluez is a simple interface, so it shouldn't be bogged down with
too much extra stuff if you just want to a way to access devices in Go.

So, basically, this is just wraps the Golang DBus API (godbus) with some bluetooth-ish commands

There is one command that is meant to be something like bluetoothctl.

