// Package bot proides functionality for managing cooking bot.
package bot

import (
	"time"

	ord "github.com/jason0w0/se-take-home-assignment/libs/order"
)

type Manager interface {
	GetNextOrder() *ord.Order
	SetOrderCompleted(orderID int)
	SetOrderPending(orderID int)
}

type Status string

const (
	IDLE = "idle"
	BUSY = "busy"
)

type Bot struct {
	manager      Manager
	StopChannel  chan bool
	ReadyChannel chan bool
	OrderChannel chan struct{}
	Status       Status
}

func NewBot(manager Manager) *Bot {
	return &Bot{
		manager:      manager,
		StopChannel:  make(chan bool, 1),
		ReadyChannel: make(chan bool, 1),
		OrderChannel: make(chan struct{}, 1),
		Status:       IDLE,
	}
}

func (bot *Bot) Run() {
	close(bot.ReadyChannel)
	for {
		select {
		case <-bot.StopChannel:
			return
		default:
			order := bot.manager.GetNextOrder()
			if order == nil {
				bot.Status = IDLE
				select {
				case <-bot.StopChannel:
					return
				case <-bot.OrderChannel:
					continue
				}
			}

			bot.processOrder(order.ID)
		}
	}
}

func (bot *Bot) processOrder(orderID int) {
	bot.Status = BUSY
	select {
	case <-bot.StopChannel:
		bot.manager.SetOrderPending(orderID)
	case <-time.After(10 * time.Second):
		bot.manager.SetOrderCompleted(orderID)
	}
}
