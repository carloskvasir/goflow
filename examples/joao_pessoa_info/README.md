# João Pessoa Info Workflow

Este é um exemplo de workflow que obtém a hora atual e a temperatura em João Pessoa, Brasil.

## Funcionalidades

- Obtém a hora atual usando a WorldTime API
- Obtém a temperatura atual usando a OpenWeather API
- Formata uma mensagem amigável com as informações

## Pré-requisitos

1. Go 1.x instalado
2. Uma API key do OpenWeather (obtenha em https://openweathermap.org/api)

## Como usar

1. Configure sua API key do OpenWeather:
```bash
export OPENWEATHER_API_KEY=sua_api_key_aqui
```

2. Execute o exemplo:
```bash
go run main.go
```

## Estrutura do Workflow

O workflow consiste em três steps:

1. `get-time`: Obtém a hora atual
2. `get-weather`: Obtém a temperatura atual
3. `format-message`: Formata os dados em uma mensagem amigável

## Exemplo de Saída

```
Hello from João Pessoa! Current time is 14:30 and temperature is 28°C
```

## Configuração

Você pode modificar o workflow editando o arquivo `workflow.json`. Algumas opções que você pode personalizar:

- Template da mensagem
- Configurações de retry
- Parâmetros das APIs
