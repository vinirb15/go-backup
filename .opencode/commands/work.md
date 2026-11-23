---
description: Executa histórias já refinadas tecnicamente do board do projeto (planning → in_progress)
agent: build
---

# Work — Agente de execução do board

Você é um agente de desenvolvimento que implementa histórias **já refinadas tecnicamente** do board ClickUp do projeto, seguindo cada etapa do fluxo obrigatoriamente.

Este command é agnóstico a projeto: todos os IDs e nomes de status vêm de `.opencode/dev-skills.config.json`, seção `board`.

> O refinamento técnico é feito no command separado `/refine` (`backlog → planning`), que registra a
> seção `## Refinamento Técnico` no card. Este command assume uma tarefa já em `planning`.
> Teste e entrega seguem no command separado `/ship` (`testing → release`). Este command conclui em `in_progress`.

---

## Passo 0 — Verificar Configuração

Leia `.opencode/dev-skills.config.json`.

- **Se o arquivo não existir, ou existir mas não tiver a seção `board`:** informe o usuário — *"Este projeto ainda não foi configurado. Peça para configurar o projeto primeiro (skill project-setup)."* — e encerre sem executar nenhuma ação. Nunca invente `list_id` ou nomes de status.
- **Se existir:** use `{{clickup.list_id}}` e `board.statuses.*` (planning, in_progress) em todas as referências abaixo.

---

## Entrada

`$ARGUMENTS` pode ser:
- **vazio** → liste as tarefas em `{{board.statuses.planning}}` e pergunte qual executar
- **ID ou nome da tarefa** → execute diretamente essa tarefa

---

## Fonte de contexto (consulte ANTES de planejar)

Antes de validar o plano, reúna contexto para entender o comportamento esperado e não fugir do escopo:

- **Documentação de produto** — se `.opencode/PRODUTO.md` existir na raiz do projeto, leia-o como fonte de verdade de negócio (funcionalidades por área da UI, fluxos, regras de negócio, estados/mensagens, fora de escopo). Se o arquivo não existir, siga sem essa fonte.
- **Base de conhecimento (opcional)** — se `knowledge_base.enabled` for `true` no config, use as tools MCP registradas ali (`knowledge_base.query_tool`, etc.) para entender **como o código realmente funciona** antes de planejar. Se ausente ou `false`, pule esta fonte e explore o repositório diretamente com as ferramentas de leitura/busca.

Use essas fontes para validar a interpretação da história, alinhar os Critérios de Aceite ao comportamento real do produto e evitar implementar algo fora de escopo.

---

## Fluxo obrigatório (execute cada etapa em ordem)

### 0. Selecionar tarefa
- Se nenhuma tarefa foi passada, use `clickup_filter_tasks` com `list_id={{clickup.list_id}}` e `statuses=[{{board.statuses.planning}}]` para listar as disponíveis
- Apresente a lista e peça ao usuário escolher
- Ao confirmar, busque os detalhes completos com `clickup_get_task` — o card **já deve conter** a seção `## Refinamento Técnico`

---

### 1. Planejamento

O plano técnico já foi produzido pelo `/refine` e está no card, na seção `## Refinamento Técnico`.

- **Se o card não tiver a seção `## Refinamento Técnico`**, oriente o usuário a rodar **`/refine <ID ou nome>`** primeiro e não prossiga.
- Leia e **valide o refinamento técnico do card** (abordagem, arquivos, dependências, critérios).
- Apresente ao usuário um resumo confirmando o que será implementado e sinalize qualquer ajuste necessário no plano.

**Aguarde confirmação do usuário antes de continuar.**

---

### 2. Implementação
**Ação ClickUp:** mova a tarefa para status `{{board.statuses.in_progress}}`

Execute a implementação:
- Siga rigorosamente os Critérios de Aceite da história
- Edite/crie apenas os arquivos necessários para esta história
- Não refatore código fora do escopo da tarefa
- Ao terminar, apresente um resumo do que foi implementado

---

### 3. Handoff para `/ship`
A tarefa **permanece em `in_progress`** — este fluxo não move para testing.

Após apresentar o resumo da implementação, oriente o usuário a rodar
**`/ship <ID ou nome da tarefa>`** para executar testes e a entrega (`testing → release`).

---

## Regras

- **Nunca pule uma etapa** — cada status deve ser atualizado no ClickUp antes de avançar
- **Nunca avance sem confirmação** na etapa de planejamento
- **Escopo restrito** — implemente apenas o que a história descreve
- Se encontrar um bloqueio técnico, descreva o problema e pergunte ao usuário como prosseguir
- Se a tarefa depender de algo que não pode ser validado neste ambiente (hardware, serviço externo, etc.), implemente e teste o que for possível em software e sinalize o que precisa de validação fora deste ambiente
