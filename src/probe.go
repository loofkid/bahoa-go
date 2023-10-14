package main

import (
	"fmt"
)

type Probe struct {
	address uint8
	id *string
	label *string
	temperature float32
	connected bool
}

func NewProbe (address uint8) *Probe {
	return &Probe{
		address: address,
		id: nil,
		label: nil,
		temperature: 500,
		connected: false,
	}
}

func (p *Probe) ToString() string {
	return fmt.Sprintf("Address: %X, ID: %s, Label: %s, Temperature: %f, Connected: %t", p.address, *p.id, *p.label, p.temperature, p.connected)
}