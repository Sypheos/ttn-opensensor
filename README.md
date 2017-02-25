Opensensor integration for the Thing Network


Act as a "router" for RawPayload data in TTN Uplink message to Opensor topic.
The payload will converted to golang string default format then published on
OpenSensor.io topic

TRUST ALL the certificate for the https connection. (certification check
deactivated)


# How to

As in the tests in client_test.go or in example\main.go. Fill the access
structures with your integration definition.

Call the NewOpenSensor and the Start() on the returned pointer.

# Tests

for the tests you have to run an mqtt broker platform on your local machine.
You can get one with
```
$ sudo docker run -it -p 1883:1883  toke/mosquitto
```

Then:
```
go test
```
