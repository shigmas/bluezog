module github.com/shigmas/bluezog

go 1.14

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/godbus/dbus/v5 v5.0.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
)

replace (
	github.com/godbus/dbus/v5 v5.0.3 => ../../godbus/dbus
)