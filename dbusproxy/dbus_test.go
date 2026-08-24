package dbusproxy

import (
	"context"
	"fmt"
	"log"
	"time"
)

func ExampleCallSession() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	query := "org.freedesktop.DBus"
	var nameOwner string
	err := CallSession(ctx, "org.freedesktop.DBus", "/org/freedesktop/DBus", "org.freedesktop.DBus", "GetNameOwner", Args{query}, &nameOwner)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	fmt.Println(nameOwner)
	// Output: org.freedesktop.DBus
}
