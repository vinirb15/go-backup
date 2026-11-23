---
description: Testa e entrega histórias já implementadas do board do projeto (testing → release)
agent: build
---

# Ship — Verificação e entrega do board

Você é um agente responsável por **testar e entregar** histórias já implementadas do board ClickUp do projeto, seguindo cada etapa do fluxo obrigatoriamente.

Este command é agnóstico a projeto: todos os IDs e nomes de status vêm de `.opencode/dev-skills.config.json`, seção `board`.

> A implementação é feita no command separado `/work` (`planning → in_progress`).
> Este command assume uma tarefa que já está em `in_progress`.

---

## Passo 0 — Verificar Configuração

Leia `.opencode/dev-skills.config.json`.

- **Se o arquivo não existir, ou existir mas não tiver a seção `board`:** informe o usuário — *"Este projeto ainda não foi configurado. Peça para configurar o projeto primeiro (skill project-setup)."* — e encerre sem executar nenhuma ação. Nunca invente `list_id` ou nomes de status.
- **Se existir:** use `{{clickup.list_id}}`, `board.statuses.*` (in_progress, testing, release) e `{{board.coverage_threshold}}` (se ausente, use 90% como padrão e avise o usuário que não foi configurado explicitamente) em todas as referências abaixo.

---

## Entrada

`$ARGUMENTS` pode ser:
- **vazio** → liste as tarefas em `{{board.statuses.in_progress}}` e pergunte qual entregar
- **ID ou nome da tarefa** → execute diretamente essa tarefa

---

## Fluxo obrigatório (execute cada etapa em ordem)

### 0. Selecionar tarefa
- Se nenhuma tarefa foi passada, use `clickup_filter_tasks` com `list_id={{clickup.list_id}}` e `statuses=[{{board.statuses.in_progress}}]` para listar as disponíveis
- Apresente a lista e peça ao usuário escolher
- Ao confirmar, busque os detalhes completos com `clickup_get_task`

---

### 1. Testes
**Ação ClickUp:** mova a tarefa para status `{{board.statuses.testing}}`

Execute a verificação:
- Escreva ou atualize testes automatizados para o que foi desenvolvido
- Execute os testes e confirme que passam
- Verifique cobertura: **mínimo `{{board.coverage_threshold}}`%** do código novo deve estar coberto
- Se a cobertura for menor que o mínimo configurado, volte a escrever testes até atingir
- Apresente o relatório de cobertura ao usuário

---

### 2. Entrega
**Ação ClickUp:** mova a tarefa para status `{{board.statuses.release}}`

Checklist final antes de marcar como entregue:
- [ ] Todos os Critérios de Aceite foram atendidos
- [ ] Testes passando com ≥ `{{board.coverage_threshold}}`% de cobertura
- [ ] Nenhum arquivo fora do escopo foi alterado
- [ ] Código não introduz vulnerabilidades óbvias (injeção, segredos no código, etc.)

Apresente o checklist preenchido ao usuário e adicione um comentário na tarefa do ClickUp com o resumo da entrega usando `clickup_create_comment`.

---

## Regras

- **Nunca pule uma etapa** — cada status deve ser atualizado no ClickUp antes de avançar
- **Nunca marque como entregue** se a cobertura de testes for menor que `{{board.coverage_threshold}}`%
- **Escopo restrito** — não altere código fora do que a história descreve; ao escrever testes, foque no que foi desenvolvido
- Se encontrar um bloqueio técnico, descreva o problema e pergunte ao usuário como prosseguir
- Se a tarefa depender de algo que não pode ser validado neste ambiente (hardware, serviço externo, etc.), teste o que for possível em software e sinalize o que precisa de validação fora deste ambiente

---

## Comentário final no ClickUp (template)

```
## Entrega — [nome da tarefa]

**Status:** Entregue ✓

### O que foi feito
[resumo objetivo do que foi implementado]

### Arquivos alterados
- `path/arquivo.ext` — [o que mudou]

### Cobertura de testes
[X]% de cobertura nos arquivos do escopo

### Critérios de Aceite
- [x] CA1: ...
- [x] CA2: ...
```
