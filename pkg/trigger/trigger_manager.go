package trigger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MinamiKotoriCute/jf/pkg/helper"
)

type TriggerManager struct {
	log      *slog.Logger
	triggers []*Trigger
}

var _ helper.Service = (*TriggerManager)(nil)

func NewTriggerManager(log *slog.Logger, triggers ...*Trigger) *TriggerManager {
	if log == nil {
		log = slog.Default()
	}

	return &TriggerManager{
		log:      log,
		triggers: triggers,
	}
}

func (o *TriggerManager) Start(ctx context.Context) error {
	for _, trigger := range o.triggers {
		o.log.InfoContext(ctx, fmt.Sprintf("start trigger %s", trigger.Name()))
		if err := trigger.Start(); err != nil {
			return err
		}
		o.log.InfoContext(ctx, fmt.Sprintf("start trigger end %s", trigger.Name()))
	}
	return nil
}

func (o *TriggerManager) Stop(ctx context.Context) error {
	for _, trigger := range o.triggers {
		o.log.InfoContext(ctx, fmt.Sprintf("stop trigger %s", trigger.Name()))
		trigger.Stop()
		o.log.InfoContext(ctx, fmt.Sprintf("stop trigger end %s", trigger.Name()))
	}
	return nil
}
