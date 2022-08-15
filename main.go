package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/taxi/rnd"
	"github.com/taxi/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

type ByViews []service.OrderStats

func (b ByViews) Len() int {
	return len(b)
}

func (b ByViews) Less(i, j int) bool {
	switch {
	case b[i].ViewCount != b[j].ViewCount:
		return b[i].ViewCount > b[j].ViewCount
	default:
		return strings.Compare(b[i].Code, b[j].Code) < 0
	}
}

func (b ByViews) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool)
	router := mux.NewRouter()
	serv := service.NewAutoOrderService(50, done)

	router.HandleFunc("/request", RequestHandler(serv))
	router.HandleFunc("/admin/request", AdminRequestHandler(serv))

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatalf("failed start Server due to: %s\n", err)
		} else {
			log.Println("Server successfully started")
		}
	}()

	<-stop
	done <- true

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("error shutting down server %s", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}

func RequestHandler(s *service.OrderService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		order := s.NextOrder(selectOrder)
		_, _ = w.Write([]byte(order.Code))
	}
}

func AdminRequestHandler(s *service.OrderService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := s.GetStats()
		sort.Sort(ByViews(stats))
		for _, s := range stats {
			_, _ = w.Write([]byte(fmt.Sprintf("%s - %d\n", s.Code, s.ViewCount)))
		}
	}
}

func selectOrder(orders []service.Order) service.Order {
	return orders[rnd.RandomInt(len(orders))]
}
