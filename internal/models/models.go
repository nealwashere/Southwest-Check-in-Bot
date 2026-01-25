package models

import "time"

type Reservation struct {
	ConfirmationNumber string
	FirstName          string
	LastName           string
	DepartureTime      time.Time
}

func (r *Reservation) CheckInTime() time.Time {
	return r.DepartureTime.Add(-24 * time.Hour)
}

type BoardingPass struct {
	FirstName        string
	LastName         string
	BoardingGroup    string
	BoardingPosition string
	FlightNumber     string
	DepartureAirport string
	ArrivalAirport   string
	DepartureTime    string
}

type CheckInResponse struct {
	Flights []FlightInfo `json:"flights"`
}

type FlightInfo struct {
	FlightNumber     string          `json:"flightNumber"`
	DepartureAirport AirportInfo     `json:"originationAirportCode"`
	ArrivalAirport   AirportInfo     `json:"destinationAirportCode"`
	DepartureTime    string          `json:"departureTime"`
	Passengers       []PassengerInfo `json:"passengers"`
}

type AirportInfo struct {
	Code string `json:"code"`
}

type PassengerInfo struct {
	Name             NameInfo `json:"name"`
	BoardingGroup    string   `json:"boardingGroup"`
	BoardingPosition string   `json:"boardingPosition"`
}

type NameInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type ReservationLookupResponse struct {
	CheckInViewReservationPage CheckInViewReservationPage `json:"checkInViewReservationPage"`
}

type CheckInViewReservationPage struct {
	PNR     string              `json:"pnr"`
	Flights []ReservationFlight `json:"flights"`
	Links   CheckInLinks        `json:"_links"`
}

type ReservationFlight struct {
	FlightNumber       string `json:"flightNumber"`
	DepartureTime      string `json:"departureTime"`
	OriginAirport      string `json:"originationAirportCode"`
	DestinationAirport string `json:"destinationAirportCode"`
}

type CheckInLinks struct {
	CheckIn CheckInLink `json:"checkIn"`
}

type CheckInLink struct {
	Href   string             `json:"href"`
	Method string             `json:"method"`
	Body   CheckInRequestBody `json:"body"`
}

type CheckInRequestBody struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	RecordLocator string `json:"recordLocator"`
}

type CheckInConfirmationResponse struct {
	CheckInConfirmationPage CheckInConfirmationPage `json:"checkInConfirmationPage"`
}

type CheckInConfirmationPage struct {
	Flights []ConfirmationFlight `json:"flights"`
}

type ConfirmationFlight struct {
	FlightNumber           string                  `json:"flightNumber"`
	OriginAirportCode      string                  `json:"originAirportCode"`
	DestinationAirportCode string                  `json:"destinationAirportCode"`
	DepartureTime          string                  `json:"departureTime"`
	Passengers             []ConfirmationPassenger `json:"passengers"`
}

type ConfirmationPassenger struct {
	Name             string `json:"name"`
	BoardingGroup    string `json:"boardingGroup"`
	BoardingPosition string `json:"boardingPosition"`
}
