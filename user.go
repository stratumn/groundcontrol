package groundcontrol

import "container/list"

var Viewer = User{}

type User struct {
	ID         string      `json:"id"`
	Workspaces []Workspace `json:"workspaces"`
}

func (u *User) Workspace(slug string) *Workspace {
	for _, v := range u.Workspaces {
		if v.Slug == slug {
			return &v
		}
	}

	return nil
}

func (u *User) Jobs(
	after *string,
	before *string,
	first *int,
	last *int,
	status []JobStatus,
) (*JobConnection, error) {
	jobList := list.New()
	element := GetJobList().Front()

	for element != nil {
		job := element.Value.(*Job)
		match := len(status) == 0

		for _, v := range status {
			if job.Status == v {
				match = true
				break
			}
		}

		if match {
			jobList.PushBack(job)
		}

		element = element.Next()
	}

	connection, err := jobPaginator.Paginate(jobList, after, before, first, last)
	if err != nil {
		return nil, err
	}

	edges := make([]JobEdge, len(connection.Edges))

	for i, v := range connection.Edges {
		edges[i] = JobEdge{
			Node:   v.Node.(*Job),
			Cursor: v.Cursor,
		}
	}

	return &JobConnection{
		Edges:    edges,
		PageInfo: connection.PageInfo,
	}, nil
}
