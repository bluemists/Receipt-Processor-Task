package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ProdItem struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string     `json:"retailer"`
	PurchaseDate string     `json:"purchaseDate"`
	PurchaseTime string     `json:"purchaseTime"`
	Items        []ProdItem `json:"items"`
	Total        string     `json:"total"`
	Id           uint64     `json:"id"`

	inTotal float32
}

type ReciptID struct {
	Id uint64 `json:"id"`
}

type PointResults struct {
	Points int `json:"points"`
}

var AllReceipts []Receipt

// /
// Reads in receipt (as json) and saves the data to global receipt structure
// Parameters: Receipt Json
// Returns:id of saved receipt(json)
// /
func ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(" An issue occured", err)
	}

	receiptData := Receipt{}

	json.Unmarshal(bodyBytes, &receiptData)

	newId := rand.Int63n(7000)
	receiptData.Id = uint64(newId)
	AllReceipts = append(AllReceipts, receiptData)

	//response
	w.Header().Set("Content-Type", "Application/json")
	response := ReciptID{Id: receiptData.Id}
	json.NewEncoder(w).Encode(response)
}

// /
// Calculate the points to be awarded based on receipt contents
// Parameters: receipt id
// Return: Point total(json)
// /
func GetPoints(w http.ResponseWriter, r *http.Request) {

	selectedReceipt := Receipt{}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(" An issue occured", err)
	}
	receiptid := ReciptID{}

	json.Unmarshal(bodyBytes, &receiptid)

	for i := 0; i < len(AllReceipts); i++ {
		if AllReceipts[i].Id == receiptid.Id {
			selectedReceipt = AllReceipts[i]
		}
	}

	tempTotal, inerr := strconv.ParseFloat(selectedReceipt.Total, 32)
	if inerr != nil {
		fmt.Println(w, "Error converting original total.")
	}

	//is used for overall points earned
	pointCounter := 0

	////////////////////
	//Determining points
	////////////////////

	alphanum := 0

	for _, r := range selectedReceipt.Retailer {
		if unicode.IsNumber(r) || unicode.IsLetter(r) {
			alphanum++
		}
	}
	pointCounter += alphanum

	//checking if is an even dollar
	if tempTotal > 1 && int32((tempTotal*100))%100 == 0 {
		pointCounter += 50
	}

	log.Printf(" CHK 2:{0}", pointCounter)
	//check if multiple of 25

	intermid := float64(tempTotal)
	modres := math.Mod(intermid, .25)
	if modres == 0 {
		pointCounter += 25
	}

	//calculating points per item pair
	pointCounter += (len(selectedReceipt.Items) / 2 * 5)

	//Calculating letters and points

	for itemcounter := 0; itemcounter < len(selectedReceipt.Items); itemcounter++ {

		count := utf8.RuneCountInString(strings.TrimSpace(selectedReceipt.Items[itemcounter].ShortDescription))
		letterMod := math.Mod(float64(count), 3)
		if letterMod == 0 {
			PointPrice, err3 := strconv.ParseFloat(selectedReceipt.Items[itemcounter].Price, 32)

			if err3 != nil {
				fmt.Print(w, "Error on string to float conv")
			}
			PointPrice *= 100 * .2
			diff := int(PointPrice) % 100

			if diff < 1 {
				PointPrice -= float64(diff)
				PointPrice /= 100
				pointCounter += int(PointPrice)
			} else {
				PointPrice -= float64(diff)
				PointPrice /= 100
				pointCounter += int(PointPrice)
				pointCounter += 1
			}
		}
	}

	//Calulating if Odd Date
	datestring := strings.Split(selectedReceipt.PurchaseDate, "-")
	result, converr := strconv.Atoi(datestring[len(datestring)-1])
	if converr != nil {
		fmt.Print(w, "Error checking date for odd value.", converr, datestring[len(datestring)-1])
	}
	if result%2 != 0 {
		pointCounter += 6
	}

	//Calculating Time
	timestring := strings.Split(selectedReceipt.PurchaseTime, ":")
	timeresultHr, converrTime := strconv.Atoi(timestring[0])
	timeresultMin, converrTimemin := strconv.Atoi(timestring[1])

	if converrTime != nil && converrTimemin != nil {
		fmt.Print(w, "Error checking date for odd value.", converrTime)
	}
	if (timeresultHr >= 14 && timeresultMin > 0) && (timeresultHr < 16) {
		pointCounter += 10
	}

	//Response
	w.Header().Set("Content-Type", "Application/json")

	response := PointResults{Points: pointCounter}
	json.NewEncoder(w).Encode(response)
}

func HandleRequest() {
	http.HandleFunc("/Receipts/process", ProcessReceipt)
	http.HandleFunc("/receipts/{id}/points", GetPoints)
	log.Fatal(http.ListenAndServe("", nil))
}

func main() {
	HandleRequest()
}
