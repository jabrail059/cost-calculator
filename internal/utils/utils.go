package utils

import "gitverse.ru/topit/12-40_team20_Zueva/internal/models"

func ExtractOrderIds[T models.OrderIDer](items []T) []int {
	m := make(map[int]bool)
	for _, item := range items {
		id := item.GetOrderId()
		m[id] = true
	}

	keyIds := make([]int, 0)
	for key := range m {
		keyIds = append(keyIds, key)
	}
	return keyIds
}
