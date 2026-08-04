package schema

type EmptyResult struct{}

const (
	ResultTypeComplete      = "complete"
	ResultTypeInputRequired = "input_required"
)

type PingResult EmptyResult

type SubscribeResult EmptyResult

type UnsubscribeResult EmptyResult

type SetLevelResult EmptyResult
