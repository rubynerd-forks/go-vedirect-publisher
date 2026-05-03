module github.com/seanhood/go-vedirect-publisher

go 1.25

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/seanhood/go-vedirect v0.0.0-20201007195155-417fb15171eb
	go.bug.st/serial v1.6.4
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	golang.org/x/net v0.0.0-20201020065357-d65d470038a5 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace go.bug.st/serial => github.com/rubynerd-forks/go-serial v0.0.0-20250705232342-d80d66543bcc

replace github.com/seanhood/go-vedirect => github.com/rubynerd-forks/go-vedirect v0.0.0-20260430235030-6acd602b47b6
