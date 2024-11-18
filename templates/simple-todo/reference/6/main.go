package main

import (
	"fmt"
	"io"

	"github.com/google/uuid"
)

type Status string

const (
	Open  Status = "open"
	Doing Status = "doing"
	Done  Status = "done"
)

type Todo struct {
	id     string
	text   string
	status Status
}

type Storage struct {
	todos []Todo
}

func (s *Storage) addTodo(text string) {
	s.todos = append(s.todos, Todo{
		id:     uuid.New().String(),
		text:   text,
		status: Open,
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

func (s *Storage) setTodoStatus(id string, status Status) {
	for i, todo := range s.todos {
		if todo.id == id {
			s.todos[i].status = status
			break
		}
	}
}

func (s *Storage) getTodosByStatus(status Status) []Todo {
	var todos []Todo
	for _, todo := range s.todos {
		if todo.status == status {
			todos = append(todos, todo)
		}
	}
	return todos
}

func (s *Storage) writeTodos(w io.Writer) {
	for _, todo := range s.todos {
		fmt.Fprintf(w, "%s: %s (%s)\n", todo.id, todo.text, todo.status)
	}
}

func main() {
	fmt.Println(getMessage())

	storage := Storage{}

	storage.addTodo("Buy milk")
	storage.addTodo("Do laundry")
	fmt.Println(storage.getTodos())

	storage.setTodoStatus(storage.getTodos()[0].id, Done)
	fmt.Println(storage.getTodos())

	storage.removeTodo("Do laundry")
	fmt.Println(storage.getTodos())
}

func getMessage() string {
	return "Welcome to a TODO list!"
}
