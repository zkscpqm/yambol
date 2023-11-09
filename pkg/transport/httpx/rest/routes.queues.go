package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"yambol/config"

	"yambol/pkg/queue"
	"yambol/pkg/transport/httpx"
)

func (s *Server) queues() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		target, err := resolveHTTPMethodTarget(r, map[string]HandlerFunc{
			"GET":  s.getQueues(),
			"POST": s.addNewQueue(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err)
		}
		return target(w, r)
	}
}

func (s *Server) getQueues() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		resp := httpx.QueuesGetResponse(s.b.Stats())
		return s.respond(w, resp)
	}
}

func (s *Server) addNewQueue() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		qInfo := httpx.QueuesPostRequest{}

		if err := json.NewDecoder(r.Body).Decode(&qInfo); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
		}
		if s.b.QueueExists(qInfo.Name) {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to create queue `%s` as it already exists", qInfo.Name))
		}
		if !isValidPath(qInfo.Name) {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("the queue name `%s` is not valid", qInfo.Name))
		}

		if err := s.b.AddQueue(qInfo.Name, config.QueueConfig{
			MinLength:    qInfo.MinLength,
			MaxLength:    qInfo.MaxLength,
			MaxSizeBytes: qInfo.MaxSizeBytes,
			TTL:          qInfo.TTL,
		}); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to create queue `%s`: %v", qInfo.Name, err))
		}

		s.addQueueRoute(qInfo.Name, httpx.DebugPrintHook(s.logger))
		return s.respond(w, httpx.EmptyResponse{StatusCode: http.StatusCreated})
	}
}

func (s *Server) queue() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		target, err := resolveHTTPMethodTarget(r, map[string]HandlerFunc{
			"GET":    s.consumeFromQueue(),
			"POST":   s.sendMessageToQueue(),
			"DELETE": s.deleteQueue(),
		})
		if err != nil {
			return s.error(w, http.StatusMethodNotAllowed, err)
		}
		return target(w, r)
	}
}

func (s *Server) consumeFromQueue() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		qName := r.URL.Path[len("/queues/"):]

		if !s.b.QueueExists(qName) {
			return s.error(w, http.StatusNotFound, fmt.Errorf("queue `%s` does not exist", qName))
		}

		message, err := s.b.Consume(qName)

		if err != nil && !errors.Is(err, queue.ErrQueueEmpty) {
			return s.error(w, http.StatusInternalServerError, err)
		}

		return s.respond(w, httpx.QueueGetResponse{StatusCode: 200, Data: message})
	}
}

func (s *Server) sendMessageToQueue() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		qName := r.URL.Path[len("/queues/"):]

		if !s.b.QueueExists(qName) {
			return s.error(w, http.StatusNotFound, fmt.Errorf("queue `%s` does not exist", qName))
		}

		body := httpx.MessageRequest{}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return s.error(w, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
		}

		if err := s.b.Publish(body.Message, qName); err != nil {
			return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to publish message: %v", err))
		}

		return s.respond(w, httpx.EmptyResponse{StatusCode: http.StatusOK})
	}
}

func (s *Server) deleteQueue() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) httpx.Response {
		qName := r.URL.Path[len("/queues/"):]

		if !s.b.QueueExists(qName) {
			return s.error(w, http.StatusNotFound, fmt.Errorf("queue `%s` does not exist", qName))
		}

		if err := s.b.RemoveQueue(qName); err != nil {
			return s.error(w, http.StatusInternalServerError, fmt.Errorf("failed to remove queue `%s`: %v", qName, err))
		}

		return s.respond(w, httpx.EmptyResponse{StatusCode: http.StatusOK})
	}
}

func (s *Server) addQueueRoute(qName string, hooks ...httpx.Middleware) {
	s.route(
		fmt.Sprintf("/queues/%s", qName),
		s.queue(),
		hooks...,
	).Methods("GET", "POST", "DELETE")
}
