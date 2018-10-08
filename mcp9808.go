// Copyright 2018 NeuralSpaz All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mcp9808 provides a API for using a I²C(i2c) temperature sensor MCP9008.
package mcp9808

import (
	"encoding/binary"
	"errors"
	"log"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
)

type MCP9808 struct {
	address uint16
	d       conn.Conn
	mfgID   uint16
	id      uint16

	Temperature float64
}

const (
	MCP9808_I2CADDR_DEFAULT  byte = 0x18
	MCP9808_REG_CONFIG       byte = 0x01
	MCP9808_REG_AMBIENT_TEMP byte = 0x05
	MCP9808_REG_MANUF_ID     byte = 0x06
	MCP9808_REG_DEVICE_ID    byte = 0x07
	MCP9808_REG_RESOLUTION   byte = 0x08
)

func New(b i2c.Bus, opts ...func(*MCP9808) error) (*MCP9808, error) {
	mcp := new(MCP9808)

	mcp.address = uint16(MCP9808_I2CADDR_DEFAULT)

	for _, option := range opts {
		option(mcp)
	}

	var err error
	mcp.d = &i2c.Dev{Bus: b, Addr: mcp.address}

	if err != nil {
		log.Panic(err)
	}

	if err := mcp.init(); err != nil {
		return nil, err
	}

	return mcp, nil
}

// Address sets the i2c address in not using the default address of 0x40
func Address(address uint16) func(*MCP9808) error {
	return func(m *MCP9808) error {
		return m.setAddress(address)
	}
}

func (m *MCP9808) setAddress(address uint16) error {
	m.address = address
	return nil
}

func (m *MCP9808) writeReg(register byte, w []byte) error {
	buf := append([]byte{register}, w...)
	return m.d.Tx(buf, nil)
}

func (m *MCP9808) readReg(register byte, r []byte) error {
	return m.d.Tx([]byte{register}, r)
}

func (m *MCP9808) init() error {

	rx := make([]byte, 2)
	if err := m.readReg(MCP9808_REG_MANUF_ID, rx); err != nil {
		return err
	}
	m.mfgID = binary.BigEndian.Uint16(rx)
	if m.mfgID != 0x0054 {
		return errors.New("part does not match driver")
	}

	if err := m.readReg(MCP9808_REG_DEVICE_ID, rx); err != nil {
		return err
	}
	m.id = binary.BigEndian.Uint16(rx)
	if m.id != 0x0400 {
		return errors.New("part does not match driver")
	}

	if err := m.writeReg(MCP9808_REG_CONFIG, []byte{0x00, 0x00}); err != nil {
		return err
	}

	if err := m.writeReg(MCP9808_REG_RESOLUTION, []byte{0x03}); err != nil {
		return err
	}
	return nil
}

func (m *MCP9808) Temp() (float64, error) {

	rx := make([]byte, 2)
	if err := m.readReg(MCP9808_REG_AMBIENT_TEMP, rx); err != nil {
		return 0, err
	}
	traw := binary.BigEndian.Uint16(rx)
	traw &= 0x1FFF
	if traw&0x1000 == 0x1000 {
		traw &= 0x0FFF
		return -float64(traw) * 0.0625, nil // °C per bit
	}
	return float64(traw) * 0.0625, nil // °C per bit

}
