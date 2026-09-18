package main

import "fmt"

// O main usa esperas longas e seu resultado depende do relógio. Aqui a janela
// é exercitada sem relógio: cinco valores entram enquanto ninguém lê a saída,
// então a janela de tamanho 3 precisa descartar os dois mais antigos.
func Example() {
	entrada := make(chan int)
	saida := make(chan int)
	go janelaDeslizante(entrada, saida, 3)

	for i := 1; i <= 5; i++ {
		entrada <- i
	}
	close(entrada)

	for valor := range saida {
		fmt.Println("recebeu", valor)
	}

	// Unordered output:
	// Janela Deslizante: Buffer cheio, descartou 1 para adicionar 4.
	// Janela Deslizante: Buffer cheio, descartou 2 para adicionar 5.
	// Janela Deslizante: Enviou 3 para o consumidor.
	// Janela Deslizante: Enviou 4 para o consumidor.
	// Janela Deslizante: Enviou 5 para o consumidor.
	// recebeu 3
	// recebeu 4
	// recebeu 5
}
