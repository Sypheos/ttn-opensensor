package opensensor

import (
	"testing"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/ttn/core/types"
	"encoding/json"
	"time"
	"io/ioutil"
	"net/http"
	"github.com/magiconair/properties/assert"
)

func TestUplink(t *testing.T) {

	sensor := OpenSensorAccess{"5885", "yEUIsPrx", "5b46e6e9-d572-49a7-bce8-bfcf5362550c",
					      "http://localhost:3000"}
	ttna := TtnAccess{"open-sensor", "ttn-account-v2.CLWM-c78CsFxUUZPfXCe9933kdVHdV1nIzrNk-kApP8",
		"tcp://eu.thethings.network:1883", "heater"}
	t.Run("httpServ", func(t *testing.T) {
		Start(ttna, sensor)
		ctx := apex.Stdout().WithField("TestClientPub", "testClient")
		client := mqtt.NewClient(ctx, "ttnctl", ttna.AppId, ttna.Key, ttna.Broker)
		if err := client.Connect(); err != nil {
			//ctx.WithError(err).Fatal("Could not connect")
			t.Fatal(err)
		}
		defer client.Disconnect()
		data, err := json.Marshal("{\"temp\":20}")
		if err != nil {
			t.Fatal(err.Error())
		}
		<-time.After(time.Millisecond*500)
		token := client.PublishUplink(types.UplinkMessage{AppID: ttna.AppId, DevID: ttna.DeviceId, PayloadRaw:data})
		token.Wait()
		if token.Error() != nil {
			t.Fatal(token.Error().Error())
		}
		ctx.Info(string(data))
	})
	ch := make(chan string)
	go func () {
		http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
			str, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err.Error())
			}
			ch <- string(str)
		})
		http.ListenAndServe(":3000", nil)
	}()
	select {
	case str := <- ch:
		assert.Equal(t, str, "{\"data\":{\"temp\":20}}")
	}
}
