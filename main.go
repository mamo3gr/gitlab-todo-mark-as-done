package main

import (
	"log"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func main() {
	token, ok := os.LookupEnv("GITLAB_TOKEN")
	if !ok {
		log.Fatalf("GITLAB_TOKEN is not set")
	}
	baseURL, ok := os.LookupEnv("GITLAB_BASEURL")
	if !ok {
		log.Fatalf("GITLAB_BASEURL is not set")
	}

	git, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		log.Fatalf("Failed to create client. error=%v", err)
	}

	todos, err := listTodos(git)
	if err != nil {
		log.Fatalf("Failed to list todos. error=%v", err)
	}

	marked := 0
	left := 0
	for _, todo := range todos {
		if isMRMergedOrClosed(todo) {
			_, err := git.Todos.MarkTodoAsDone(todo.ID)
			if err != nil {
				log.Printf("Failed to mark as done. id=%d,error=%v", todo.ID, err)
				continue
			}
			marked += 1
			log.Printf("Marked as MR is merged or closed. id=%d,target=%s", todo.ID, todo.TargetURL)
			continue
		}

		log.Printf("Left. id=%d,target=%s", todo.ID, todo.TargetURL)
		left += 1
	}
	log.Printf("Finished. found=%d,marked=%d,left=%d", len(todos), marked, left)
}

func isMRMergedOrClosed(todo *gitlab.Todo) bool {
	if todo.TargetType != "MergeRequest" {
		return false
	}

	return todo.Target.State == "merged" || todo.Target.State == "closed"
}

func listTodos(git *gitlab.Client) ([]*gitlab.Todo, error) {
	opt := &gitlab.ListTodosOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
			Page:    1,
		},
		State: gitlab.Ptr("pending"),
	}

	var allTodos []*gitlab.Todo

	for {
		todos, resp, err := git.Todos.ListTodos(opt)
		if err != nil {
			return nil, err
		}

		allTodos = append(allTodos, todos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allTodos, nil
}
