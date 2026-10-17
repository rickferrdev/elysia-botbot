 <div align="center">

  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:07130D,50:123D2A,100:2EA66B&height=180&section=header&text=Elysia-BotBot&fontSize=52&fontColor=E8FFF0&fontAlignY=38&desc=Discord%20Bot%20%E2%80%A2%20Built%20with%20Go&descSize=17&descAlignY=60&animation=fadeIn" width="100%" alt="ElysiaBotBot — Discord Bot built with Go" />

  <a href="https://github.com/rickferrdev/elysia-botbot">
    <img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=500&size=16&duration=2800&pause=900&color=52D991&center=true&vCenter=true&width=600&lines=Building+more+than+a+Discord+bot.;Learning+Go+through+real+software.;Clean+code.+Thoughtful+architecture.;One+commit+at+a+time." alt="Mensagens animadas sobre o projeto" />
  </a>

  <p>
    <strong>Um projeto de estudo com mentalidade de produto.</strong><br/>
    Explorando arquitetura backend, boas práticas e desenvolvimento em Go.
  </p>

  <p>
    <a href="#-funcionalidades"><img src="https://img.shields.io/badge/Features-Explore-239B56?style=for-the-badge" alt="Funcionalidades"/></a>
    <a href="#-tecnologias"><img src="https://img.shields.io/badge/Stack-Go%20%7C%20Redis%20%7C%20SQLite-183D2B?style=for-the-badge" alt="Tecnologias"/></a>
    <a href="#-começando"><img src="https://img.shields.io/badge/Getting_Started-Quickstart-276749?style=for-the-badge" alt="Começando"/></a>
  </p>

  <p>
    <img src="https://img.shields.io/badge/Go-Backend-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go"/>
    <img src="https://img.shields.io/badge/Discord-Bot-5865F2?style=flat-square&logo=discord&logoColor=white" alt="Discord"/>
    <img src="https://img.shields.io/badge/Redis-Cache-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis"/>
    <img src="https://img.shields.io/badge/SQLite-Database-003B57?style=flat-square&logo=sqlite&logoColor=white" alt="SQLite"/>
  </p>

</div>

---


# Elysia-BotBot

Um bot para Discord desenvolvido em Go, criado como um projeto de estudo com foco em arquitetura, organização de código e boas práticas de desenvolvimento de software.

O **ElysiaBotBot** reúne funcionalidades para servidores Discord em uma aplicação estruturada para facilitar a manutenção, a evolução e a compreensão do código. Mais do que construir um bot, a proposta é explorar como desenvolver e organizar uma aplicação backend em Go utilizando ferramentas e padrões presentes em projetos reais.

> [!NOTE]
> O projeto é mantido com objetivos educacionais e de aprendizado prático. A versão `1.0.0` representa a conclusão de uma etapa de estabilização, correções e refatorações.

## Objetivos

- Desenvolver um bot para Discord utilizando Go.
- Explorar separação de responsabilidades e organização por camadas.
- Praticar injeção de dependências e composição de componentes.
- Trabalhar com persistência de dados e cache.
- Construir uma base de código compreensível, testável e preparada para evoluções futuras.
- Consolidar conhecimentos por meio de um projeto prático, em vez de exemplos isolados.

## Funcionalidades

O projeto reúne funcionalidades voltadas à configuração e à personalização de servidores Discord.

- [x] Estrutura para comandos e eventos do Discord.
- [x] Configurações específicas por servidor.
- [x] Sistema de autorreações.
- [x] Suporte a embeds.
- [x] Persistência de dados com SQLite.
- [x] Integração com Redis para operações de cache.

A disponibilidade e o comportamento de cada funcionalidade dependem da implementação presente na versão publicada.

## Tecnologias

| Tecnologia | Utilização |
| --- | --- |
| [Go](https://go.dev/) | Linguagem principal |
| [discord.go](https://github.com/diamondburned/arikawa) | Integração com a API do Discord |
| [SQLite](https://www.sqlite.org/) | Persistência de dados |
| [Bun](https://bun.uptrace.dev/) | Acesso ao banco de dados |
| [Redis](https://redis.io/) | Cache |
| [Uber Fx](https://github.com/uber-go/fx) | Injeção de dependências e ciclo de vida da aplicação |
| [Docker](https://www.docker.com/) | Empacotamento e execução em contêineres |

## Começando

### Pré-requisitos

Para executar o projeto localmente, você precisará de:

- [Go](https://go.dev/dl/)
- Uma aplicação e um bot configurados no [Discord Developer Portal](https://discord.com/developers/applications)
- [Redis](https://redis.io/) em execução
- Docker, caso prefira executar a aplicação em um contêiner

### Configuração

Clone o repositório e acesse a pasta do projeto:

```bash
git clone https://github.com/rickferrdev/elysia-botbot.git
cd elysia-botbot
```

Configure as variáveis de ambiente necessárias para a aplicação, incluindo o token do bot, a conexão com o banco de dados e as credenciais de acesso ao Redis.

Use o arquivo `.env.example` como referência, conferindo os nomes das variáveis com a configuração da versão que você pretende executar. **Não compartilhe seu token do Discord nem envie credenciais para o repositório.**

### Executando localmente

Com as dependências configuradas, execute:

```bash
go mod download
go run ./cmd/bot
```

### Executando com Docker

Para construir a imagem:

```bash
docker build -t elysiabotbot .
```

A execução do contêiner também precisa receber as variáveis de ambiente e ter acesso aos serviços externos necessários, como o Redis.

## Aprendizados e evolução

O ElysiaBotBot é também um registro prático de evolução como desenvolvedor. Seu desenvolvimento permite experimentar decisões de arquitetura, revisar abstrações, refatorar código existente e entender os compromissos envolvidos na manutenção de uma aplicação.

A intenção não é apresentar uma arquitetura perfeita ou definitiva, mas construir uma base consistente e aprender continuamente com as decisões tomadas ao longo do projeto.

## Contribuição

Sugestões, relatos de problemas e ideias de melhoria são bem-vindos. Ao contribuir, procure manter a consistência com a organização e os padrões já adotados no código.

## Licença

Este projeto está disponível sob a licença MIT. Consulte o arquivo `LICENSE` para obter os termos completos.