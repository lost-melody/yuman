package fcitx5

import (
	"context"
	"fmt"
	"log"
	"time"
)

func ExampleControllerImpl_CanRestart() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	canRestart, err := Controller.CanRestart(ctx)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	fmt.Println("CanRestart:", canRestart)
	// Output: CanRestart: true
}

func ExampleControllerImpl_GetAddonConfig() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	config, err := Controller.GetAddonConfig(ctx, "yume")
	if err != nil {
		log.Println("Error:", err)
		return
	}

	fmt.Println(len(config.Schemes))

	// Output: 1
}
