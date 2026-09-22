package main

func Example() {
	main()

	// Output:
	// Bia 0
	// Bia 1
	// Bia 2
	// gerador: liberando recursos...
	// gerador: context canceled
	// trabalhando...
	// trabalhando...
	// contexto cancelado: context deadline exceeded
	// trabalhando...
	// um dos sinais de parada chegou (aqui, o colega terminou)
}
