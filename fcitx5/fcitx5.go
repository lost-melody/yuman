// Package fcitx5 proxies Fcitx5 Controller DBus service.
package fcitx5

import (
	"context"
	"fmt"

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

// Config stores '(va(sa(sssva{sv})))'.
type Config struct {
	Options map[string]any  `json:"options"`
	Schemes []*ConfigScheme `json:"schemes"`
}

// ConfigScheme stores '(sa(sssva{sv}))'.
type ConfigScheme struct {
	Name    string          `json:"name"`
	Options []*ConfigOption `json:"options"`
}

// ConfigOption stores '(sssva{sv})'.
type ConfigOption struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Title   string         `json:"title"`
	Default any            `json:"default"`
	Extras  map[string]any `json:"extras"`
}

// AddonInfo stores '(sssibb)'.
type AddonInfo struct {
	UniqueName   string `json:"unique_name"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Priority     int32  `json:"priority"`
	Enabled      bool   `json:"enabled"`
	Configurable bool   `json:"configurable"`
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

func (controller *ControllerImpl) GetAddons(ctx context.Context) (addons []*AddonInfo, err error) {
	err = controller.Call(ctx, "GetAddons", nil, &addons)
	return
}

func (controller *ControllerImpl) GetConfig(ctx context.Context, uri string) (config *Config, err error) {
	config = &Config{}
	err = controller.Call(ctx, "GetConfig", dbusproxy.Args{uri}, &config.Options, &config.Schemes)
	return
}

func (controller *ControllerImpl) GetGlobalConfig(ctx context.Context) (config *Config, err error) {
	config, err = controller.GetConfig(ctx, "fcitx://config/global")
	return
}

func (controller *ControllerImpl) GetAddonConfig(ctx context.Context, addon string) (config *Config, err error) {
	config, err = controller.GetConfig(ctx, fmt.Sprintf("fcitx://config/addon/%s", addon))
	return
}
