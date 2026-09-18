package main

import "fmt"

// Com mais de um trabalhador, o id que processa cada valor muda a cada
// execução. O exemplo usa um único trabalhador para que o conjunto de linhas
// seja sempre o mesmo; a ordem entre elas continua variável.
func Example() {
	for resultado := range grupoDeTrabalhadores(sequenciaNumeros(1, 3), 1) {
		fmt.Println(resultado)
	}

	// Unordered output:
	// id: 1 processou valor: 1
	// id: 1 processou valor: 2
	// id: 1 processou valor: 3
	// id: 1 terminou
	// 2
	// 4
	// 6
}
