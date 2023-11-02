package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-memdb"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type ReceiptPoints struct {
	Id     string
	Points int
}

type ReceiptID struct {
	Id string `json:"id"`
}

type Points struct {
	Points int `json:"points"`
}

var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		"point": {
			Name: "point",
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Id"},
				},
			},
		},
	},
}

func lenAlphanumericChars(s string) int {
	count := 0
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsNumber(char) {
			count++
		}
	}

	return count
}

func hasDecimalPart(s string) bool {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return f != math.Floor(f)
}

func isMultipleOfPoint25(s string) bool {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return math.Mod(f, 0.25) == 0
}

func isTrimmedLengthMultipleOf3(s string) bool {
	trimmed := strings.TrimSpace(s)
	length := len(trimmed)

	return length%3 == 0
}

func hasOddIntAtTheEnd(s string) bool {
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}

	day := date.Day()
	return day%2 == 1
}

func isBetween2pmAnd4pm(s string) bool {
	return strings.Compare(s, "14:01") >= 0 && strings.Compare(s, "15:59") <= 0
}

func calcReceiptPoints(receipt Receipt) int {
	points := 0

	points += 1 * lenAlphanumericChars(receipt.Retailer)

	if !hasDecimalPart(receipt.Total) {
		points += 50
	}

	if isMultipleOfPoint25(receipt.Total) {
		points += 25
	}

	points += 5 * (len(receipt.Items) / 2)

	for _, value := range receipt.Items {
		if isTrimmedLengthMultipleOf3(value.ShortDescription) {
			price, err := strconv.ParseFloat(value.Price, 64)
			if err != nil {
				panic(err)
			}

			points += int(math.Ceil(price * 0.2))
		}
	}

	if hasOddIntAtTheEnd(receipt.PurchaseDate) {
		points += 6
	}

	if isBetween2pmAnd4pm(receipt.PurchaseTime) {
		points += 10
	}

	return points
}

func main() {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/receipts/process", func(w http.ResponseWriter, r *http.Request) {
		var receipt Receipt
		json.NewDecoder(r.Body).Decode(&receipt)

		uuid := uuid.New()
		uuidString := uuid.String()
		points := calcReceiptPoints(receipt)

		txn := db.Txn(true)
		err := txn.Insert("point", &ReceiptPoints{uuidString, points})
		if err != nil {
			panic(err)
		}
		txn.Commit()

		json.NewEncoder(w).Encode(ReceiptID{Id: uuidString})

		fmt.Fprintf(w, "%d", points)
	})

	r.HandleFunc("/receipts/{id}/points", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		txn := db.Txn(false)
		defer txn.Abort()
		raw, err := txn.First("point", "id", id)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "%d", raw.(*ReceiptPoints).Points)
	})

	http.ListenAndServe(":80", r)
}
