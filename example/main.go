package main

import (
	"context"
	"log"
	"time"

	"github.com/emcodest/go-supervisor"
)

func main() {
	ctx := context.Background()

	supervisor.Start(ctx, supervisor.DefaultConfig(), func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Println("worker shutting down")
				return
			default:
			}

			log.Println("worker running...")
			time.Sleep(1 * time.Second)

			// simulate crash
			if time.Now().Unix()%5 == 0 {
				panic("simulated failure")
			}
		}
	})

	select {} // keep main alive
}
