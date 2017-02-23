package main

import "opensensor"

func main() {
	sensor := opensensor.OpenSensorAccess{"5885", "yEUIsPrx", "5b46e6e9-d572-49a7-bce8-bfcf5362550c",
		"http://localhost:3000"/*"/users/sypheos/home/firstfloor/temperature"*/}
	ttna := opensensor.TtnAccess{"open-sensor", "ttn-account-v2.CLWM-c78CsFxUUZPfXCe9933kdVHdV1nIzrNk-kApP8",
		"tcp://eu.thethings.network:1883", "heater"}
	opensensor.Start(ttna, sensor)
	//opensensor.Uplink(nil, sensor)
	for true {

	}
}
