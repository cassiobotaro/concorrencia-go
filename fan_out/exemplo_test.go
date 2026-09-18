package main

// Com mais de um trabalhador, o id que processa cada valor muda a cada
// execução. O exemplo usa um único trabalhador para ter uma saída previsível.
func Example() {
	fanout(sequenciaNumeros(1, 3), 1)

	// Output:
	// id: 1 processando valor: 1
	// id: 1 processando valor: 2
	// id: 1 processando valor: 3
}
