package controller

import (
	"fmt"
	"sync"
	"time"

	botpkg "github.com/jason0w0/se-take-home-assignment/libs/bot"
	orderpkg "github.com/jason0w0/se-take-home-assignment/libs/order"
)

type Controller struct {
	Bots           []*botpkg.Bot
	PendingQueue   []*orderpkg.Order
	CompletedQueue []*orderpkg.Order

	botMu       sync.Mutex
	pendingMu   sync.Mutex
	completedMu sync.Mutex
}

func NewController() *Controller {
	return &Controller{}
}

// AddNormalOrder append normal order to pending queue
func (controller *Controller) AddNormalOrder() {
	controller.pendingMu.Lock()
	defer controller.pendingMu.Unlock()

	normalOrder := orderpkg.NormalOrder()
	controller.PendingQueue = append(controller.PendingQueue, normalOrder)
	writeToFile(fmt.Sprintf("Order %d has been added", normalOrder.ID))

	controller.AssignOrderToBot(normalOrder)
}

// AddVipOrder insert vip order into the front of pending queue but behind existing VIP orders
func (controller *Controller) AddVipOrder() {
	controller.pendingMu.Lock()
	defer controller.pendingMu.Unlock()

	vipOrder := orderpkg.VIPOrder()

	pos := 0
	for i, order := range controller.PendingQueue {
		if order.OrderType == orderpkg.Normal {
			break
		}

		pos = i + 1
	}

	controller.PendingQueue = append(controller.PendingQueue, nil)
	copy(controller.PendingQueue[pos+1:], controller.PendingQueue[pos:])
	controller.PendingQueue[pos] = vipOrder
	writeToFile(fmt.Sprintf("Order %d has been added", vipOrder.ID))

	controller.AssignOrderToBot(vipOrder)
}

// AssignOrderToBot finds an IDLE bot assign the order to it
func (controller *Controller) AssignOrderToBot(order *orderpkg.Order) {
	controller.botMu.Lock()
	defer controller.botMu.Unlock()

	for _, bot := range controller.Bots {
		if bot.Status == botpkg.IDLE {
			bot.OrderChannel <- struct{}{}
			break
		}
	}

	writeToFile(fmt.Sprintf("Order %d has been assigned to a bot", order.ID))
}

// AddBot adds a bot to controller and starts the bot immediately
func (controller *Controller) AddBot() {
	controller.botMu.Lock()
	defer controller.botMu.Unlock()

	bot := botpkg.NewBot(controller)
	controller.Bots = append(controller.Bots, bot)

	go bot.Run()

	<-bot.ReadyChannel
	writeToFile("A bot has been added")
}

// RemoveBot stops the bot and remove it from controller
func (controller *Controller) RemoveBot() {
	controller.botMu.Lock()
	defer controller.botMu.Unlock()

	if len(controller.Bots) == 0 {
		return
	}

	bot := controller.Bots[len(controller.Bots)-1]
	close(bot.StopChannel)

	controller.Bots = controller.Bots[:len(controller.Bots)-1]
	writeToFile("A bot has been removed")
}

// GetNextOrder find the next pending order to be process, return nil if non
func (controller *Controller) GetNextOrder() *orderpkg.Order {
	controller.pendingMu.Lock()
	defer controller.pendingMu.Unlock()

	for _, order := range controller.PendingQueue {
		if order.Status == orderpkg.Pending {
			order.SetOrderProcessing()
			return order
		}
	}

	return nil
}

// SetOrderCompleted moves the order from pending queue to completed queue and mark it as complete
func (controller *Controller) SetOrderCompleted(orderID int) {
	controller.pendingMu.Lock()
	defer controller.pendingMu.Unlock()

	idx := 0
	for i, order := range controller.PendingQueue {
		if order.ID == orderID {
			controller.completedMu.Lock()
			defer controller.completedMu.Unlock()

			order.SetOrderComplete()
			controller.CompletedQueue = append(controller.CompletedQueue, order)
			writeToFile(fmt.Sprintf("Order %d has been completed", orderID))
			idx = i
			break
		}
	}

	controller.PendingQueue = append(controller.PendingQueue[:idx], controller.PendingQueue[idx+1:]...)
}

// SetOrderPending sets the order status to pending and try to assign the order to a bot.
// Called when a bot is stopped when processing an order.
func (controller *Controller) SetOrderPending(orderID int) {
	controller.pendingMu.Lock()
	defer controller.pendingMu.Unlock()

	for _, order := range controller.PendingQueue {
		if order.ID == orderID {
			order.SetOrderPending()
			controller.AssignOrderToBot(order)
		}
	}
}

func (controller *Controller) ListOrders() {
	fmt.Println("Pending orders:")
	for _, order := range controller.PendingQueue {
		fmt.Println(order)
	}

	fmt.Println("Completed orders:")
	for _, order := range controller.CompletedQueue {
		fmt.Println(order)
	}
}

func (controller *Controller) ListBots() {
	fmt.Println("Bots:")
	for _, bot := range controller.Bots {
		fmt.Println(bot)
	}
}

func writeToFile(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("15:04:05"), msg)
}
