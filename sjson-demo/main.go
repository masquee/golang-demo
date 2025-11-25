package main

import "github.com/tidwall/sjson"

const json = `{"name":{"first":"Janet","last":"Prichard"},"age":47}`

func main() {
	value, _ := sjson.Set(json, "name.last", "Anderson")
	println(value) // {"name":{"first":"Janet","last":"Anderson"},"age":47}

	value, _ = sjson.Set(`{"friends":["Andy","Carol"]}`, "friends.2", "Sara")
	println(value) // {"friends":["Andy","Carol","Sara"]}

	// Append an array value by using the -1 key in a path:
	value, _ = sjson.Set(`{"friends":["Andy","Carol"]}`, "friends.-1", "Sara")
	println(value) // {"friends":["Andy","Carol","Sara"]}

	// Delete a value:
	value, _ = sjson.Delete(`{"name":{"first":"Sara","last":"Anderson"}}`, "name.first")
	println(value) // {"name":{"last":"Anderson"}}
}
