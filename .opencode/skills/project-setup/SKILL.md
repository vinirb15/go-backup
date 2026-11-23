---
name: project-setup
description: Use when setting up dev-skills for a new project for the first time, or to reconfigure GitHub/repo settings — interactively links this project's GitHub repository, maps pipeline stages/priorities/task types to GitHub issue labels, detects the repo layout (including test/lint/build commands and ORM), and configures the work/story commands' board into .opencode/dev-skills.config.json. Trigger phrases: "configurar o projeto", "vincular o github", "setup do dev-skills", "primeira configuração", "reconfigurar o projeto". Run this before using business-analysis, create-issue, grooming, develop, review-pr-gh, review-pr-git, or the work/story commands for the first time in a new project.
---

# project-setup

## Overview

Todas as outras skills deste pacote (`business-analysis`, `create-issue`, `grooming`, `develop`, `review-pr-gh`, `review-pr-git`), além dos commands `work` e `story` em `.opencode/commands/`, são agnósticos a projeto: eles não têm nenhum repositório, nome de label, caminho de pacote ou comando de teste/lint/build hardcoded. Em vez disso, leem `.opencode/dev-skills.config.json` na raiz do projeto.

O gerenciador de tarefas é o **GitHub Issues**, operado via `gh` CLI. Como issues do GitHub só têm os estados `open`/`closed`, o pipeline de estágios, as prioridades e os tipos de task são representados por **labels**.

Esta skill constrói (ou atualiza) esse arquivo de forma interativa — buscando dados reais no GitHub e no repositório sempre que possível, em vez de pedir ao usuário para digitar valores às cegas. Ao final, todas as skills e os commands `work`/`story` estão prontos para uso, sem nenhuma configuração manual adicional.

## Arquivo de Configuração

**Caminho:** `.opencode/dev-skills.config.json` (raiz do projeto — deve ser commitado, é config de equipe, não segredo).

**Schema:**

```json
{
  "version": 1,
  "tracker": "github",
  "project_name": "<nome do projeto>",
  "github": {
    "repo": "<owner>/<nome> — repositório onde as issues vivem"
  },
  "statuses": {
    "backlog": "<label real>",
    "scoping": "<label real>",
    "ready_to_grooming": "<label real>",
    "in_grooming": "<label real>",
    "ready_to_develop": "<label real>",
    "in_develop": "<label real>",
    "blocked": "<label real>",
    "in_review": "<label real>",
    "testing": "<label real>",
    "homologating": "<label real>",
    "waiting_for_deploy": "<label real>",
    "deployed": "<label real>",
    "cancelled": "<label real>"
  },
  "priorities": { "urgent": "<label>", "high": "<label>", "normal": "<label>", "low": "<label>" },
  "task_types": { "bug": "<label>", "task": "<label>" },
  "repo": {
    "default_branch": "<branch>",
    "package_manager": "<pnpm|npm|yarn|bun|N/A>",
    "monorepo": true,
    "packages": { "backend": "<caminho>", "frontend": "<caminho>", "shared": "<caminho>" },
    "test_command": "<comando para rodar os testes, ex.: 'go test ./...', 'pnpm test', 'pytest'>",
    "lint_command": "<comando de lint, ex.: 'go vet ./...', 'pnpm lint', 'ruff check .'>",
    "build_command": "<comando de build, ex.: 'go build ./...', 'pnpm build'>",
    "coverage_config_path": "<caminho do arquivo que define o gate de cobertura (package.json, jest.config.*, .coveragerc, etc.) ou null se o projeto não tiver gate de cobertura configurado>",
    "orm": "<nome do ORM/ferramenta de migration do projeto (ex.: TypeORM, Prisma, Alembic, golang-migrate) ou 'none' se o projeto não usa banco relacional/migrations>"
  },
  "board": {
    "repo": "<owner>/<nome> do repositório usado pelos commands work/story — pode ser o mesmo de github.repo ou outro>",
    "statuses": {
      "backlog": "<label real>",
      "planning": "<label real>",
      "in_progress": "<label real>",
      "testing": "<label real>",
      "release": "<label real>"
    },
    "coverage_threshold": 90
  }
}
```

Chaves `task_types`, `priorities`, `repo.packages.frontend`/`repo.packages.shared`, `repo.coverage_config_path`, e a seção `board` inteira podem ficar **ausentes** se não fizerem sentido para o projeto (ex.: projeto sem separação de frontend, sem labels de tipo, sem gate de cobertura configurado, ou que não usa os commands `work`/`story`). As demais skills e os commands tratam a ausência como "pule esta parte" — nunca inventam um valor.

Um exemplo preenchido está em `.opencode/dev-skills.config.example.json` (neste repositório de skills) para referência.

---

## Fluxo Obrigatório

```
[início] → Verificar config existente (manter / reconfigurar tudo / reconfigurar uma seção)
        → Verificar gh CLI + escolher o repositório GitHub
        → Mapear os 13 estágios semânticos para labels reais (criando os que faltarem, com confirmação)
        → Mapear prioridades para labels
        → Mapear task_type para labels (ou marcar ausente)
        → Detectar layout do repositório (branch, package manager, pacotes, test/lint/build, ORM)
        → (opcional) Configurar o board dos commands work/story
        → Confirmar e gravar .opencode/dev-skills.config.json
        → (opcional) Smoke test de leitura no GitHub
[fim]
```

---

## Passo 0 — Verificar Configuração Existente

Tente ler `.opencode/dev-skills.config.json`.

- **Se não existir:** siga para o Passo 1 (setup do zero).
- **Se existir com `tracker` diferente de `"github"` (ex.: config da versão ClickUp, com seção `clickup`):** avise o usuário que o config é de outro gerenciador e que as skills desta versão não o reconhecem; ofereça reconfigurar do zero.
- **Se existir:** mostre os valores atuais agrupados por seção (GitHub / Status / Prioridades / Task Types / Repositório / Board) e pergunte ao usuário:
  1. Manter como está (encerrar sem alterações)
  2. Reconfigurar tudo do zero
  3. Reconfigurar apenas uma seção específica

**Nunca sobrescreva silenciosamente.** Se a opção for (3), execute apenas o(s) passo(s) relevante(s) abaixo e mescle o resultado com o restante do arquivo existente na gravação final (Passo 8).

---

## Passo 1 — Verificar o `gh` e Escolher o Repositório

1. Rode `gh auth status`.
   - **Se o `gh` não estiver instalado ou autenticado:** oriente o usuário a instalar/autenticar (sugira digitar `! gh auth login` no prompt) e encerre — todas as skills dependem do `gh`.
2. Detecte o repositório atual com `gh repo view --json nameWithOwner,hasIssuesEnabled,defaultBranchRef`.
3. Pergunte se as issues vivem nesse repositório ou em outro (ex.: um repositório central de tracking). Se for outro, peça o `owner/nome` e valide com `gh repo view <owner/nome> --json nameWithOwner,hasIssuesEnabled`.
4. Se `hasIssuesEnabled` for `false`, avise o usuário — as issues precisam estar habilitadas nas configurações do repositório antes de seguir.

Registre `github.repo`.

---

## Passo 2 — Mapear os Estágios do Pipeline para Labels

Busque os labels existentes no repositório:
```bash
gh label list -R <github.repo> --json name,description,color --limit 500
```

Para cada um dos 13 estágios semânticos (`backlog`, `scoping`, `ready_to_grooming`, `in_grooming`, `ready_to_develop`, `in_develop`, `blocked`, `in_review`, `testing`, `homologating`, `waiting_for_deploy`, `deployed`, `cancelled`):

1. Sugira uma correspondência (match exato ou aproximado por nome) com os labels existentes. Se não houver, sugira o padrão `status: <estágio>` (ex.: `status: ready to grooming`).
2. Peça confirmação explícita do usuário — **nunca grave no config um label que não exista no repositório.**
3. Para os labels que ainda não existem, liste todos de uma vez e pergunte se pode criá-los. Só após a confirmação, crie cada um:
   ```bash
   gh label create "<label>" -R <github.repo> --color <cor> --description "Pipeline dev-skills: <estágio>"
   ```
   Use uma mesma cor para todos os labels de status, para que fiquem visualmente agrupados.
4. Se o usuário não quiser usar algum estágio, ofereça, sempre com aviso visível:
   - mapear mais de um estágio semântico para o mesmo label, ou
   - marcar o estágio como não utilizado.

> Os estágios `deployed` e `cancelled` também **fecham** a issue (`gh issue close --reason completed` / `--reason "not planned"`) — o label registra o estágio final e o fechamento tira a issue da lista de abertas. Explique isso ao usuário ao mapeá-los.

Registre o resultado em `statuses.*`.

---

## Passo 3 — Mapear Prioridades

A partir da mesma lista de labels do Passo 2, procure labels de prioridade já existentes (ex.: `priority: high`, `P1`, `urgente`). Sugira a correspondência para `urgent`, `high`, `normal` e `low`; se não existirem, sugira o padrão `priority: <nível>` e crie-os seguindo a mesma regra de confirmação do Passo 2.

Registre em `priorities.*`.

---

## Passo 4 — Mapear Task Type

Pergunte ao usuário quais labels distinguem Bug de outros tipos de trabalho. Sugira os labels padrão do GitHub quando existirem no repositório: `bug` para bug e `enhancement` para task/feature.

- Se o projeto **não usa** labels de tipo: **não crie** a chave `task_types` no config. A skill `create-issue` deve então omitir o label de tipo ao criar issues.

Registre em `task_types.*` (ou omita).

---

## Passo 5 — Detectar Layout do Repositório

Faça uma sondagem **somente leitura** na raiz do projeto (não deste repositório de skills, e sim do projeto onde o config está sendo criado):

- **Estrutura:** verifique se existem diretórios `backend/`, `frontend/`, `shared/` (ou `apps/*`, `packages/*`), ou se é um pacote único.
- **Gerenciador de pacotes:** identifique pelo lockfile — `pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm, `bun.lockb` → bun; se o projeto não for Node.js (ex.: Go, Python, Rust), registre `"N/A"`.
- **Branch padrão:** use o `defaultBranchRef` obtido no Passo 1 (ou `git symbolic-ref refs/remotes/origin/HEAD`); se não for possível detectar, pergunte ao usuário.
- **Comandos de teste/lint/build:** procure scripts já definidos no projeto (`package.json` → `scripts.test`/`lint`/`build`; `Makefile`; `justfile`; convenção da linguagem — `go test ./...`, `pytest`, `cargo test`, etc.) e proponha como sugestão, em vez de perguntar às cegas.
- **Gate de cobertura:** identifique se existe um arquivo de config de cobertura (`jest.config.*`/`coverageThreshold` em `package.json`, `.coveragerc`, `pyproject.toml` com `[tool.coverage]`, etc.) e registre o caminho. Se não houver nenhum, registre `null` — não invente um piso de cobertura.
- **ORM / ferramenta de migration:** identifique pela stack (`typeorm`/`prisma` em `package.json`, `alembic` em Python, `golang-migrate`/`goose` em Go, etc.). Se o projeto não usa banco relacional com migrations, registre `"none"`.

Apresente o que foi detectado para confirmação (*"Detectei monorepo pnpm com backend/, frontend/, shared/, branch padrão main, testes via `pnpm test`, cobertura configurada em `backend/jest.config.js`, ORM TypeORM. Confirma?"*) em vez de perguntar às cegas. Se algo estiver errado ou ausente, peça o valor correto ao usuário.

Registre em `repo.*`.

---

## Passo 6 — Configurar o Board dos Commands `work`/`story` (Opcional)

Os commands `.opencode/commands/work.md` e `.opencode/commands/story.md` implementam um fluxo mais simples e direto (5 estágios: backlog → planning → in_progress → testing → release), separado do pipeline de 13 estágios das skills. Pergunte ao usuário se o projeto usa (ou vai usar) esses commands.

- **Se sim:**
  1. Pergunte se o repositório é o mesmo já escolhido no Passo 1 (`github.repo`) ou outro — se outro, valide-o como no Passo 1.
  2. Mapeie os 5 estágios semânticos (`backlog`, `planning`, `in_progress`, `testing`, `release`) contra os labels desse repositório, seguindo as mesmas regras do Passo 2 (nunca gravar label inexistente, criar só com confirmação). Sugira o padrão `board: <estágio>` para não colidir com os labels do pipeline das skills.
  3. Pergunte o percentual mínimo de cobertura de testes exigido para considerar uma entrega pronta (sugira 90% como padrão, mas confirme).
  4. Registre em `board.repo`, `board.statuses.*` e `board.coverage_threshold`.
- **Se não:** não crie a chave `board` no config. Os commands `work`/`story` devem então informar o usuário para rodar esta skill antes de serem usados.

---

## Passo 7 — Confirmar e Gravar

Monte o JSON completo (mesclando com o config existente, se esta foi uma reconfiguração parcial) e mostre-o ao usuário para revisão final. Após confirmação, grave em `.opencode/dev-skills.config.json`.

Lembre o usuário de commitar o arquivo para que toda a equipe compartilhe o mesmo mapeamento.

---

## Passo 8 — Smoke Test (Opcional)

Ofereça rodar uma chamada somente leitura usando o repositório e o label de backlog recém-configurados, para confirmar que a configuração resolve corretamente antes de encerrar:
```bash
gh issue list -R <github.repo> --state open --label "<statuses.backlog>" --limit 5
```

---

## Princípios

| Princípio | Aplicação |
|-----------|-----------|
| Nunca inventar | Todo repositório e label vem de uma escolha real do usuário ou de dados buscados ao vivo no GitHub/repositório |
| Confirmar, não perguntar às cegas | Sempre que possível, detecte/busque e peça confirmação, em vez de pedir para o usuário digitar de memória |
| Criar só com confirmação | Labels novos são criados no repositório apenas depois que o usuário aprova a lista |
| Nunca sobrescrever silenciosamente | Config existente é sempre mostrado e a decisão de reconfigurar é do usuário |
| Ausência é uma opção válida | Se o projeto não usa uma seção (task_type, frontend, etc.), a chave fica ausente — as demais skills tratam isso como "pule" |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Sobrescrever config existente sem perguntar | Sempre mostrar o que existe e oferecer manter/reconfigurar/reconfigurar parcial |
| Gravar no config um label que não existe no repositório | Listar os labels reais com `gh label list` e criar os que faltam (com confirmação) antes de gravar |
| Criar labels sem perguntar | Mostrar a lista de labels a criar e só criar após aprovação |
| Perguntar repo/branch/gerenciador às cegas quando dá para detectar | Buscar via `gh`/git primeiro, pedir confirmação depois |
| Seguir sem `gh` autenticado | Verificar `gh auth status` no Passo 1 e orientar `gh auth login` |
| Gravar o arquivo sem revisão final do usuário | Sempre mostrar o JSON montado antes de gravar |
| Assumir TypeORM/Jest/pnpm como padrão para qualquer projeto | Detectar a stack real (Passo 5) e registrar em `repo.test_command`/`lint_command`/`build_command`/`orm`; nunca assumir um ecossistema fixo |
| Pular a configuração do `board` sem perguntar | Sempre perguntar no Passo 6 se o projeto usa os commands `work`/`story`, mesmo que a resposta seja "não" |
