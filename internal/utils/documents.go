package utils

import (
	"encoding/json"
	"os"

	"github.com/jcserv/portfolio-api/internal/model"
)

func ReadExperience() ([]model.Experience, error) {
	file, err := os.Open("dist/experience.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var experiences []model.Experience
	if err := json.NewDecoder(file).Decode(&experiences); err != nil {
		return nil, err
	}
	return experiences, nil
}

func ReadProjects() ([]model.Project, error) {
	file, err := os.Open("dist/projects.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var projects []model.Project
	if err := json.NewDecoder(file).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}
