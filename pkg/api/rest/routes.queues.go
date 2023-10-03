package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yambol/pkg/broker"

	"yambol/pkg/queue"
)

func (s *YambolHTTPServer) queues() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		target, err := resolveHTTPMethodTarget(r, map[string]yambolHandlerFunc{
			"GET":  s.getQueues(),
			"POST": s.postQueues(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err)
		}
		return target(w, r)
	}
}

func (s *YambolHTTPServer) getQueues() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		resp := QueuesGetResponse(s.b.Stats())
		s.respond(w, resp)
		return resp
	}
}

func (s *YambolHTTPServer) addQueueRoute(qName string, hooks ...Hook) {
	s.route(
		fmt.Sprintf("/queues/%s", qName),
		s.queue(),
		hooks...,
	).Methods("GET", "POST")
}

func (s *YambolHTTPServer) postQueues() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		qInfo := QueuesPostRequest{}

		if err := json.NewDecoder(r.Body).Decode(&qInfo); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
		}
		if s.b.QueueExists(qInfo.Name) {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to create queue `%s` as it already exists", qInfo.Name))
		}
		if !isValidPath(qInfo.Name) {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("the queue name `%s` is not valid", qInfo.Name))
		}

		if err := s.b.AddQueue(qInfo.Name, broker.QueueOptions{
			MinLen:       qInfo.MinLength,
			MaxLen:       qInfo.MaxLength,
			MaxSizeBytes: qInfo.MaxSizeBytes,
			DefaultTTL:   qInfo.TTLSeconds(),
		}); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to create queue `%s`: %v", qInfo.Name, err))
		}

		s.addQueueRoute(qInfo.Name, debugPrintHook())

		resp := EmptyResponse{statusCode: http.StatusCreated}
		s.respond(w, resp)
		return resp
	}
}

func (s *YambolHTTPServer) queue() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		target, err := resolveHTTPMethodTarget(r, map[string]yambolHandlerFunc{
			"GET":  s.getQueue(),
			"POST": s.postQueue(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err)
		}
		return target(w, r)
	}
}

func (s *YambolHTTPServer) getQueue() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		qName := r.URL.Path[len("/queues/"):]

		if !s.b.QueueExists(qName) {
			return s.error(w, http.StatusNotFound, fmt.Errorf("queue `%s` does not exist", qName))
		}

		message, err := s.b.Consume(qName)

		if err != nil && !errors.Is(err, queue.ErrQueueEmpty) {
			return s.error(w, http.StatusInternalServerError, err)
		}

		resp := QueueGetResponse{statusCode: 200, Data: message}
		s.respond(w, resp)
		return resp
	}
}

func (s *YambolHTTPServer) postQueue() yambolHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) Response {
		qName := r.URL.Path[len("/queues/"):]

		if !s.b.QueueExists(qName) {
			return s.error(w, http.StatusNotFound, fmt.Errorf("queue `%s` does not exist", qName))
		}

		body := YambolMessageRequest{}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
		}

		if err := s.b.Publish(body.Message, qName); err != nil {
			return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to publish message: %v", err))
		}

		resp := EmptyResponse{statusCode: http.StatusOK}
		s.respond(w, resp)
		return resp
	}
}
