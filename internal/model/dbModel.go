package model

import "time"

type FlightRequest struct {
	AircraftType    string    `json:"aircraft_type"`
	FlightNumber    string    `json:"flight_number"`
	DepartureDate   time.Time `json:"departure_date"`
	ArrivalDate     time.Time `json:"arrival_date"`
	PassengersCount int       `json:"passengers_count"`
}

type FlightRequestData struct {
	Request FlightRequest
	MetaID  int
}
