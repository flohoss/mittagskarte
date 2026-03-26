package events

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse/v2"
)

var server *sse.Server

const StreamName = "restaurants"

type Event struct {
	SSE *sse.Server
}

func New(onSubscribe func(streamID string, sub *sse.Subscriber)) *Event {
	server = sse.NewWithCallback(onSubscribe, nil)
	server.AutoReplay = false
	server.CreateStream(StreamName)
	return &Event{SSE: server}
}

func (e *Event) SendUpdate(content any) {
	data, _ := json.Marshal(content)
	e.SSE.Publish(StreamName, &sse.Event{
		Data: data,
	})
}

func (e *Event) GetHandler() echo.HandlerFunc {
	return echo.WrapHandler(http.HandlerFunc(e.SSE.ServeHTTP))
}
