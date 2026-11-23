---
name: create-issue
description: Use when creating a task or bug report as a GitHub issue — covers feature requests, technical tasks, bugs, and pending items identified during development.
---

# create-issue

## Overview

Cria uma issue no GitHub no repositório do projeto com o label de status inicial de backlog, usando o template de issue padrão do projeto.

## Configuração do Projeto

Esta skill é agnóstica a projeto. Antes de qualquer chamada ao GitHub, leia `.opencode/dev-skills.config.json` na raiz do projeto.

> **Se o arquivo não existir, ou não tiver a seção `github`:** informe o usuário — *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* — e encerre sem executar nenhuma ação. Nunca invente repositório, labels de status ou de tipo.

| Referência nesta skill | Chave no config |
|---|---|
| Repositório das issues | `github.repo` |
| Label de status inicial | `statuses.backlog` |

Todas as operações usam a `gh` CLI, sempre com `-R {{github.repo}}`. Se o `gh` não estiver autenticado, oriente o usuário a rodar `gh auth login` e encerre.

## Fluxo de Status

Este é o pipeline completo de estágios que uma issue percorre ao longo das demais skills (business-analysis, grooming, develop, review). Cada estágio é um **label** na issue — a issue tem sempre exatamente um label de status. Os nomes abaixo são **semânticos** — o label literal de cada estágio vem de `statuses.<estágio>` no config (ver tabela de resolução):

```
backlog → scoping → ready_to_grooming → in_grooming
→ ready_to_develop → in_develop → blocked? → in_review
→ testing → homologating → waiting_for_deploy → deployed
                                                  cancelled (qualquer ponto)
```

**Tabela de resolução (chave do estágio → chave no config):**

| Estágio semântico | Chave no config | Estado da issue |
|---|---|---|
| Backlog | `statuses.backlog` | aberta |
| Scoping | `statuses.scoping` | aberta |
| Ready To Grooming | `statuses.ready_to_grooming` | aberta |
| In Grooming | `statuses.in_grooming` | aberta |
| Ready To Develop | `statuses.ready_to_develop` | aberta |
| In Develop | `statuses.in_develop` | aberta |
| Blocked | `statuses.blocked` | aberta |
| In Review | `statuses.in_review` | aberta |
| Testing | `statuses.testing` | aberta |
| Homologating | `statuses.homologating` | aberta |
| Waiting For Deploy | `statuses.waiting_for_deploy` | aberta |
| Deployed | `statuses.deployed` | fechada (`gh issue close --reason completed`) |
| Cancelled | `statuses.cancelled` | fechada (`gh issue close --reason "not planned"`) |

**Mudar de estágio** = trocar o label de status em uma única chamada:
```bash
gh issue edit <número> -R {{github.repo}} --remove-label "<label de status atual>" --add-label "<novo label de status>"
```

## Tipo da Task

| Situação | Chave no config |
|----------|-----------|
| Bug, erro, comportamento inesperado | `task_types.bug` |
| Feature, melhoria, pendência técnica | `task_types.task` |

> Se o projeto não configurou `task_types`, não adicione label de tipo na criação da issue.

## Passo 0 — Detectar se é um Épico

**Antes de criar a issue**, analise o texto recebido para identificar sinais de épico:

| Sinal | Exemplos |
|-------|---------|
| Escopo amplo | "módulo completo de X", "sistema de Y", "fluxo inteiro de Z" |
| Múltiplas funcionalidades | conjunções como "e também", "além disso", "incluindo", "com X e Y e Z" |
| Múltiplos domínios | frontend + backend + banco + integração na mesma descrição |
| Longa duração implícita | "fase 1", "v2", "reformulação", "migração completa" |
| Múltiplos tipos de usuário | "admin pode... usuário pode... empresa pode..." |

**Se épico detectado:**
1. Alertar: *"Isso parece ser um épico — uma tarefa grande com múltiplas entregas."*
2. Perguntar: *"Deseja (1) criar como épico com sub-issues, ou (2) criar como issue única?"*
   - **Se épico (opção 1):** criar a issue mãe com título descritivo; listar sub-issues sugeridas e perguntar se deve criá-las; se sim:
     1. criar cada uma com o label `{{task_types.task}}` (se configurado) e `{{statuses.backlog}}`, incluindo no corpo a linha `Parte de #<número da issue mãe>`;
     2. adicionar à descrição da issue mãe uma seção `## Sub-issues` com uma task list (`- [ ] #<número>` por sub-issue) — o GitHub renderiza o progresso a partir dela.
   - **Se issue única (opção 2):** continuar fluxo normal abaixo

**Se não é épico:** continuar fluxo normal.

---

## Como Criar

1. Escreva a descrição (seguindo `ISSUE_TEMPLATE.md`) em um arquivo temporário — nunca passe corpos longos inline via `--body`, para evitar problemas de escape.
2. Chame:
   ```bash
   gh issue create -R {{github.repo}} \
     --title "<título claro e objetivo>" \
     --body-file <arquivo temporário> \
     --label "{{statuses.backlog}}" \
     --label "{{task_types.task}} | {{task_types.bug}}"
   ```
   Omita o segundo `--label` se o projeto não configurou `task_types`. Se a prioridade for conhecida e o projeto tiver `priorities`, adicione também `--label "{{priorities.<nível>}}"`.
3. Retorne ao usuário o número e a URL da issue criada (o `gh issue create` imprime a URL).

## Template de Descrição

Use o template em `ISSUE_TEMPLATE.md` (mesmo diretório desta skill). Todo conteúdo usa **lista alternada** — campos e itens alternam entre `-` e `1.`, sem blocos de prosa:

- **Task / Feature** — Contexto, Objetivo, Critérios de Aceitação, Notas Técnicas
- **Bug** — Descrição do Bug, Comportamento Atual, Comportamento Esperado, Passos para Reproduzir, Arquivo/Linha Afetado, Notas Adicionais

## Checklist Antes de Criar

- [ ] O título descreve claramente o problema ou objetivo?
- [ ] O texto foi avaliado para épico (Passo 0)?
- [ ] O label de tipo está correto (`task_types.bug` ou `task_types.task`), ou omitido se o projeto não usa labels de tipo?
- [ ] A descrição segue o template em lista alternada (não está vazia ou genérica)?
- [ ] O repositório é o valor de `github.repo` no config?
- [ ] O label de status é `statuses.backlog` (e só ele)?

## Erros Comuns

| Erro | Correção |
|------|----------|
| Criar no repositório errado | Sempre passe `-R {{github.repo}}` do config |
| Omitir critérios de aceitação | Preencha ao menos 1 critério |
| Usar label de status diferente do backlog | Novas issues sempre entram com `statuses.backlog` |
| Criar sem nenhum label de status | A issue sumiria do pipeline — sempre adicionar `statuses.backlog` |
| Título vago ("fix bug", "ajuste") | Seja específico: "Corrigir erro 401 no endpoint /auth/refresh" |
