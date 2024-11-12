package main

import "fmt"

type Storage struct {
	todos []string
}

func (s *Storage) addTodo(text string) {
	s.todos = append(s.todos, text)
}

func (s *Storage) getTodos() []string {
	return s.todos
}

func (s *Storage) removeTodo(text string) {
	for i, todo := range s.todos {
		if todo == text {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			break
		}
	}
}

func main() {
	fmt.Println(getMessage())

	storage := Storage{}

	storage.addTodo("Buy milk")
	storage.getTodos()
}

func getMessage() string {
	return "This will be a TODO list!"
}
