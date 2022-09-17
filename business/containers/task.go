package containers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/pkg/errors"
)

type ClientModel interface {
	ListTaskDefinitionFamilies(ctx context.Context, params *ecs.ListTaskDefinitionFamiliesInput, optFns ...func(*ecs.Options)) (*ecs.ListTaskDefinitionFamiliesOutput, error)
	ListTasks(ctx context.Context, params *ecs.ListTasksInput, optFns ...func(*ecs.Options)) (*ecs.ListTasksOutput, error)
	DescribeTasks(ctx context.Context, params *ecs.DescribeTasksInput, optFns ...func(*ecs.Options)) (*ecs.DescribeTasksOutput, error)
}

type cluster struct {
	Name   string // name of the cluster to run on
	client ClientModel
}

// List returns a list of running tasks
func (C cluster) List() ([]taskInstance, error) {
	var tasks []taskInstance
	result, err := C.client.ListTaskDefinitionFamilies(context.TODO(), &ecs.ListTaskDefinitionFamiliesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "listing taskInstance families")
	}

	for _, family := range result.Families {
		task, err := C.GetTask(family)
		if err != nil {
			continue
		}
		tasks = append(tasks, *task)
	}

	return tasks, nil
}

func (C cluster) GetTask(family string) (*taskInstance, error) {
	tasksResult, err := C.client.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Family:  &family,
		Cluster: &C.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing tasks for family")
	}

	if len(tasksResult.TaskArns) > 1 {
		return nil, errors.New(fmt.Sprintf("found %d tasks for server '%s'", len(tasksResult.TaskArns), family))
	}

	task := taskInstance{
		Name: family,
	}

	if len(tasksResult.TaskArns) == 0 {
		return nil, nil
	}

	task.taskARN = tasksResult.TaskArns[0]

	taskResult, err := C.client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Cluster: &C.Name,
		Tasks:   []string{tasksResult.TaskArns[0]},
	})

	if err != nil || len(taskResult.Tasks) == 0 {
		return nil, errors.Wrap(err, "describing task")

	}

	task.Status = *taskResult.Tasks[0].DesiredStatus

	for _, container := range taskResult.Tasks[0].Containers {
		for _, binding := range container.NetworkBindings {
			if *binding.BindIP != "" {
				task.IP = *binding.BindIP
				break
			}
		}
	}

	return &task, nil
}

// taskInstance represents an ECS task
type taskInstance struct {
	Name    string // name of the task
	Status  string // status of the task
	taskARN string // where the ECS task is running
	IP      string // what IP the task is running at
}

func (t taskInstance) Start() error {
	// register taskInstance definition as new version
	// start taskInstance
	// store t.taskARN
	return nil
}

func (t taskInstance) Stop() error {
	// stop taskInstance
	return nil
}
