package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func MockOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders := []models.MockOrder{{
		Id:        1,
		StartDate: "09-03-2026",
		EndDate:   "10-03-2026",
		TotalCost: 100,
		Status:    "Р’С‹РїРѕР»РЅРµРЅ",
		ErrorId:   nil,
	}, {
		Id:        2,
		StartDate: "05-03-2026",
		EndDate:   "07-03-2026",
		TotalCost: 250,
		Status:    "Р’ СЂР°Р±РѕС‚Рµ",
		ErrorId:   nil,
	}, {
		Id:        3,
		StartDate: "01-03-2026",
		EndDate:   "16-03-2026",
		TotalCost: 75,
		Status:    "РЎРѕР·РґР°РЅ",
		ErrorId:   nil,
	}}
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func MockOrderCostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	w.Header().Set("Content-Type", "application/json")
	switch id {
	case "1":
		json.NewEncoder(w).Encode(models.MockOrderCost{
			OrderId:   1,
			Materials: 80.50,
			Labor:     30.00,
			Overhead:  15.75,
			Total:     126.25,
		})
	case "2":
		json.NewEncoder(w).Encode(models.MockOrderCost{
			OrderId:   2,
			Materials: 60.00,
			Labor:     30.00,
			Overhead:  10.50,
			Total:     100.50,
		})
	case "3":
		json.NewEncoder(w).Encode(models.MockOrderCost{
			OrderId:   3,
			Materials: 75.15,
			Labor:     18.85,
			Overhead:  12.30,
			Total:     106.30,
		})
	default:
		json.NewEncoder(w).Encode(models.MockOrderCost{
			OrderId:   4,
			Materials: 120.53,
			Labor:     12.45,
			Overhead:  8.77,
			Total:     141.75,
		})
	}
}
