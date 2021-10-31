package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TIME_UNIT  = 100
	ORDER_TIME = 800

	WAITERS_NUMBER = 5
	ORDERS_NUMBER  = 0
	TABLES_NUMBER  = 10
)

var (
	TotalRating   = 0.0
	ReviewsNumber = 0.0
	RatingChan    = make(chan float64, 100)

	waitersMutex [WAITERS_NUMBER]sync.Mutex

	Foods []Food

	FinishedOrders = make(chan Order, 100)

	Tables = make([]chan Order, TABLES_NUMBER)

	OrderID      uint64
	OrdersNumber int64
)

func main() {

	// Prepare the data
	rand.Seed(time.Now().UnixNano())
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

	go CalculateRating()

	// Generate some orders
	go generateOrders()

	app := App{}
	app.Init()
	app.Run(":81")

}

func waiter(id int) {
	for {
	FOR:
		for i := 0; i < len(Tables); i++ {
			waitersMutex[id].Lock()
			select {
			case order, ok := <-Tables[i]:
				if ok == false {
					waitersMutex[id].Unlock()
					continue
				}
				fmt.Printf("Order %v picked by %v\n", order.ID, id)
				close(Tables[i])
				order.TableID = i
				order.WaiterID = id
				order.PickUpTime = time.Now().UnixMilli()
				sendOrder(order)
				waitersMutex[id].Unlock()
				break FOR
			default:
			}
			waitersMutex[id].Unlock()
		}
	}
}

func sendOrder(order Order) {
	fmt.Printf("Order %v sent to the kitchen\n", order.ID)
	url := "http://172.17.0.2:80/order"
	jsonValue, _ := json.Marshal(order)
	http.Post(url, "application/json", bytes.NewBuffer(jsonValue))
}

func generateOrders() {
L:
	for {
		if atomic.LoadInt64(&OrdersNumber) > 0 {
			tableID := rand.Intn(len(Tables))
			select {
			case tmp, ok := <-Tables[tableID]:
				if ok == false {
					continue L
				}
				Tables[tableID] <- tmp

			default:
				time.Sleep(time.Millisecond * ORDER_TIME)
				order := generateOrder()
				Tables[tableID] <- order
				atomic.AddInt64(&OrdersNumber, -1)
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
	time.Sleep(time.Duration(TIME_UNIT * i))
}

func waitForOrders() {
	for order := range FinishedOrders {
		go deliverOrder(order)
	}
}

func deliverOrder(order Order) {
	waitersMutex[order.WaiterID].Lock()
	rating := getStar(order.CookingTime, order.MaxWait)
	fmt.Printf("%v order maxWait: %v, cookingTime: %v, rating: %v\n", order.ID, order.MaxWait, order.CookingTime, rating)
	RatingChan <- float64(rating)
	Tables[order.TableID] = make(chan Order)
	waitersMutex[order.WaiterID].Unlock()
}

func getStar(wait float32, max int) int {
	waitTime := float32(wait)
	maxTime := float32(max)
	if waitTime < maxTime {
		return 5
	}
	if waitTime <= maxTime*1.1 {
		return 4
	}
	if waitTime <= maxTime*1.2 {
		return 3
	}
	if waitTime <= maxTime*1.3 {
		return 2
	}
	if waitTime <= maxTime*1.4 {
		return 1
	}
	return 0
}

func CalculateRating() {
	for rating := range RatingChan {
		TotalRating = (TotalRating*ReviewsNumber + rating) / (ReviewsNumber + 1)
		fmt.Printf("Restaurant raiting: %.2f\n", TotalRating)
		ReviewsNumber += 1
	}
}
