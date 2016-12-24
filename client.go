package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	baseUrl = "https://owner-api.teslamotors.com"
)

type Auth struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int    `json:"created_at"`
}

type Vehicle struct {
	DisplayName string `json:"display_name"`
	ID          int64  `json:"id"`
	VehicleID   int64  `json:"vehicle_id"`
}

type VehicleResponse struct {
	Response []Vehicle `json:"response"`
}

type VehicleState struct {
	Odometer        float64 `json:"odometer"`
	FirmwareVersion string  `json:"car_version"`
}

type ChargeState struct {
	BatteryRange    float64 `json:"battery_range"`
	EstBatteryRange float64 `json:"est_battery_range"`
	BatteryLevel    int     `json:"battery_level"`
}

type StateResponse struct {
	Response struct {
		*VehicleState
		*ChargeState
	} `json:"response"`
}

type TeslaClient struct {
	client      *http.Client
	accessToken string
}

func NewTeslaClient() *TeslaClient {
	return &TeslaClient{
		client:      &http.Client{},
		accessToken: "",
	}
}

func (c *TeslaClient) createRequest(method string, endpoint string, body []byte) (*http.Request, error) {
	path := baseUrl + endpoint

	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return req, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	return req, nil
}

func (c *TeslaClient) issueRequest(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *TeslaClient) Authenticate(config Configuration) error {
	auth := &Auth{
		GrantType:    "password",
		ClientID:     config.ClientId,
		ClientSecret: config.ClientSecret,
		Email:        config.Username,
		Password:     config.Password,
	}
	data, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	req, err := c.createRequest("POST", "/oauth/token", data)
	if err != nil {
		return err
	}

	body, err := c.issueRequest(req)
	if err != nil {
		return err
	}

	token := &Token{}
	err = json.Unmarshal(body, token)
	c.accessToken = token.AccessToken

	return nil
}

func (c *TeslaClient) ListVehicles() ([]Vehicle, error) {
	req, err := c.createRequest("GET", "/api/1/vehicles", nil)
	if err != nil {
		return nil, err
	}

	body, err := c.issueRequest(req)
	if err != nil {
		return nil, err
	}

	resp := &VehicleResponse{}
	err = json.Unmarshal(body, resp)
	return resp.Response, err
}

func (c *TeslaClient) GetVehicleState(vehicle Vehicle) (*VehicleState, error) {
	endpoint := fmt.Sprintf("/api/1/vehicles/%d/data_request/vehicle_state", vehicle.ID)
	req, err := c.createRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.issueRequest(req)
	if err != nil {
		return nil, err
	}

	resp := &StateResponse{}
	err = json.Unmarshal(body, resp)
	return resp.Response.VehicleState, nil
}

func (c *TeslaClient) GetChargeState(vehicle Vehicle) (*ChargeState, error) {
	endpoint := fmt.Sprintf("/api/1/vehicles/%d/data_request/charge_state", vehicle.ID)
	req, err := c.createRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.issueRequest(req)
	if err != nil {
		return nil, err
	}

	resp := &StateResponse{}
	err = json.Unmarshal(body, resp)
	return resp.Response.ChargeState, nil
}
