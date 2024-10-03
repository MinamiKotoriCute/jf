package trigger

import (
	"bytes"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/MinamiKotoriCute/serr"
)

type HandlerFunc func() error

type Trigger struct {
	done     chan struct{}
	wg       sync.WaitGroup
	duration time.Duration
	name     string
	handler  HandlerFunc
}

func NewTrigger(duration time.Duration,
	name string,
	handler HandlerFunc) *Trigger {
	return &Trigger{
		duration: duration,
		name:     name,
		handler:  handler,
	}
}

func (o *Trigger) Start() error {
	o.wg.Add(1)
	o.done = make(chan struct{})

	go func() {
		defer o.wg.Done()

		ticker := time.NewTicker(o.duration)
		defer ticker.Stop()

		for {
			if o.done == nil {
				return
			}

			select {
			case <-o.done:
				return
			case <-ticker.C:
				if err := o.runHandlerAndCapturePanic(); err != nil {
					attrs := []interface{}{
						slog.Any("err", serr.ToJSON(err, true)),
					}
					if o.name != "" {
						attrs = append(attrs, slog.String("name", o.name))
					}
					slog.Warn("trigger handle fail", attrs...)
				}
			}
		}
	}()

	return nil
}

func (o *Trigger) runHandlerAndCapturePanic() (err error) {
	defer func() {
		if v := recover(); v != nil {
			stack := debug.Stack()
			goroutines, _ := gostackparse.Parse(bytes.NewReader(stack))

			err = serr.Errors(map[string]interface{}{
				"goroutines": goroutines,
				"value":      v,
			}, "panic")
		}
	}()

	err = o.handler()
	return
}

func (o *Trigger) Stop() {
	if o.done == nil {
		return
	}

	close(o.done)
	o.done = nil
	o.wg.Wait()
}

func (o *Trigger) Name() string {
	return o.name
}
