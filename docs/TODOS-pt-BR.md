# Pendências

## Funcionalidades e Melhorias

### Concluídas
- [X] Implementar o tratamento de anexos JSON em `HandlerAttachments`, em `internal/core/services/embeds/embeds.go`.
- [X] Corrigir a consulta de incorporações em `Obtain`, em `internal/inbound/interactions/commands/embeds/embeds.go`, para exibir também registros que tenham apenas `Embeds` e nenhum `Content`.
- [X] Padronizar comandos de inbound (`embeds`, `autoreactions`, `server`) com mensagens claras e embeds estéticos finais (`⋆. 𐙚˚࿔ ... 𝜗𝜚˚⋆`).
- [X] Revisar ortografia, pontuação de placeholders e concordância nominal de gênero nas mensagens e status de resposta dos comandos em inbound.
- [X] **Exibição do ID na criação de incorporações:** Alterar `HandlerMessageContent` e `HandlerAttachments` em `internal/core/services/embeds/embeds.go` para retornar `(*domain.Embed, error)` em vez de apenas `error`, permitindo exibir o `ID` gerado na resposta final do comando para o usuário.
- [X] **Diferenciação entre criação e atualização de embeds:** No repositório/serviço de embeds, utilizar `CreateOne` quando `id == 0` e `UpsertOne` quando `id > 0`, garantindo a correta geração de chave autoincrement no SQLite.
- [X] **Listagem de incorporações (`GetAll`):** Expor a consulta de todas as incorporações de um servidor no serviço de embeds e adicionar a opção correspondente no menu de seleção de `/embeds`.
- [X] **Listagem de autorreações (`GetAll`):** Expor a consulta de todas as autorreações de um servidor no serviço de autorreações e adicionar a opção correspondente no menu de seleção de `/autoreactions`.
- [X] **Fluxo de upsert e atualização completa em autorreações:** Ajustar o tratamento de conflito em `CreateOne` para realizar upsert (atualizando emoji em caso de canal já cadastrado) e transformar `UpdateOne` em atualização completa (tipo e emoji) permitindo upsert, alinhando os schemas outbound e os fluxos dos comandos em inbound.

### Pendentes
- [ ] **Cache em memória para autorreações:** Implementar cache local das autorreações ativas para reduzir a frequência de consultas ao banco SQLite no evento `AutoReactionsMessageCreate`.

## Testes

- [ ] Adicionar testes automatizados para o mapeamento de erros SQLite em `internal/outbound/databases/repoerrors` e para os fluxos de sucesso e erro dos serviços de embeds e autoreactions. O projeto ainda não possui arquivos `*_test.go`.
