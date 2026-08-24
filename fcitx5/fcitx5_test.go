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
