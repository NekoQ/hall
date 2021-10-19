package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Food struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	PreparationTime  int    `json:"preparation-time"`
	Complexity       int    `json:"complexity"`
	CookingApparatus string `json:"cooking-apparatus"`
	orderID          int
	cookID           int
}

func UnmarshalFood() {

	foodJson, err := os.Open("foods.json")
	if err != nil {
		fmt.Println(err)
	}
	defer foodJson.Close()

	bypeValue, _ := ioutil.ReadAll(foodJson)

	err = json.Unmarshal(bypeValue, &Foods)
}
