package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"fmt"
	"github.com/d2r2/go-i2c"
	"errors"
)

type ProbesController struct {
	probes []*Probe
	i2cs []*i2c.I2C
}

var validAddresses = []int{0x42, 0x43, 0x44, 0x45}

func NewProbesController() chan *ProbesController {
	r := make(chan *ProbesController)
	go func () {
		var probesController = ProbesController{
			probes: []*Probe{},
			i2cs: []*i2c.I2C{},
		}
		_ = probesController.GetAvailableI2Cs()
		for _, i2c := range probesController.i2cs {
			log.Printf("0x%X,", i2c.GetAddr())
		}
		r <- &probesController
	}()
	return r
}

func (c *ProbesController) GetAvailableI2Cs() int {
	r := make(chan int)
	go func() {
		for _, address := range validAddresses {
			i2cCreateChannel := make(chan int)
			go func () {
				i2c, err := i2c.NewI2C(uint8(address), 1)
				if err != nil {
					i2cCreateChannel <- 1
				}
				// Request online status using 0x01
				req_buffer := make([]byte, 1)
				req_buffer = append(req_buffer, 0x01)
				_, err = i2c.WriteBytes(req_buffer)
				if err != nil {
					i2cCreateChannel <- 1
				}
				// Read online status - 0x02 means online
				online_buffer := make([]byte, 1)
				_, err = i2c.ReadBytes(online_buffer)
				if err != nil {
					i2cCreateChannel <- 1
				}
				if online_buffer[0] != 0x02 {
					i2cCreateChannel <- 1
				}
				log.Printf("0x%X\n", i2c.GetAddr())
				c.i2cs = append(c.i2cs, i2c)
				i2cCreateChannel <- 0
			}()
			<-i2cCreateChannel
		}
		if len(c.i2cs) == 0 {
			c.GetAvailableI2Cs()
		}
		r <- 1
	}()
	return <-r
}

var attempts = 0

func (c *ProbesController) ReadFromI2C(i2c *i2c.I2C) error {
	var err error
	attempts += 1
	address := i2c.GetAddr()
	log.Printf("Reading from address %X\n", address)

	existingProbesOnAddress := []*Probe{}
	for _, probe := range c.probes {
		if (probe.address == address) {
			existingProbesOnAddress = append(existingProbesOnAddress, probe)
		}
	}

	fullSum := 0
	
	startBuffer := make([]byte, 1)
	for startBuffer[0] != 0x69 {
		startBuffer = make([]byte, 1)
		i2c.ReadBytes(startBuffer)
	}

	if startBuffer[0] == 0x69 {
		fullSum += int(startBuffer[0])

		for i := 0; i < 4; i++ {
			buffer := make([]byte, 7)
			_, err := i2c.ReadBytes(buffer)
			if err != nil {
				log.Println(err)
			}
			checksum := 0
			for _, v := range buffer {
				checksum += int(v)
				fullSum += int(v)
			}
			checksum = checksum % 256

			sumBuffer := make([]byte, 1)
			_, err = i2c.ReadBytes(sumBuffer)
			if err != nil {
				log.Println(err)
			}
			fullSum += int(sumBuffer[0])
			if int(sumBuffer[0]) == checksum {
				log.Println("Checksum OK")

				probeNotFound := true
				for _, probe := range existingProbesOnAddress {
					if (*probe.id == string(buffer[0:2])) {
						probeNotFound = false
						log.Printf("Probe %s found on address %X", *probe.id, address)

						// Get temperature from buffer
						var floatValue float32
						floatBuffer := bytes.NewReader(buffer[2:7])
						err = binary.Read(floatBuffer, binary.LittleEndian, &floatValue)
						if err != nil {
							log.Println("binary.Read failed:", err)
						}

						probe.temperature = floatValue

						// Get connection status from buffer
						if (buffer[6] == 0x01) {
							probe.connected = true
						} else {
							probe.connected = false
						}
					}
				}
				if probeNotFound {
					log.Printf("Probe %s not found on address %X", string(buffer[0:2]), address)
					
					// Create new probe
					probe := NewProbe(address)
					probe.id = new(string)
					probe.label = new(string)
					*probe.id = string(buffer[0:2])
					if address == 0x42 {
						if (string(buffer[0:2]) == "A0") {
							*probe.label = "Ambient Probe"
						} else {
							*probe.label = fmt.Sprintf("Probe %d", i)
						}
					} else {
						*probe.label = "Probe " + string(rune(len(c.probes) - 1))
					}

					// Get temperature from buffer
					var floatValue float32
					floatBuffer := bytes.NewReader(buffer[2:7])
					err = binary.Read(floatBuffer, binary.LittleEndian, &floatValue)
					if err != nil {
						log.Println("binary.Read failed:", err)
					}
					probe.temperature = floatValue

					if (buffer[6] == 0x01) {
						probe.connected = true
					} else {
						probe.connected = false
					}
					c.probes = append(c.probes, probe)
				}
			} else {
				log.Println("Checksum ERROR")
			}
		}

		endBuffer := make([]byte, 1)
		_, err := i2c.ReadBytes(endBuffer)
		if err != nil {
			log.Println(err)
		}

		fullSumBuffer := make([]byte, 1)
		_, err = i2c.ReadBytes(fullSumBuffer)
		if err != nil {
			log.Println(err)
		}

		fullSum = fullSum % 256

		if fullSum != int(fullSumBuffer[0]) && attempts < 5 {
			log.Println("Full sum ERROR")
			c.ReadFromI2C(i2c)
		} else {
			return err
		}
	} else {
		err = errors.New("no new data")
	}
	attempts = 0
	return err
}

func (c *ProbesController) ReadFromAllI2Cs() chan int {
	if len(c.i2cs) > 0 {
		// i2cAddresses := []string{}
		// for _, i2c := range c.i2cs {
		// 	i2cAddresses = append(i2cAddresses, fmt.Sprintf("%X", i2c.GetAddr()))
		// }
		// log.Println("i2c Addresses", i2cAddresses)
		r := make(chan int)
		go func() {
			for _, i2c := range c.i2cs {
				c.ReadFromI2C(i2c)
			}
			r <- 1
		}()
		return r
	}
	return nil
}

func (c *ProbesController) Close() {
	for _, i2c := range c.i2cs {
		i2c.Close()
	}
}