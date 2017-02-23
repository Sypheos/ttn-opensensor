package opensensor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"time"
)

const id string = "ttnctl"
const logName = "Opensensor"
const OpenSensorURI string = "https://realtime.opensensors.io/v1/topics/"
const uriClientId string = "client-id"
const uriPassword string = "password"

//The Thing Network client parameters
type TtnAccess struct {
	AppId, Key, Broker, DeviceId string
}

//OpenSensor client parameters
type OpenSensorAccess struct {
	ClientId, Pw, Key, Topic string
}

type openSensorData struct {
	Data []byte `json:"data"`
}

type OpenSensor struct {
	client       mqtt.Client
	sensorAccess OpenSensorAccess
	ttnAccess    TtnAccess
}

func NewOpenSensor(ttnAccess TtnAccess, sensorAccess OpenSensorAccess) {

}

func Start(ttnAccess TtnAccess, sensorAccess OpenSensorAccess) {

	ctx := apex.Stdout().WithField(logName, "Go Client")
	log.Set(ctx)

	client := mqtt.NewClient(ctx, id, ttnAccess.AppId, ttnAccess.Key, ttnAccess.Broker)
	if err := client.Connect(); err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}
	token := client.SubscribeDeviceUplink(ttnAccess.AppId, ttnAccess.DeviceId,
		func(client mqtt.Client, appID string, devID string, req types.UplinkMessage) {
			log.Get().Info(string(req.PayloadRaw))
			uplink(req.PayloadRaw, sensorAccess)
		})
	token.Wait()
	if err := token.Error(); err != nil {
		ctx.WithError(err).Fatal("Could not subscribe")
	}
}

func uplink(payload []byte, access OpenSensorAccess) {

	data := openSensorData{payload}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(data)
	if err != nil {
		log.Get().Error(err.Error())
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	/*resp, err := http.Post(urlFormat(access), "application/json", b)
	defer 	resp.Body.Close()*/
	resp, err := (&http.Client{Transport: tr, Timeout: time.Second * 10}).Post(urlFormat(access), "application/json", b)
	//resp, err := http.Get("www.google.com")
	if err != nil {
		log.Get().Error(err.Error())
		return
	}
	re, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Get().Error(err.Error())
		return
	}
	fmt.Println("Parsed "+string(re))
}

func prepareRequest(method string, access OpenSensorAccess, b []byte) *http.Request {
	req, err := http.NewRequest(method, urlFormat(access), bytes.NewBuffer(b))
	if err != nil {
		log.Get().Error(err.Error())
		return nil
	}
	v := req.URL.Query()
	v.Add(uriClientId, access.ClientId)
	v.Add(uriPassword, access.Pw)
	req.URL.RawQuery = v.Encode()
	req.Header.Add("Authorization", "api-key "+access.Key)
	req.Header.Add("Content-Type", "application/json")
	return req
}

func urlFormat(access OpenSensorAccess) string {
	return /*OpenSensorURI +*/ access.Topic
}
