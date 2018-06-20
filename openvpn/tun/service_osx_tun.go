/*
 * Copyright (C) 2018 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package tun

import (
	"fmt"
	"os/exec"

	"bytes"
	log "github.com/cihub/seelog"
	"os"
)

type serviceDarwinTun struct {
	device Device
}

func (service *serviceDarwinTun) Add(device Device) {
	service.device = device
}

func (service *serviceDarwinTun) Start() error {
	return service.createTunDevice()
}

func (service *serviceDarwinTun) Stop() {
	var err error
	var exists bool

	log.Info(tunLogPrefix, "checking device exists")
	log.Flush()

	if exists, err = service.deviceExists(); err != nil {
		log.Info(tunLogPrefix, err)
		log.Flush()
	}

	if !exists {
		return
	}

	log.Info(tunLogPrefix, "deleting device")
	if err = service.deleteDevice(); err != nil {
		log.Info(tunLogPrefix, err)
		log.Flush()
	}

	log.Info(tunLogPrefix, "device deleted")
	log.Flush()
}

func (service *serviceDarwinTun) createTunDevice() (err error) {
	exists, err := service.deviceExists()
	if err != nil {
		return err
	}

	if exists {
		log.Info(tunLogPrefix, service.device.Name+" device already exists, attempting to use it")
		return nil
	}

	var stderr bytes.Buffer
	cmd := exec.Command(
		"sh",
		"-c",
		"sudo ip tuntap add dev "+service.device.Name+" mode tun",
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to add tun device: %s: %s", err, stderr.String())
	}

	log.Info(tunLogPrefix, service.device.Name+" device created")
	return nil
}

func (service *serviceDarwinTun) deviceExists() (exists bool, err error) {
	if _, err := os.Stat("/sys/class/net/" + service.device.Name); err == nil {
		return true, nil
	}

	return false, err
}

func (service *serviceDarwinTun) deleteDevice() (err error) {
	var stderr bytes.Buffer
	cmd := exec.Command(
		"sh",
		"-c",
		"sudo openvpn --rmtun --dev "+service.device.Name,
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Info(tunLogPrefix, service.device.Name, stderr.String())
		log.Flush()
		return fmt.Errorf("Failed to remove tun device: %s", err)
	}

	log.Info(tunLogPrefix, service.device.Name, " device removed")
	log.Flush()
	return nil
}
