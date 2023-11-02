package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
)

func TestMemDB(t *testing.T) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}

	txn := db.Txn(true)

	receipts := []*ReceiptPoints{
		{"7335f748-ec9d-417b-a335-36ce7a17290f", 30},
		{"e07db9ff-3e21-4db0-98b2-bba2070c3eac", 35},
		{"2b6c8c7f-f74b-4d2c-91a3-d2403a474007", 0},
		{"f76ff8c4-6737-4bbd-8e68-523f7c253af6", 530},
	}
	for _, p := range receipts {
		if err := txn.Insert("point", p); err != nil {
			panic(err)
		}
	}

	txn.Commit()

	txn = db.Txn(false)
	defer txn.Abort()

	_, err = txn.First("point", "id", "e07db9ff-3e21-4db0-98b2-bba2070c3eac")
	if err != nil {
		panic(err)
	}

	_, err = txn.Get("point", "id")
	if err != nil {
		panic(err)
	}
}

func TestCalcReceiptPoints(t *testing.T) {
	receipts := []Receipt{
		{
			Retailer:     "Target",
			PurchaseDate: "2022-01-01",
			PurchaseTime: "13:01",
			Items: []Item{
				{
					ShortDescription: "Mountain Dew 12PK",
					Price:            "6.49",
				},
				{
					ShortDescription: "Emils Cheese Pizza",
					Price:            "12.25",
				},
				{
					ShortDescription: "Knorr Creamy Chicken",
					Price:            "1.26",
				},
				{
					ShortDescription: "Doritos Nacho Cheese",
					Price:            "3.35",
				},
				{
					ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
					Price:            "12.00",
				},
			},
			Total: "35.35",
		},
		{
			Retailer:     "M&M Corner Market",
			PurchaseDate: "2022-03-20",
			PurchaseTime: "14:33",
			Items: []Item{
				{
					ShortDescription: "Gatorade",
					Price:            "2.25",
				},
				{
					ShortDescription: "Gatorade",
					Price:            "2.25",
				},
				{
					ShortDescription: "Gatorade",
					Price:            "2.25",
				},
				{
					ShortDescription: "Gatorade",
					Price:            "2.25",
				},
			},
			Total: "9.00",
		},
	}
	points := []int{
		28,
		109,
	}

	for i, receipt := range receipts {
		msg := fmt.Sprintf("Expected the result to be %d", points[i])
		assert.Equal(t, calcReceiptPoints(receipt), points[i], msg)
	}
}
