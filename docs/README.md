 <div align="center">

   <img src="https://capsule-render.vercel.app/api?type=waving&color=0:07130D,45:123D2A,100:2EA66B&height=200&section=header&text=ELYSIA-BOTBOT&fontSize=54&fontColor=E8FFF0&fontAlignY=38&desc=YOUR%20SERVER%2C%20YOUR%20VIBE&descSize=16&descAlignY=60&animation=fadeIn" width="100%" alt="Elysia-BotBot — your server, your vibe" />

   <a href="https://github.com/rickferrdev/elysia-botbot">
     <img src="https://readme-typing-svg.demolab.com?font=JetBrains+Mono&weight=500&size=16&duration=2600&pause=850&color=52D991&center=true&vCenter=true&width=650&lines=Templates+that+bring+your+server+to+life.;Tomatoes,+slippers,+and+a+tiny+shop.;Made+with+Go.+Ready+for+Discord." alt="Templates, a playful tomato economy, and Go-powered Discord tools" />
   </a>

   <p>
     <a href="#features"><img src="https://img.shields.io/badge/EXPLORE-Features-239B56?style=for-the-badge&logo=sparkles&logoColor=white" alt="Explore features"/></a>
     <a href="#tech-stack"><img src="https://img.shields.io/badge/BUILT_WITH-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Built with Go"/></a>
     <a href="#getting-started"><img src="https://img.shields.io/badge/GET_STARTED-Quickstart-276749?style=for-the-badge&logo=rocket&logoColor=white" alt="Get started"/></a>
   </p>

   <p>
     <img src="https://img.shields.io/badge/DISCORD-Bot-5865F2?style=flat-square&logo=discord&logoColor=white" alt="Discord bot"/>
     <img src="https://img.shields.io/badge/REDIS-Cache-DC382D?style=flat-square&logo=redis&logoColor=white" alt="Redis cache"/>
     <img src="https://img.shields.io/badge/SQLITE-Storage-003B57?style=flat-square&logo=sqlite&logoColor=white" alt="SQLite storage"/>
     <img src="https://img.shields.io/badge/TOMATO_AND_SLIPPER-Tiny_Economy-2EA66B?style=flat-square" alt="Tomato and slipper economy"/>
   </p>

</div>

---

# Elysia-BotBot

A Discord bot built with Go as a hands-on learning project focused on software architecture, code organization, and software engineering best practices.

**ElysiaBotBot** brings together features for Discord servers in an application structured to support maintainability, extensibility, and code readability. Beyond building a bot, the goal is to explore how to design and organize a Go backend using tools and patterns found in real-world software projects.

> [!NOTE]
> This project is maintained for educational purposes and practical learning. Version `2.0.0` introduces a breaking update with reusable templates and a new tomato-and-slipper economy.

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
- [x] Reusable message templates.
- [x] Funny commands, including `/humiliator` and `/bestchoice`.
- [x] Tomato and slipper balances, with a shop to exchange 5 slippers for 1 tomato.
- [x] Data persistence with SQLite.
- [x] Redis integration for caching.

### Breaking changes in 2.0.0

- The legacy `/embeds` and `/bestphoto` commands are replaced by `/templates` and `/bestchoice`.
- Data from the legacy `embeds` and `fun_slippers` tables is not migrated. Back up any data you need before upgrading.

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
