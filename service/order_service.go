package service

import (
	"errors"
	"github.com/taxi/rnd"
	"sync"
	"sync/atomic"
	"time"
)

// Order contains taxi order information
type Order struct {
	Code string
}

// OrderStats contain specific order statistics
type OrderStats struct {
	Order
	ViewCount uint32
}

// OrderService responsible for taxi orders management
type OrderService struct {
	orderMut sync.RWMutex // locks orders access
	orders   []Order
	statsMut sync.RWMutex // locks stats access
	stats    map[string]*uint32
	newOrder func() Order
}

type OrderServiceConfig struct {
	PoolSize   int  // max orders size handled by OrderService
	AutoUpdate bool // if true - activates async order updates
	NewOrder   func() Order
}

func NewOrderService(config OrderServiceConfig, done chan bool) *OrderService {
	if config.PoolSize <= 0 {
		panic(errors.New("config.PoolSize must be greater than 0"))
	}

	service := &OrderService{
		orders:   generateOrders(config.NewOrder, config.PoolSize),
		stats:    make(map[string]*uint32),
		newOrder: config.NewOrder,
	}

	if config.AutoUpdate {
		service.runOrderUpdater(done)
	}

	return service
}

// NewAutoOrderService creates new OrderService which automatically updates orders
func NewAutoOrderService(ordersPoolSize int, done chan bool) *OrderService {
	newOrder := func() Order {
		return Order{rnd.RandomStr(2)}
	}

	conf := OrderServiceConfig{
		PoolSize:   ordersPoolSize,
		AutoUpdate: true,
		NewOrder:   newOrder,
	}

	return NewOrderService(conf, done)
}

// NextOrder Returns the next taxi order available in the pool using selector
func (s *OrderService) NextOrder(selector func([]Order) Order) Order {
	s.orderMut.RLock()
	defer s.orderMut.RUnlock()

	order := selector(s.orders)

	// updates order's view stats
	go func(o Order) {
		s.statsMut.Lock()
		defer s.statsMut.Unlock()

		if _, ok := s.stats[o.Code]; ok {
			atomic.AddUint32(s.stats[o.Code], 1)
		} else {
			var v uint32 = 1
			s.stats[o.Code] = &v
		}

	}(order)

	return order
}

// GetStats returns all orders statistics
func (s *OrderService) GetStats() []OrderStats {
	s.statsMut.RLock()
	defer s.statsMut.RUnlock()

	res := make([]OrderStats, 0, len(s.stats))

	for key, val := range s.stats {
		res = append(res, OrderStats{
			Order:     Order{key},
			ViewCount: *val,
		})
	}

	return res
}

// runOrderUpdater runs process which updates orders in async mode
func (s *OrderService) runOrderUpdater(done chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				// replace one random order
				s.orderMut.Lock()
				rndOrder := rnd.RandomInt(len(s.orders))
				s.orders[rndOrder] = s.newOrder()
				s.orderMut.Unlock()
			case <-done:
				return
			}
		}
	}()
}

func generateOrders(orderGen func() Order, size int) []Order {
	arr := make([]Order, size)

	for i := range arr {
		arr[i] = orderGen()
	}

	return arr
}
