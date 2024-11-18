package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Status string

const (
	Open  Status = "open"
	Doing Status = "doing"
	Done  Status = "done"
)

type IdGenerator interface {
	NewId() string
}

type UUIDGenerator struct{}

func (u *UUIDGenerator) NewId() string {
	return uuid.New().String()[:8]
}

type SeqIdGenerator struct {
	seq int
}

func (s *SeqIdGenerator) NewId() string {
	s.seq++
	return fmt.Sprintf("%d", s.seq)
}

type Todo struct {
	id     string
	text   string
	status Status
}

type Storage struct {
	idGen IdGenerator
	todos []Todo
}

func NewStorage(idGen IdGenerator) *Storage {
	return &Storage{
		idGen: idGen,
	}
}

func (s *Storage) addTodo(text string) {
	s.todos = append(s.todos, Todo{
		id:     s.idGen.NewId(),
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
	idGen := SeqIdGenerator{}
	s := NewStorage(&idGen)

	inChan := make(chan string)
	defer close(inChan)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for {
			s.Scan()
			inChan <- s.Text()
		}
	}()

	for {
		select {
		case input := <-inChan:
			parts := strings.Split(input, " ")
			if len(parts) == 0 || len(parts) > 2 {
				fmt.Println("Invalid input")
				continue
			}

			if len(parts) == 1 {
				if parts[0] == "exit" {
					os.Exit(0)
				}
				if parts[0] == "list" {
					s.writeTodos(os.Stdout)
					continue
				}
			}

			if parts[0] == "add" {
				s.addTodo(parts[1])
				continue
			}
			if parts[0] == "remove" {
				s.removeTodo(parts[1])
				continue
			}

			fmt.Println("Invalid input")
		}
	}
}
