 <div align="center">

  <img src="https://capsule-render.vercel.app/api?type=waving&color=0:07130D,50:123D2A,100:2EA66B&height=180&section=header&text=Elysia-BotBot&fontSize=52&fontColor=E8FFF0&fontAlignY=38&desc=Discord%20Bot%20%E2%80%A2%20Built%20with%20Go&descSize=17&descAlignY=60&animation=fadeIn" width="100%" alt="ElysiaBotBot — Discord Bot built with Go" />

  <a href="https://github.com/rickferrdev/elysia-botbot">
    <img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=500&size=16&duration=2800&pause=900&color=52D991&center=true&vCenter=true&width=600&lines=Building+more+than+a+Discord+bot.;Learning+Go+through+real+software.;Clean+code.+Thoughtful+architecture.;One+commit+at+a+time." alt="Animated messages about the project" />
  </a>

  <p>
    <strong>A learning project with a product-oriented mindset.</strong><br/>
    Exploring backend architecture, software engineering practices, and Go development.
  </p>

  <p>
    <a href="#-features"><img src="https://img.shields.io/badge/Features-Explore-239B56?style=for-the-badge" alt="Explore features"/></a>
    <a href="#-tech-stack"><img src="https://img.shields.io/badge/Stack-Go%20%7C%20Redis%20%7C%20SQLite-183D2B?style=for-the-badge" alt="Technology stack"/></a>
    <a href="#-getting-started"><img src="https://img.shields.io/badge/Getting_Started-Quickstart-276749?style=for-the-badge" alt="Getting started"/></a>
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

A Discord bot built with Go as a hands-on learning project focused on software architecture, code organization, and software engineering best practices.

**ElysiaBotBot** brings together features for Discord servers in an application structured to support maintainability, extensibility, and code readability. Beyond building a bot, the goal is to explore how to design and organize a Go backend using tools and patterns found in real-world software projects.

> [!NOTE]
> This project is maintained for educational purposes and practical learning. Version `1.0.0` marks a milestone following stabilization, bug fixes, and refactoring.

## Goals

- Build a Discord bot using Go.
- Explore separation of concerns and layered architecture.
- Practice dependency injection and component composition.
- Work with data persistence and caching.
- Maintain a codebase that is understandable, testable, and ready for future improvements.
- Consolidate knowledge through a practical project rather than isolated examples.

## Features

The project brings together features for configuring and customizing Discord servers.

- [x] Discord command and event handling structure.
- [x] Server-specific configuration.
- [x] Auto-reaction system.
- [x] Embed support.
- [x] Data persistence with SQLite.
- [x] Redis integration for caching.

Feature availability and behavior depend on the implementation included in the published release.

## Tech Stack

| Technology | Purpose |
| --- | --- |
| [Go](https://go.dev/) | Primary programming language |
| [discord.go](https://github.com/discord-go/discord.go) | Discord API integration |
| [SQLite](https://www.sqlite.org/) | Data persistence |
| [Bun](https://bun.uptrace.dev/) | Database access |
| [Redis](https://redis.io/) | Caching |
| [Uber Fx](https://github.com/uber-go/fx) | Dependency injection and application lifecycle management |
| [Docker](https://www.docker.com/) | Containerization and deployment |

## Getting Started

### Prerequisites

To run the project locally, you will need:

- [Go](https://go.dev/dl/)
- A Discord application and bot configured in the [Discord Developer Portal](https://discord.com/developers/applications)
- A running [Redis](https://redis.io/) instance
- Docker, if you prefer to run the application in a container

### Configuration

Clone the repository and navigate to the project directory:

```bash
git clone https://github.com/rickferrdev/elysia-botbot.git
cd elysia-botbot
```

Configure the environment variables required by the application, including the bot token, database connection settings, and Redis credentials.

Use `.env.example` as a reference, and verify that the variable names match the configuration of the version you intend to run. **Never expose your Discord bot token or commit credentials to the repository.**

### Running Locally

Once the dependencies are configured, run:

```bash
go mod download
go run ./cmd/bot
```

### Running with Docker

Build the Docker image:

```bash
docker build -t elysiabotbot .
```

When running the container, provide the required environment variables and ensure it can access external services such as Redis.

## Learning and Continuous Improvement

ElysiaBotBot is also a practical record of growth as a developer. Its development provides an opportunity to experiment with architectural decisions, revisit abstractions, refactor existing code, and understand the trade-offs involved in maintaining an application.

The goal is not to claim a perfect or definitive architecture, but to build a consistent foundation and keep learning through the process.

## Contributing

Suggestions, bug reports, and ideas for improvement are welcome. When contributing, please follow the project's existing structure and coding conventions.

## License

This project is licensed under the MIT License. See the `LICENSE` file for the full license text.
