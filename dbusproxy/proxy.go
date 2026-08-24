package dbusproxy

import (
	"context"

	"github.com/godbus/dbus/v5"
)

type BusType int8

const (
	BusTypeUnknown BusType = iota
	BusTypeSession
	BusTypeSystem
)

type Proxy struct {
	BusType       BusType
	Destination   string
	ObjectPath    dbus.ObjectPath
	InterfaceName string
}

func New(busType BusType, destination string, objectPath dbus.ObjectPath, interfaceName string) Proxy {
	switch busType {
	case BusTypeSession, BusTypeSystem:
	default:
		panic("unreachable")
	}
	return Proxy{
		BusType:       busType,
		Destination:   destination,
		ObjectPath:    objectPath,
		InterfaceName: interfaceName,
	}
}

func (proxy *Proxy) Call(ctx context.Context, methodName string, args Args, retValues ...any) (err error) {
	switch proxy.BusType {
	case BusTypeSession:
		err = CallSession(ctx, proxy.Destination, proxy.ObjectPath, proxy.InterfaceName, methodName, args, retValues...)
	case BusTypeSystem:
		err = CallSystem(ctx, proxy.Destination, proxy.ObjectPath, proxy.InterfaceName, methodName, args, retValues...)
	default:
		panic("unreachable")
	}
	if err != nil {
		return
	}

	return
}
