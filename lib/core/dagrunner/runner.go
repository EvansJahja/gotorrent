package dagrunner

import (
	"fmt"
	"reflect"
	"sync"

	"example.com/gotorrent/lib/core/dag"
)

type DagRun interface {
}

func Walk(base dag.Task) DagRun {
	d := dagRunImpl{}

	tasks := base.Run()

	var wg sync.WaitGroup
	for dependent := range tasks {
		fmt.Printf("%s depends on %s\n", nodeName(base), nodeName(dependent))
		wg.Add(1)
		go func(dependent dag.Task) {
			Walk(dependent)
			wg.Done()
		}(dependent)
	}
	wg.Wait()
	/*
		d.q = append(d.q, base)
		d.visited = make(map[dag.Task]struct{})
		d.visited[base] = yes
		d.walk()
	*/
	return d
}

type dagRunImpl struct {
	q       []dag.Task
	visited map[dag.Task]struct{}
}

/*
func (d dagRunImpl) walk() {
	if len(d.q) == 0 {
		return
	}

	for len(d.q) > 0 {
		q := d.q[0]
		d.q = d.q[1:]
		for r := range q.Run() {
			if _, isVisited := d.visited[r]; isVisited {
				continue
			}
			fmt.Printf("%s depends on %s\n", nodeName(q), nodeName(r))
			d.q = append(d.q, r)
			d.visited[r] = yes
		}
	}

}
*/

func nodeName(n dag.Task) string {
	if x, ok := n.(fmt.Stringer); ok {
		return x.String()
	}
	t := reflect.TypeOf(n)
	return t.Name()
}

/*
func Walk(base dag.Node) DagRun {
	d := dagRunImpl{}
	d.q = append(d.q, base)
	d.visited = make(map[dag.Node]struct{})
	d.visited[base] = yes
	d.walk()
	return d
}

type dagRunImpl struct {
	q       []dag.Node
	visited map[dag.Node]struct{}
}

func (d dagRunImpl) walk() {
	if len(d.q) == 0 {
		return
	}

	for len(d.q) > 0 {
		q := d.q[0]
		d.q = d.q[1:]
		for _, r := range q.Start() {
			if _, isVisited := d.visited[r]; isVisited {
				continue
			}
			fmt.Printf("%s depends on %s\n", nodeName(q), nodeName(r))
			d.q = append(d.q, r)
			d.visited[r] = yes
		}
	}

}

func nodeName(n dag.Node) string {
	if x, ok := n.(fmt.Stringer); ok {
		return x.String()
	}
	t := reflect.TypeOf(n)
	return t.Name()
}
*/
