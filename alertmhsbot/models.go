package main

import (
	"encoding/json"
	"fmt"
)

type RegionsData struct {
	Districts [][]string   `json:"districts"`
	Regions   [][][]string `json:"regions"`
}

type SessionData struct {
	Status     string
	DistrictId string
	RegionId   string
}

func NewSessionData(status string, districtId, regionId string) []byte {
	var sessionData SessionData
	sessionData.Status = status
	sessionData.DistrictId = districtId
	sessionData.RegionId = regionId

	b, err := json.Marshal(sessionData)
	if err != nil {
		fmt.Println("NewSessionData > ", err)
	}
	return b
}

func ParseSessionData(data []byte) SessionData {
	var sessionData SessionData
	err := json.Unmarshal(data, &sessionData)
	if err != nil {
		fmt.Println("ParseSessionData > ", err)
	}
	return sessionData
}

type PageData struct {
	Region []string `pagser:"td a"`
	Events []string `pagser:"td img->eachAttr(title)"`
}

type AlertTimer struct {
	ChatID int64
	Region string
}

type FunFacts struct {
	Facts []string `json:"facts"`
}

func ParseFunFacts(data []byte) FunFacts {
	var ff FunFacts
	err := json.Unmarshal(data, &ff)
	if err != nil {
		fmt.Println("FunFacts > ", err)
	}
	return ff
}
