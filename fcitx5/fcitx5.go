// Package fcitx5 proxies Fcitx5 Controller DBus service.
package fcitx5

import (
	"context"

	"github.com/godbus/dbus/v5"
	"github.com/lost-melody/yuman/dbusproxy"
)

const (
	ServiceDestination   string          = "org.fcitx.Fcitx5"
	ControllerObjectPath dbus.ObjectPath = "/controller"
	ControllerInterface  string          = "org.fcitx.Fcitx.Controller1"
)

type ControllerImpl struct {
	dbusproxy.Proxy
}

var Controller = ControllerImpl{
	Proxy: dbusproxy.New(dbusproxy.BusTypeSession, ServiceDestination, ControllerObjectPath, ControllerInterface),
}

func (controller *ControllerImpl) Activate(ctx context.Context) (err error) {
	err = controller.Call(ctx, "Activate", nil)
	return
}

func (controller *ControllerImpl) Deactivate(ctx context.Context) (err error) {
	err = controller.Call(ctx, "Deactivate", nil)
	return
}

func (controller *ControllerImpl) Exit(ctx context.Context) (err error) {
	err = controller.Call(ctx, "Exit", nil)
	return
}

func (controller *ControllerImpl) CanRestart(ctx context.Context) (canRestart bool, err error) {
	err = controller.Call(ctx, "CanRestart", nil, &canRestart)
	return
}

func (controller *ControllerImpl) Restart(ctx context.Context) (err error) {
	err = controller.Call(ctx, "Restart", nil)
	return
}
