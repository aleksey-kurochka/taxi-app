package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/taxi/rnd"
	"testing"
)

func TestOrderService_Init(t *testing.T) {
	done := make(chan bool, 1)
	done <- true

	conf := OrderServiceConfig{
		AutoUpdate: false,
		NewOrder: func() Order {
			return Order{rnd.RandomStr(2)}
		},
		PoolSize: 10,
	}

	s := NewOrderService(conf, done)
	assert.Equal(t, 10, len(s.orders))
}

func TestOrderService_NextOrder(t *testing.T) {
	done := make(chan bool, 1)
	done <- true

	conf := OrderServiceConfig{
		AutoUpdate: false,
		NewOrder: func() Order {
			return Order{rnd.RandomStr(2)}
		},
		PoolSize: 10,
	}
	s := NewOrderService(conf, done)
	order := s.NextOrder(func(orders []Order) Order {
		return orders[0]
	})
	assert.Equal(t, order, s.orders[0])
}

func TestOrderService_GetStats(t *testing.T) {
	done := make(chan bool, 1)
	done <- true

	codes := []string{"a1", "a2", "a3"}
	i := 0
	newOrder := func() Order {
		order := Order{codes[i]}
		i++
		return order
	}

	conf := OrderServiceConfig{
		PoolSize:   3,
		AutoUpdate: false,
		NewOrder:   newOrder,
	}

	s := NewOrderService(conf, done)

	go s.NextOrder(func(orders []Order) Order {
		return orders[0]
	})

	go s.NextOrder(func(orders []Order) Order {
		return orders[0]
	})

	go s.NextOrder(func(orders []Order) Order {
		return orders[1]
	})

	stats := s.GetStats()

	for _, stat := range stats {
		if stat.Code == "a1" {
			assert.Equal(t, 2, stat.ViewCount)
		}
		if stat.Code == "a2" {
			assert.Equal(t, 1, stat.ViewCount)
		}
	}
}
