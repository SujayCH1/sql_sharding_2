package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EventEmitter struct {
	ctx context.Context
}

func NewEventEmitter(ctx context.Context) *EventEmitter {
	return &EventEmitter{ctx: ctx}
}

func (e *EventEmitter) Emit(name string, payload any) {
	runtime.EventsEmit(e.ctx, name, payload)
}
