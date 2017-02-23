package opensensor

import (
	"bytes"
	"encoding/json"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/magiconair/properties/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestUplink(t *testing.T) {

	sensor := OpenSensorAccess{"5885", "yEUIsPrx", "5b46e6e9-d572-49a7-bce8-bfcf5362550c",
		"http://localhost:3000"}
	ttna := TtnAccess{"open-sensor", "ttn-account-v2.CLWM-c78CsFxUUZPfXCe9933kdVHdV1nIzrNk-kApP8",
		"tcp://localhost:1883", "heater"}
	o := NewOpenSensor(ttna, sensor)
	o.Start(ttna, sensor)
	t.Run("httpServ", func(t *testing.T) {
		ctx := apex.Stdout().WithField("TestClientPub", "testClient")
		client := mqtt.NewClient(ctx, "ttnctl", ttna.AppId, ttna.Key, ttna.Broker)
		if err := client.Connect(); err != nil {
			t.Fatal(err)
		}
		defer client.Disconnect()
		<-time.After(time.Millisecond * 500)
		token := client.PublishUplink(types.UplinkMessage{AppID: ttna.AppId, DevID: ttna.DeviceId, PayloadRaw: []byte("{\"temp\":20}")})
		token.Wait()
		if token.Error() != nil {
			t.Fatal(token.Error().Error())
		}
	})
	ch := make(chan []byte)
	go func() {
		http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
			str, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err.Error())
			}
			ch <- str
		})
		http.ListenAndServe(":3000", nil)
	}()
	select {
	case str := <-ch:
		buff := new(bytes.Buffer)
		buff.Write(str)
		r := openSensorData{}
		err := json.NewDecoder(buff).Decode(&r)
		if err != nil {
			t.Fatal(err.Error())
		}
		d := openSensorData{[]byte("{\"temp\":20}")}
		assert.Equal(t, r, d)
	}
}
