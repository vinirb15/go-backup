---
name: dev-loop
description: Use when asked to run the pipeline, run the dev loop, automate the issue flow, or run the full flow. Triggers on phrases like "rodar o pipeline", "dev loop", "automatizar o fluxo de issues", "rodar o fluxo completo". Runs the full pipeline once, sequentially, as business-analysis → grooming → develop → review-pr-gh. Always invoke this skill to run the pipeline — never run the stages manually.
---

# dev-loop

## Overview

Executa uma passada do pipeline de issues do GitHub a cada invocação. Em sequência, invoca as skills `business-analysis` → `grooming` → `develop` → `review-pr-gh`, cada uma processando a próxima issue disponível para a sua etapa, e encerra ao final.

Esta skill **não agenda**: não usa a skill `loop`, não cria cron, scheduler, launchd nem sleep loops. Cada invocação roda o pipeline uma única vez; para rodar de novo, o usuário invoca a skill novamente.

## Configuração

- **Pré-requisito:** `.opencode/dev-skills.config.json` na raiz do projeto. Se ausente, informe *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* e encerre **sem executar nenhuma etapa**.

## Fluxo

### Passo 1 — Verificar configuração

Leia `.opencode/dev-skills.config.json`. Se ausente, informe a mensagem acima e encerre **sem executar nenhuma etapa**.

### Passo 2 — Executar as etapas

Execute as 4 etapas, em ordem e sequencialmente, invocando cada skill via Skill tool:

1. `business-analysis` — pega a próxima issue disponível para análise de negócio.
2. `grooming` — pega a próxima issue pronta para grooming.
3. `develop` — pega a próxima issue pronta para desenvolvimento.
4. `review-pr-gh` — pega a próxima issue em revisão.

Cada etapa escolhe a issue pelo status correspondente. Se não houver issue naquela etapa, registre **"nada a fazer"** e siga para a próxima.

### Passo 3 — Resumo final

Ao final, emita um resumo curto da execução:

- issue e etapa processadas;
- resultado de cada etapa (ok / nada a fazer / erro);
- pendências (perguntas em aberto, bugs, PRs aguardando).

## Regras

| Regra | Aplicação |
|-------|-----------|
| Uma etapa por vez | Nunca em paralelo: as etapas mexem no mesmo board e repositório |
| Falha não derruba o ciclo | Se uma etapa falhar, registre o erro, pule para a próxima etapa e **não repita** a mesma etapa na mesma execução |
| Limites das skills | Nunca fazer merge de PR nem mover issue além do que cada skill já faz |
| Segredos | Nunca commitar segredos nem `.env` |
| Regras próprias | Cada skill mantém suas regras e confirmações; o dev-loop não as substitui |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Executar sem `dev-skills.config.json` | Rodar `project-setup` primeiro |
| Rodar as etapas em paralelo | Sempre sequencial, na ordem `business-analysis` → `grooming` → `develop` → `review-pr-gh` |
| Repetir a etapa que falhou na mesma execução | Registrar o erro e seguir para a próxima etapa |
| Tentar agendar ou repetir a execução (loop, cron, sleep) | Rodar uma única passada; para repetir, o usuário invoca de novo |
