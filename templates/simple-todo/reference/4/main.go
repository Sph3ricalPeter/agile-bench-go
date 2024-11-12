package main

import (
	"fmt"

	"github.com/google/uuid"
)

type Todo struct {
	id     string
	text   string
	isDone bool
}

type Storage struct {
	todos []Todo
}

func (s *Storage) addTodo(text string) {
	s.todos = append(s.todos, Todo{
		id:     uuid.New().String(),
		text:   text,
		isDone: false,
	})
}

func (s *Storage) getTodos() []Todo {
	return s.todos
}

func (s *Storage) removeTodo(id string) {
	for i, todo := range s.todos {
		if todo.id == id {
			s.todos = append(s.todos[:i], s.todos[i+1:]...)
			break
		}
	}
}

func (s *Storage) markTodoAsDone(id string) {
	for i, todo := range s.todos {
		if todo.id == id {
			s.todos[i].isDone = true
			break
		}
	}
}

func main() {
	fmt.Println(getMessage())

	storage := Storage{}

	storage.addTodo("Buy milk")
	storage.addTodo("Do laundry")
	fmt.Println(storage.getTodos())

	storage.markTodoAsDone(storage.getTodos()[0].id)
	fmt.Println(storage.getTodos())

	storage.removeTodo("Do laundry")
	fmt.Println(storage.getTodos())
}

func getMessage() string {
	return "Welcome to a TODO list!"
}
