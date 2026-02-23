# MiniTwit UML Diagrams

PlantUML-diagrammer der dokumenterer MiniTwit-systemets arkitektur på forskellige abstraktionsniveauer.

## Diagrammer

### Context Diagram (`context-formal.puml`)
Viser systemet udefra: tre aktører (User, Simulator, Developer) og deres kommunikation med MiniTwit via HTTP/HTML, HTTP/JSON (Basic Auth) og SSH/Git. Indeholder interne komponenter (Go Web Server, SQLite Database).

### Deployment Diagram (`deployment.puml`)
Viser den fysiske infrastruktur: Hetzner VM (Ubuntu 22.04) med Docker Engine, container med Go Web Server og SQLite, samt et Docker Volume til data-persistering.

### Component Diagram (`component.puml`)
Zoomer ind i Go Web Serveren og viser interne komponenter: Router (gorilla/mux), Web Handlers, API Handlers, Session Manager, HTML Templates og Database Layer, samt deres indbyrdes forbindelser.

## Hvordan man ser diagrammerne

Brug PlantUML preview i VS Code (Alt+D) eller generer PNG/SVG:

```bash
plantuml diagrams/*.puml
```

## Næste iteration

- **Sequence diagram** - vis runtime-flows som "bruger poster en besked" eller "simulator henter beskeder"
- **Class diagram** - vis Go structs (User, Message) og deres relationer
