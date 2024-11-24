# Go-Agile Bench
LLM benchmark for agile software development in Golang via functional requirements using TDD approach for evaluation.

## Usage

```bash
go run main.go -t=functions,simple-website,simple-todo,url-shortener -s -e=pass-k -k=1 -T=0.2
```

## Procedures

1. patch procedure - init folder, then for each req. send a prompt, parse patch file from response & apply it, evaluate by checking fail->pass tests. Optionally use history for chat-like conversation.
2. file procedure - init folder, then for each req. send prompt, parse individual full files & save them, evaluate by checking fail-pass tests.

TODO: put the benchmark files into a .gz to avoid GitHub using it for training :)

## Projects

### Simple
1. functions
    - similar to HumanEval
    - tests whether a the model can implement a single requirement in the form of a function
    - ranging from elementary school math to complex algorithms with recursion
    - possibly different variants of the same function
2. (functions-2)
    - more difficult problems across the board than functions

### Std library
3. text manipulation util
    - small CLI tool using only the std library
    - single file application, using only the std library
4. coingecko api proxy
    - small service app which periodically queries CoinGecko for some data and provides its own API
    - in-memory storage
    - single file application, using the std library

### 3rd party libs / DB
5. url shortener api
    - small service app which allows users to shorten URLs
    - can be single file, use in-memory storage and transition to SQLite, validates URLs
6. expense tracker api
    - medium size service to track user's spending
    - simple auth (basic or session), Chi, SQLite & gorm
    - UML schema as input
7. (openai.com/api/pricing scraper) EXPERIMENTAL
    - medium size web scraper to track OpenAI model prices
    - questionable if it's gonna be able to write it, potentially unstable tests

### More complex app (many technologies, concurrency, specialized problems)
7. blog platform website
    - medium size web app which allows users to post content, share it, read it and comment on it
    - SQLite / PostgreSQL, Chi/Gin, session auth 
    - full stack, frontend done with html/template, (optionally CDN version of Bootstrap)
    - input validation, sanitization
    - UML schema for DB, maybe OpenAPI spec for REST endpoints
8. concurrent blog content indexer
    - build inverted indexing for search
    - medium size concurrent app using goroutines & channels, worker pools
    - DB query throttling, logging
9. live blog updates streaming via websockets
    - pub/sub channels