package main

import (
	"fmt"
	"slices"
	"github.com/kelvins/geocoder"
)

type Item struct {
	Name  string
	Price float32
}

type Store struct {
	Name  string
	Items []Item
}

func listStores(stores []Store) []string {
	storeNames := []string{}
	for _, store := range stores {
		if !slices.Contains(storeNames, store.Name) {
			storeNames = append(storeNames, store.Name)
		}
	}
	return storeNames
}

func cheapest(stores []Store, item string) Store {
	cheapestStore := Store{}
	var price float32 = 10000000
	for _, store := range stores {
		for _, i := range store.Items {
			if i.Name == item {
				if i.Price <= price {
					price = i.Price
					cheapestStore = store
				}
			}
		}
	}
	return cheapestStore
}

func find(stores []Store, item string) []Store {
	storesWith := []Store{}
	for _, store := range stores {
		for _, i := range store.Items {
			if i.Name == item {
				storesWith = append(storesWith, store)
			}
		}
	}
	return storesWith
}

func main() {
	stores := []Store{
		{
			"Kroger",
			[]Item{
				{"Apple", 1.15},
				{"Banana", 1.75},
			},
		},
		{
			"Aldi's",
			[]Item{
				{"Pizza", 5.00},
				{"Lasagna", 4.25},
				{"Banana", 1.00},
			},
		},
		{
			"Target",
			[]Item{
				{"Banana", 0.1},
			},
		},
		{
			"Kroger",
			[]Item{
				{"Banana", 3.00},
			},
		},
	}

	// fmt.Println(stores)
	// for _, store := range stores {
	// 	msg := ""
	// 	msg += store.Name + " has: "
	// 	for _, item := range store.Items {
	// 		msg += item.Name + " for $" + fmt.Sprintf("%.2f", item.Price) + ", "
	// 	}
	// 	fmt.Println(msg[:len(msg)-2])
	// }
	fmt.Println("The stores offered are ", listStores(stores))
	fmt.Println(cheapest(stores, "Banana").Name + " has the cheapest bananas")
	fmt.Println(find(stores, "Banana"))
	address := geocoder.Address{
		Street:  "Central Park West",
		Number:  115,
		City:    "New York",
		State:   "New York",
		Country: "United States",
	}

	location, err := geocoder.Geocoding(address)

	if err != nil {
		fmt.Println("Could not get the location: ", err)
	} else {
		fmt.Println("Latitude: ", location.Latitude)
		fmt.Println("Longitude: ", location.Longitude)
	}

	location = geocoder.Location{
		Latitude:  40.775807,
		Longitude: -73.97632,
	}

	// Convert location (latitude, longitude) to a slice of addresses
	addresses, err := geocoder.GeocodingReverse(location)

	if err != nil {
		fmt.Println("Could not get the addresses: ", err)
	} else {
		// Usually, the first address returned from the API
		// is more detailed, so let's work with it
		address = addresses[0]

		// Print the address formatted by the geocoder package
		fmt.Println(address.FormatAddress())
		// Print the formatted address from the API
		fmt.Println(address.FormattedAddress)
		// Print the type of the address
		fmt.Println(address.Types)
	}
}
