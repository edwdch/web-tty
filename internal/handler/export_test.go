package handler

import "time"

func OverrideKeepAlive(pong, ping time.Duration) func() {
	prevPong, prevPing := pongWait, pingPeriod
	pongWait, pingPeriod = pong, ping
	return func() {
		pongWait, pingPeriod = prevPong, prevPing
	}
}
