package trigger

import (
	"bytes"
	"runtime/debug"
	"sync"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
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
					fields := logrus.Fields{
						"error": serr.ToJSON(err, true),
					}
					if o.name != "" {
						fields["name"] = o.name
					}
					logrus.WithFields(fields).Warning("trigger handle fail")
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
