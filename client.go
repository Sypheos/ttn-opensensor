package opensensor

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"io"
	"net/http"
	"time"
)

const id string = "ttnctl"
const logName = "Opensensor"
const OpenSensorURI string = "https://realtime.opensensors.io/v1/topics/"
const uriClientId string = "client-id"
const uriPassword string = "password"

//The Thing Network mqtt parameters
type TtnAccess struct {
	AppId, Key, Broker, DeviceId string
}

//OpenSensor mqtt parameters
type OpenSensorAccess struct {
	ClientId, Pw, Key, Topic string
}

type openSensorData struct {
	Data []byte `json:"data"`
}

type OpenSensor struct {
	ctx          log.Interface
	mqtt         mqtt.Client
	sensorAccess OpenSensorAccess
	ttnAccess    TtnAccess
}

func NewOpenSensor(ttnAccess TtnAccess, sensorAccess OpenSensorAccess) *OpenSensor {
	c := apex.Stdout().WithField(logName, fmt.Sprint(ttnAccess.AppId, ":", ttnAccess.DeviceId,
		" with ", sensorAccess.Topic))
	return &OpenSensor{ctx: c, mqtt: mqtt.NewClient(c, id, ttnAccess.AppId, ttnAccess.Key, ttnAccess.Broker),
		ttnAccess: ttnAccess, sensorAccess: sensorAccess}
}

func (o *OpenSensor) Start(ttnAccess TtnAccess, sensorAccess OpenSensorAccess) {

	if err := o.mqtt.Connect(); err != nil {
		o.ctx.WithError(err).Fatal("Could not connect")
	}
	token := o.mqtt.SubscribeDeviceUplink(ttnAccess.AppId, ttnAccess.DeviceId,
		func(client mqtt.Client, appID string, devID string, req types.UplinkMessage) {
			o.uplink(req.PayloadRaw, sensorAccess)
		})
	token.Wait()
	if err := token.Error(); err != nil {
		o.ctx.WithError(err).Fatal("Could not subscribe")
	}
}

func (o *OpenSensor) uplink(payload []byte, access OpenSensorAccess) {

	b, err := encode(payload)
	if err != nil {
		o.ctx.WithError(err).Fatal(fmt.Sprintln("Could not encode payload", payload))
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	resp, err := (&http.Client{Transport: tr, Timeout: time.Second * 10}).Post(o.sensorAccess.Topic, "application/json", b)
	if err != nil {
		o.ctx.WithError(err).Fatal("Could not reash Opensor http endpoint")
		return
	}
	if resp.StatusCode != 200 {
		o.ctx.Error(resp.Status)
	}
}

func encode(payload []byte) (io.Reader, error) {
	data := openSensorData{payload}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return nil, err
	}
	return io.Reader(b), nil
}
