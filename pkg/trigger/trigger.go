package trigger

import (
	"sync"
	"time"

	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
)

type HandleFunc func() error

type Trigger struct {
	done     chan struct{}
	wg       sync.WaitGroup
	duration time.Duration
	name     string
	handle   HandleFunc
}

func NewTrigger(duration time.Duration,
	name string,
	handle HandleFunc) *Trigger {
	return &Trigger{
		duration: duration,
		name:     name,
		handle:   handle,
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
			select {
			case <-o.done:
				return
			case <-ticker.C:
				if err := o.handle(); err != nil {
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
