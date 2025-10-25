// Package order proides functionality for managing orders.
package order

type Status string

const (
	Pending    Status = "pending"
	Processing Status = "processing"
	Complete   Status = "complete"
)

type Role string

const (
	Normal Role = "normal"
	VIP    Role = "vip"
)

var id = 0

type Order struct {
	Status    Status
	OrderType Role
	ID        int
}

func (order *Order) SetOrderComplete() {
	order.Status = Complete
}

func (order *Order) SetOrderProcessing() {
	order.Status = Processing
}

func (order *Order) SetOrderPending() {
	order.Status = Pending
}

func NormalOrder() *Order {
	normalOrder := &Order{
		Status:    Pending,
		OrderType: Normal,
		ID:        id,
	}

	id++

	return normalOrder
}

func VIPOrder() *Order {
	vipOrder := &Order{
		Status:    Pending,
		OrderType: VIP,
		ID:        id,
	}

	id++

	return vipOrder
}
