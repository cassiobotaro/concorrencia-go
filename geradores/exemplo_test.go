package main

import "fmt"

func Example() {
	for valor := range sequenciaNumeros(1, 5) {
		fmt.Printf("valor: %v\n", valor)
	}

	// Output:
	// valor: 1
	// valor: 2
	// valor: 3
	// valor: 4
	// valor: 5
}
