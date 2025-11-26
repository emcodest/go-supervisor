package supervisor

import (
	"context"
	"log"
	"math"
	"time"
)

/*
Package supervisor provides a simple, production-ready supervisor pattern
for running goroutines that automatically restart when they panic.

The supervisor isolates your worker code and ensures that crashes do not
kill your process. If a worker panics or returns, the supervisor catches
the panic, logs it, waits using exponential backoff, and restarts it.

Usage example:

    supervisor.Start(ctx, supervisor.DefaultConfig(), func(ctx context.Context) {
        for {
            select {
            case <-ctx.Done():
                return
            default:
            }

            // do work...

            if somethingBad {
                panic("boom")
            }
        }
    })

To stop all supervised workers, cancel the context you passed into Start().
*/

type Config struct {
	MinBackoff time.Duration
	MaxBackoff time.Duration
	Logger     *log.Logger
}

func DefaultConfig() Config {
	return Config{
		MinBackoff: 1 * time.Second,
		MaxBackoff: 30 * time.Second,
		Logger:     log.Default(),
	}
}

// Start launches a supervised worker that auto-restarts on panic.
//
// The worker function must block (usually via an infinite loop) and exit
// only when ctx.Done() is closed. If the worker panics, the supervisor
// catches the panic and restarts it using exponential backoff.
func Start(ctx context.Context, cfg Config, worker func(ctx context.Context)) {
	if cfg.MinBackoff == 0 {
		cfg.MinBackoff = 1 * time.Second
	}
	if cfg.MaxBackoff == 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}

	go func() {
		backoff := cfg.MinBackoff

		for {
			select {
			case <-ctx.Done():
				cfg.Logger.Println("[supervisor] stopped")
				return
			default:
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						cfg.Logger.Printf("[supervisor] worker crashed: %v", r)
					}
				}()

				worker(ctx)
			}()

			cfg.Logger.Printf("[supervisor] restarting worker in %v", backoff)
			time.Sleep(backoff)

			// exponential backoff
			backoff = time.Duration(math.Min(float64(backoff*2), float64(cfg.MaxBackoff)))
		}
	}()
}
