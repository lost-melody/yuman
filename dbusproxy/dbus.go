// Package dbusproxy proxies DBus method calls.
package dbusproxy

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type (
	Args = []any
)

var (
	MsgErrConnectDBus = func(busType string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrConnectDBus",
				Other: "connecting {{.BusType}} bus: %w",
			},
			TemplateData: map[string]any{
				"BusType": busType,
			},
		}
	}
	MsgErrCallDBusMethod = func(busType, methodName string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrCallDBusMethod",
				Other: "calling {{.BusType}} bus method '{{.Method}}': %w",
			},
			TemplateData: map[string]any{
				"BusType": busType,
				"Method":  methodName,
			},
		}
	}
)

// CallSession connects to the session bus and call a method.
//
//   - param "args" is a slice of arguments, or nil if not needed.
//   - param "retValues" is a slice of pointers to return values, or nil if not expected.
func CallSession(ctx context.Context,
	destination string, objectPath dbus.ObjectPath,
	interfaceName, methodName string,
	args Args, retValues ...any,
) (err error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		err = tr.LocalizeErrorf(MsgErrConnectDBus("session"), err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	method := interfaceName + "." + methodName
	err = conn.
		Object(destination, objectPath).
		CallWithContext(ctx, method, 0, args...).
		Store(retValues...)
	if err != nil {
		err = tr.LocalizeErrorf(MsgErrCallDBusMethod("session", method), err)
		return
	}

	return
}

// CallSystem connects to the system bus and call a method.
//
// See also [CallSession].
func CallSystem(ctx context.Context,
	destination string, objectPath dbus.ObjectPath,
	interfaceName, methodName string,
	args Args, retValues ...any,
) (err error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		err = tr.LocalizeErrorf(MsgErrConnectDBus("system"), err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	method := interfaceName + "." + methodName
	err = conn.
		Object(destination, objectPath).
		CallWithContext(ctx, method, 0, args...).
		Store(retValues...)
	if err != nil {
		err = tr.LocalizeErrorf(MsgErrCallDBusMethod("system", method), err)
		return
	}

	return
}
