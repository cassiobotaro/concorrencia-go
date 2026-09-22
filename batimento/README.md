# 💓 Heartbeat

**Também conhecido como:** sinal de vida, _liveness_.

Um trabalhador que roda por muito tempo pode travar sem que ninguém perceba. Com um _heartbeat_ (batimento), ele emite um sinal em um canal a cada intervalo, e o supervisor usa `select` com timeout para decidir que o trabalhador morreu se o sinal não chegar. Com isso dá para diferenciar um trabalhador que está demorando de um que parou de responder. A forma apresentada aqui segue a do livro _Concurrency in Go_, de Katherine Cox-Buday (O'Reilly, 2017).

Repare em dois detalhes do exemplo. O primeiro é que o batimento é enviado com `select` e `default`. Se ninguém estiver ouvindo, o sinal se perde, e o trabalho nunca fica bloqueado por causa dele. O segundo é que o timeout do supervisor usa `time.After` dentro do laço, recriado a cada volta, como no [tee com timeout](../tee/README.md#tee-com-timeout). Assim, qualquer batimento ou resultado renova o prazo.

No exemplo, o trabalhador leva três intervalos e meio para produzir cada resultado e, de propósito, trava ao produzir o terceiro. O supervisor fica dois intervalos sem notícia e o declara morto. Ao sair, o supervisor cancela o contexto, para que o trabalhador termine caso volte a responder. Esperar por ele não faria sentido, já que um trabalhador travado de verdade pode nunca voltar.

O exemplo inteiro está em [`batimento.go`](./batimento.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
