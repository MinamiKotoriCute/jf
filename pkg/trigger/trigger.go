package trigger

import (
	"sync"
	"time"

	"github.com/MinamiKotoriCute/serr"
	"github.com/sirupsen/logrus"
)

type HandleFunc func() error

type Trigger struct {
	done chan struct{}
	wg   sync.WaitGroup
}

func (o *Trigger) Start(d time.Duration, name string, handle HandleFunc) error {
	o.wg.Add(1)
	o.done = make(chan struct{})

	go func() {
		defer o.wg.Done()

		ticker := time.NewTicker(d)
		defer ticker.Stop()

		for {
			select {
			case <-o.done:
				return
			case <-ticker.C:
				if err := handle(); err != nil {
					fields := logrus.Fields{
						"error": serr.ToJSON(err, true),
					}
					if name != "" {
						fields["name"] = name
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
