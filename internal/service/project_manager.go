package service

import (
	"os"
	"path/filepath"
	"time"
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}

type ProjectManager struct {
	projects map[string]Project
}

func NewProjectManager() *ProjectManager {
	return &ProjectManager{
		projects: make(map[string]Project),
	}
}

func (pm *ProjectManager) CreateProject(name, path, description string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", os.ErrInvalid
	}

	project := Project{
		ID:          generateID(),
		Name:        name,
		Path:        path,
		Description: description,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
	}

	pm.projects[project.ID] = project
	return project.ID, nil
}

func (pm *ProjectManager) GetProject(id string) (*Project, error) {
	project, exists := pm.projects[id]
	if !exists {
		return nil, os.ErrNotExist
	}
	return &project, nil
}

func (pm *ProjectManager) ListProjects() []Project {
	projects := make([]Project, 0, len(pm.projects))
	for _, p := range pm.projects {
		projects = append(projects, p)
	}
	return projects
}

func (pm *ProjectManager) UpdateProject(id, name, description string) error {
	project, exists := pm.projects[id]
	if !exists {
		return os.ErrNotExist
	}

	if name != "" {
		project.Name = name
	}
	if description != "" {
		project.Description = description
	}
	project.ModifiedAt = time.Now()

	pm.projects[id] = project
	return nil
}

func (pm *ProjectManager) DeleteProject(id string) error {
	_, exists := pm.projects[id]
	if !exists {
		return os.ErrNotExist
	}
	delete(pm.projects, id)
	return nil
}

func (pm *ProjectManager) GetProjectFiles(path string) ([]string, error) {
	var files []string
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, filePath)
		return nil
	})
	return files, err
}

func generateID() string {
	return time.Now().Format("20060102150405")
}