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
	"net/url"
	"time"
)

const id string = "ttnctl"
const logName = "Opensensor"
const openSensorURI = "https://realtime.opensensors.io/v1/topics/"

//TtnAccess structure representing The Thing Network mqtt parameters
type TtnAccess struct {
	AppID, Key, Broker, DeviceID string
}

//SensorAccess OpenSensor.io http endpoint parameters http parameters
type SensorAccess struct {
	ClientID, Pw, Key, Topic string
}

type openSensorData struct {
	Data interface{} `json:"data"`
}

//OpenSensor integration structure definition
type OpenSensor struct {
	ctx          log.Interface
	mqtt         mqtt.Client
	sensorAccess SensorAccess
	ttnAccess    TtnAccess
	httpTopic    *url.URL
}

//NewOpenSensor create a new OpenSensor integration for one device
func NewOpenSensor(ttnAccess TtnAccess, sensorAccess SensorAccess) (*OpenSensor, error) {
	c := apex.Stdout().WithField(logName, fmt.Sprint(ttnAccess.AppID, ":", ttnAccess.DeviceID,
		" with ", sensorAccess.Topic))
	u, err := prepareURL(sensorAccess)
	if err != nil {
		return nil, err
	}
	return &OpenSensor{ctx: c, mqtt: mqtt.NewClient(c, id, ttnAccess.AppID, ttnAccess.Key, ttnAccess.Broker),
		ttnAccess: ttnAccess, sensorAccess: sensorAccess, httpTopic: u}, nil
}

//Start integration. Will fatal if connection or mqtt subscription is impossible.
func (o *OpenSensor) Start() {

	if err := o.mqtt.Connect(); err != nil {
		o.ctx.WithError(err).Fatal("Could not connect")
	}
	token := o.mqtt.SubscribeDeviceUplink(o.ttnAccess.AppID, o.ttnAccess.DeviceID,
		func(client mqtt.Client, appID string, devID string, req types.UplinkMessage) {
			o.uplink(req.PayloadRaw)
		})
	token.Wait()
	if err := token.Error(); err != nil {
		o.ctx.WithError(err).Fatal("Could not subscribe")
	}
}

//Stop integration. Will fatal if it cannot properly close the connection
func (o *OpenSensor) Stop() {

	token := o.mqtt.UnsubscribeDeviceUplink(o.ttnAccess.AppID, o.ttnAccess.DeviceID)
	token.Wait()
	if err := token.Error(); err != nil {
		o.ctx.WithError(err).Fatal("Could not unsubcribe from devive uplink")
	}
	o.mqtt.Disconnect()
}

func (o *OpenSensor) uplink(payload []byte) {

	b, err := encode(payload)
	if err != nil {
		o.ctx.WithError(err).Fatal(fmt.Sprintln("Could not encode payload", payload))
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	req, err := o.prepareReq("POST", b)
	if err != nil {
		o.ctx.WithError(err).Fatal("Could not reach create request for http endpoint")
		return
	}
	resp, err := (&http.Client{Transport: tr, Timeout: time.Second * 10}).Do(req)
	if err != nil {
		o.ctx.WithError(err).Fatal("Could not reach Opensor http endpoint")
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		o.ctx.Error(resp.Status)
	}
}

func prepareURL(sensor SensorAccess) (*url.URL, error) {

	u, err := url.Parse(openSensorURI + url.QueryEscape(sensor.Topic))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Add("client-id", sensor.ClientID)
	q.Add("password", sensor.Pw)
	u.RawQuery = q.Encode()
	return u, nil
}

func (o *OpenSensor) prepareReq(method string, b io.Reader) (*http.Request, error) {

	req, err := http.NewRequest(method, o.httpTopic.String(), b)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "api-key "+o.sensorAccess.Key)
	return req, nil
}

func encode(payload []byte) (io.Reader, error) {
	data := openSensorData{string(payload)}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		return nil, err
	}
	return io.Reader(b), nil
}
