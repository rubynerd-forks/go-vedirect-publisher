module github.com/seanhood/go-vedirect-publisher

go 1.14

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/seanhood/go-vedirect v0.0.0-20201007195155-417fb15171eb
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	go.bug.st/serial v1.6.4
	golang.org/x/net v0.0.0-20201020065357-d65d470038a5 // indirect
)

replace go.bug.st/serial => /Users/rubynerd/Developer/go-serial