# GoFlow

GoFlow é um sistema flexível de integração de APIs construído em Go, projetado para criar integrações baseadas em workflows entre diferentes APIs e serviços.

## Configuração

1. Clone o repositório:
```bash
git clone https://github.com/carloskvasir/goflow.git
cd goflow
```

2. Crie um arquivo `.env` na raiz do projeto com as seguintes variáveis:
```bash
GOFLOW_PORT=3000
OPENWEATHER_API_KEY=your_api_key_here
```

Nota: Você pode obter uma chave da API do OpenWeather em https://openweathermap.org/api

3. Execute o servidor:
```bash
go run cmd/main.go
```

## API Endpoints

- `POST /api/v1/workflows`: Registra um novo workflow
- `GET /api/v1/workflows/:id`: Obtém detalhes de um workflow
- `POST /api/v1/workflows/:id/execute`: Executa um workflow
- `DELETE /api/v1/workflows/:id`: Remove um workflow

## Exemplo: Workflow de João Pessoa

O workflow de exemplo em `examples/joao_pessoa_info/workflow.json` demonstra como obter informações sobre João Pessoa:

1. Obtém a hora atual usando a API TimeAPI.io
2. Obtém a temperatura atual usando a API OpenWeatherMap
3. Formata uma mensagem combinando os dados

Para executar o exemplo:

```bash
# Registra o workflow
curl -X POST http://localhost:3000/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d @examples/joao_pessoa_info/workflow.json

# Executa o workflow
curl -X POST http://localhost:3000/api/v1/workflows/joao-pessoa-info/execute
```

## Tipos de Steps

- `rest`: Executa requisições HTTP
- `transform`: Processa e formata dados usando templates
- `echo`: Retorna uma mensagem simples (usado para testes)

## Licença

Este projeto está licenciado sob a Mozilla Public License 2.0 - veja o arquivo [LICENSE](LICENSE) para detalhes.