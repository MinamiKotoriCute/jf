package tcp

import "context"

type Delivery struct {
}

func NewDelivery() *Delivery {
	return &Delivery{}
}

func (o *Delivery) Start(ctx context.Context) error {
	return nil
}

func (o *Delivery) Stop(ctx context.Context) error {
	return nil
}
