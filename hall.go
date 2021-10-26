package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	TIME_UNIT = 250

	WAITERS_NUMBER = 5
	ORDERS_NUMBER  = 10

	Foods []Food

	FinishedOrders = make(chan Order, 100)

	Tables = make([]chan Order, 10)

	OrderID      uint64
	OrdersNumber int64
)

func main() {

	rand.Seed(time.Now().UnixNano())
	// Prepare the data
	UnmarshalFood()

	// Prepare the tables chans
	for i := range Tables {
		Tables[i] = make(chan Order)
	}

	// Start the waiters
	for i := 0; i < WAITERS_NUMBER; i++ {
		go waiter(i)
	}

	// Wait for finished orders
	go waitForOrders()

	// Generate some orders
	go generateOrders()

	app := App{}
	app.Init()
	app.Run(":81")

}

func waiter(id int) {
	for {
		for i := 0; i < len(Tables); i++ {
			select {
			case order, ok := <-Tables[i]:
				if ok == false {
					continue
				}
				fmt.Printf("Order %v picked by %v\n", order.ID, id)
				close(Tables[i])
				order.TableID = i
				order.WaiterID = id
				order.PickUpTime = time.Now().UnixMilli()
				sendOrder(order)
			default:
			}
		}
	}
}

func sendOrder(order Order) {
	fmt.Printf("Order %v send to the kitchen\n", order.ID)
	url := "http://127.0.0.1:80/order"
	jsonValue, _ := json.Marshal(order)
	http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
}

func generateOrders() {
L:
	for {
		time.Sleep(time.Second * time.Duration(rand.Intn(5)))
		order := generateOrder()
		fmt.Printf("Order %v generated\n", order.ID)
		tableID := rand.Intn(len(Tables))
		select {
		case _, ok := <-Tables[tableID]:
			if ok == false {
				continue
			}
		default:
			fmt.Printf("Order %v sent to table %v\n", order.ID, tableID)
			Tables[tableID] <- order
			if atomic.AddInt64(&OrdersNumber, 1) >= int64(ORDERS_NUMBER) {
				break L
			}
		}
	}
}

func generateOrder() Order {
	var order Order
	var maxTime = 0
	order.ID = atomic.AddUint64(&OrderID, 1)
	for i := 0; i < rand.Intn(10)+1; i++ {
		foodID := rand.Intn(len(Foods))
		if maxTime < Foods[foodID].PreparationTime {
			maxTime = Foods[foodID].PreparationTime
		}
		order.Items = append(order.Items, foodID+1)
	}
	order.Priority = rand.Intn(5) + 1
	order.MaxWait = int(math.Ceil(float64(maxTime) * 1.3))
	return order
}

func sleep(i int) {
	time.Sleep(time.Second * time.Duration(i))
}

func waitForOrders() {
	for order := range FinishedOrders {
		Tables[order.TableID] = make(chan Order)
		fmt.Printf("%+v\n", order)
	}
}
