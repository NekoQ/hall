package main

var (
	TIME_UNIT = 250

	Foods = make([]Food, 100)
)

func main() {

	// Prepare the data
	UnmarshalFood()

	app := App{}
	app.Init()
	app.Run(":81")

}
