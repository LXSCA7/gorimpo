# 🤝 Contribuindo para o GOrimpo

Fico feliz que você quer ajudar a tornar o GOrimpo o melhor garimpeiro de ofertas retrô. Para manter a casa organizada, peço que siga estas diretrizes:

### Arquitetura
O projeto segue os princípios da Arquitetura Hexagonal. 
- **Core:** Regras de negócio e entidades.
- **Ports:** Interfaces que definem como o sistema se comunica.
- **Adapters:** Implementações reais (Scrapers, Notificadores, Repositórios).

Se for adicionar um novo scraper, ele deve implementar a interface `ports.Scraper` dentro de uma nova pasta em `internal/adapters/scraper/<seu-scraper>`.

### Fluxo de Trabalho
1. Crie uma Issue detalhando o que pretende fazer;
2. Faça um Fork do projeto;
3. Crie sua branch de feature: 
```bash
git checkout -b feat/minha-nova-feature
```
4. Commit suas mudanças seguindo o padrão Conventional Commits;
5. Abra um Pull Request detalhando o que foi feito.

### Testes e Qualidade
Antes de abrir o PR, garanta que:
- O código compila sem erros (`go build ./...`).
- Você rodou o `go fmt ./...` para formatar o código.
- O Dockerfile foi testado caso tenha alterado algo de infra.

---

*Dúvidas? Abra uma Issue e vamos trocar uma ideia!*