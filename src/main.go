package main

import (
	"syscall"

	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"

	// "github.com/ambelovsky/gosf"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

const interruptPin = rpi.J8p29

var probesController *ProbesController

func handleInterrupt(evt gpiod.LineEvent) {
	fmt.Println("Interrupt!")
	// log.Println(probesController)
	// if probesController == nil {
	// 	probesController = NewProbesController()
	// }
	if evt.Type == gpiod.LineEventFallingEdge && probesController != nil && len(probesController.i2cs) > 0 {
		val := probesController.ReadFromAllI2Cs()
		<-val
		// for _, probe := range probesController.probes {
		// 	fmt.Println(probe.ToString())
		// }
	}
}
// func init() {
// 	gosf.Listen("probes", )
// }

func forever() {

	// gosf.Startup(map[string]interface{}{"port": 8080, "path": "/socket"})

}

func resetMcu() {
	reset, err := gpiod.RequestLine("gpiochip0", rpi.J8p32, gpiod.WithPullUp, gpiod.AsOutput(1))
	if err != nil {
		log.Fatal(err)
	}
	defer reset.Close()

	reset.SetValue(0)
	time.Sleep(2 * time.Millisecond)
	reset.SetValue(1)
}

func main() {

	l, err := gpiod.RequestLine("gpiochip0", interruptPin, gpiod.WithPullUp, gpiod.WithFallingEdge, gpiod.WithEventHandler(handleInterrupt))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	probesController = <- NewProbesController()

	go forever()

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	probesController.Close()
	fmt.Println("Exiting")
}