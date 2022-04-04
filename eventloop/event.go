package main

type event struct {
	time        float64
	callbackFun callback
	payload     interface{}
}

type callback func(t float64, payload interface{}) []event
