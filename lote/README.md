# 📦 Processamento em lote (batch processing)

**Também conhecido como:** _batching_, _micro-batching_.

Um processamento em lote (_batch processing_) agrupa itens que chegam um por um, para que o consumidor os processe em blocos. Um canal de descarga força o envio do lote antes de ele encher, e um canal de conclusão avisa quando o último lote foi processado.

Na prática: em vez de gravar cada item no banco assim que ele chega, o programa junta 100 itens, ou 100ms de itens, e grava tudo em uma requisição só.

No exemplo, o lote tem capacidade para três itens. Quando o terceiro chega, o lote enche e segue para o canal de saída.

Um lote pode ser descarregado de três formas:

- Quando ele enche.
- Quando o `intervalo` passa sem que ele tenha enchido. Isso é feito com um `time.Ticker` dentro do `select`, e evita que um item fique esperando companhia por tempo indeterminado.
- Sob demanda, pelo canal `descarga`.

No exemplo, o item 6 é enviado sozinho e sai pelo intervalo de 100ms.

Repare que, depois de enviar um lote, o código cria um novo _slice_ em vez de reaproveitar o anterior com `buf[:0]`. O consumidor pode ainda estar lendo o lote enviado, e reutilizar o mesmo _array_ de apoio sobrescreveria dados em uso. Parece uma otimização óbvia, mas quebraria o programa.

Se a entrada for fechada com itens no buffer, o lote parcial ainda é enviado.

O exemplo inteiro está em [`lote.go`](./lote.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).
