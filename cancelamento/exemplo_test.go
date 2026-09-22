package main

func Example() {
	main()

	// Output:
	// valor: 1
	// valor: 2
	// valor: 3
	// goroutines presas: 1
	// valor: 1
	// valor: 2
	// valor: 3
	// gerador: cancelado, encerrando
	// gerador cancelável encerrado
	// Duda 0
	// Duda 1
	// Duda 2
	// gerador: liberando recursos...
	// gerador: parei
	// Bia 0
	// Bia 1
	// Bia 2
	// gerador: liberando recursos...
	// gerador: context canceled
	// trabalhando...
	// trabalhando...
	// um dos sinais de parada chegou (aqui, o prazo de 250ms)
}
