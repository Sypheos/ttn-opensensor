package main

import (
	"github.com/apex/log"
	"opensensor"
)

func main() {
	sensor := opensensor.SensorAccess{"5885", "yEUIsPrx", "5b46e6e9-d572-49a7-bce8-bfcf5362550c",
		"/users/sypheos/home/firstfloor/temperature"}
	ttna := opensensor.TtnAccess{"open-sensor", "ttn-account-v2.CLWM-c78CsFxUUZPfXCe9933kdVHdV1nIzrNk-kApP8",
		"tcp://eu.thethings.network:1883", "heater"}
	o, err := opensensor.NewOpenSensor(ttna, sensor)
	if err != nil {
		log.WithError(err).Fatal("couldn't start integration")
	}
	o.Start()
	for true {

	}
}
