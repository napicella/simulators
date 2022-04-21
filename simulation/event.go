package sim

// Event in the simulation. An event is a timestamped structured which contains that wraps
// the logic that needs to run when the event triggers
type Event struct {
	// Time of the event
	Time float64
	// The callback to execute when the event triggers
	CallbackFun Callback
	// Payload to pass to the CallbackFun
	Payload interface{}
}

// Callback is a function associated to an event. It receives as input the time of the
// event that triggered the callback and the payload of the event. Returns zero or more
// events that are triggered by the callback function
type Callback func(t float64, payload interface{}) []Event
