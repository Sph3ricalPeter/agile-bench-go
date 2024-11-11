# Go-Agile Bench
LLM benchmark for agile software development in Golang via functional requirements using TDD approach for evaluation.

## Procedures

1. patch procedure - init folder, then for each req. send a prompt, parse patch file from response & apply it, evaluate by checking fail->pass tests. Optionally use history for chat-like conversation.
2. file procedure - init folder, then for each req. send prompt, parse individual full files & save them, evaluate by checking fail-pass tests.

TODO: put the benchmark files into a .gz to avoid GitHub using it for training :)