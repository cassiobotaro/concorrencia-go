# ⛓️ Daisy-chain

**Também conhecido como:** corrente de gorrotinas, telefone sem fio.

Gorrotinas são baratas, e é comum ter dezenas de milhares delas. Este exemplo, tirado da palestra [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide), liga 10 mil gorrotinas em uma corrente, cada uma somando 1 ao valor que recebe da vizinha da direita e passando o resultado para a esquerda. O valor 1 entra por uma ponta e sai 10001 pela outra.

Ninguém usa isso no dia a dia. O exemplo serve para mostrar que dividir o trabalho em pedaços bem pequenos não custa caro. Criar 10 mil _threads_ do sistema operacional para somar 1 seria impensável. Cada uma nasce com uma pilha de alguns megabytes. Uma gorrotina começa com cerca de 2 KB de pilha, que cresce só quando precisa, então as 10 mil ficam na casa de 20 MB. Com gorrotinas, o programa termina em uma fração de segundo.

O exemplo inteiro está em [`corrente.go`](./corrente.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
