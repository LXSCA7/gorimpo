# 🤝 Contributing to GOrimpo

I'm glad you want to help make GOrimpo the best retro deals scraper out there. To keep the project organized, please follow these guidelines:

### Arquitetura
O projeto segue os princípios da Arquitetura Hexagonal. 
- **Core:** Regras de negócio e entidades.
- **Ports:** Interfaces que definem como o sistema se comunica.
- **Adapters:** Implementações reais (Scrapers, Notificadores, Repositórios).

Se for adicionar um novo scraper, ele deve implementar a interface `ports.Scraper` dentro de uma nova pasta em `internal/adapters/scraper/<seu-scraper>`.

### Architecture

The project follows Hexagonal Architecture principles.

- **Core**: Business rules and entities.
- **Ports**: Interfaces defining how the system communicates.
- **Adapters**: Real implementations (Scrapers, Notifiers, Repositories).

If you are adding a new scraper, it must implement the ports.Scraper interface within a new folder in `internal/adapters/scraper/<your-scraper>`.

### Workflow
1. Create an Issue detailing what you intend to do;
2. Fork the project;
3. Create your feature branch:
```bash
git checkout -b feat/my-new-feature
```
4. Commit your changes following the Conventional Commits standard;
5. Open a Pull Request explaining what was done.

### Testing and Quality

Before opening a PR, ensure that:
- The code compiles without errors `go build ./...`.
- You have run `go fmt ./...` to format the code.
- The Dockerfile was tested if any infrastructure changes were made

---

*Questions? Open an Issue and let's talk!*