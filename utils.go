package main

func ExtractOrderIds[T OrderIDer](items []T) []int {
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
