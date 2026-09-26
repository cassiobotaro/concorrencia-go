# 🔐 Gorrotina dona do estado

**Também conhecido como:** monitor, confinamento, ator. O último é aproximado. No modelo de atores a mensagem vai para o ator pelo nome, e aqui ela vai por canais. É a mesma diferença entre Erlang e Go comentada na introdução.

O provérbio diz "_Don't communicate by sharing memory, share memory by communicating_", ou seja, não comunique compartilhando memória, compartilhe memória comunicando. Em vez de proteger uma variável com mutex e deixar várias gorrotinas mexerem nela, uma única gorrotina é dona do estado, e as outras pedem alterações e leituras por canais. Não há corrida porque só uma gorrotina toca o dado. A [janela deslizante](../janelas_deslizantes/README.md) já usa essa técnica por dentro. Aqui ela é o assunto principal.

A palestra [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide), de Sameer Ajmani (2013), apresenta a técnica como um laço `for` com `select` e estado local, e a resume assim: a gorrotina serializa o acesso ao próprio estado mutável, sem mutex, sem variável de condição e sem _callback_. É a primeira das três técnicas da palestra. As outras duas, o canal de resposta e o canal `nil`, estão em [parada com confirmação](../cancelamento/README.md#-parada-com-confirmação) e no [fan-in com select](../fan_in/README.md#fan-in-com-uma-gorrotina-e-select).

No exemplo, a gorrotina `contador` é dona de um mapa de contagem por chave. Três gorrotinas enviam mil incrementos cada uma pelo canal `incrementar`, e as leituras usam o canal `consultar`, com o canal de resposta dentro da mensagem, como em [requisição e resposta](../requisicao_resposta/README.md). O `select` atende um pedido por vez. Para encerrar, a função principal fecha `incrementar`.

O exemplo inteiro está em [`dono_do_estado.go`](./dono_do_estado.go) e o teste em [`exemplo_test.go`](./exemplo_test.go).

## E com mutex?

O contraponto também vem de Pike, no provérbio "_Channels orchestrate; mutexes serialize_". Se tudo o que você precisa é serializar o acesso a um contador ou a um mapa, um `sync.Mutex` é mais simples e mais claro. A [versão abaixo](./com_mutex.go) faz isso e produz o mesmo resultado.

Quando a gorrotina dona do estado compensa?

- Quando há regras sobre _como_ o estado muda, como validação, ordem ou eventos.
- Quando ela precisa reagir a vários canais com `select`, como entradas, prazos e cancelamento. É o caso da janela deslizante.
- Quando o estado tem ciclo de vida próprio.

Se nada disso se aplica, use o mutex. Quando as leituras dominam, um `sync.RWMutex` deixa vários `consultar` rodarem juntos, com `RLock`, e só o `incrementar` trava todo mundo.

> **Corrida de dados e condição de corrida.** São duas coisas, e o detector de corrida só pega a primeira. Corrida de dados é duas gorrotinas tocando a mesma variável sem sincronização, com pelo menos uma escrevendo. O `-race`, que o CI liga em todos os testes e exemplos, aponta a linha. Condição de corrida é cada acesso estar protegido e a operação composta não. Imagine um saque escrito com as funções da variante: chama `consultar`, vê saldo 50, e chama uma baixa de 40. Duas gorrotinas fazem isso ao mesmo tempo, as duas veem 50, as duas sacam, e o saldo termina em 30 negativos. Cada chamada travou e destravou o mutex direitinho, então o detector não diz nada. A correção é segurar o mutex durante a operação inteira, da consulta à baixa. Na gorrotina dona do estado isso sai de graça: "saque se houver saldo" vira uma mensagem, e o `select` atende uma por vez.

A variante está em [`com_mutex.go`](./com_mutex.go).
