package dbusproxy

import (
	"context"
	"fmt"
	"log"
	"time"
)

func ExampleProxy_Call() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	proxy := New(BusTypeSession, "org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus")
	query := "org.freedesktop.DBus"
	var nameOwner string
	err := proxy.Call(ctx, "GetNameOwner", Args{query}, &nameOwner)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	fmt.Println(nameOwner)
	// Output: org.freedesktop.DBus
}
