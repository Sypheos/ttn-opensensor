Opensensor integration for the Thing Network

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
