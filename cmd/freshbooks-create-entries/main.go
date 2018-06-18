package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/scoiatael/gofreshbooks"
)

type Date struct {
	time.Time
}

func (date *Date) MarshalCSV() (string, error) {
	return date.Time.Format("2006-01-02"), nil
}

func (date *Date) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02", csv)
	if err != nil {
		return err
	}
	return nil
}

type TimeEntry struct {
	Id      string  `csv:"id"`
	Project string  `csv:"project_name"`
	Task    string  `csv:"task_name"`
	Hours   float64 `csv:"hours"`
	Notes   string  `csv:"notes"`
	Date    Date    `csv:"date"`
}

func LoadEntries() (entries []TimeEntry, err error) {
	entriesFile, maybeErr := os.OpenFile("entries.csv", os.O_RDWR, os.ModePerm)
	if maybeErr != nil {
		err = maybeErr
		return
	}
	defer entriesFile.Close()

	err = gocsv.UnmarshalFile(entriesFile, &entries)
	return
}

type ProjectName string
type ProjectMap map[ProjectName]string

func projectMap(projectList freshbooks.ProjectList) (ret ProjectMap) {
	ret = make(ProjectMap)
	for _, project := range projectList.Projects {
		ret[ProjectName(project.Name)] = project.ID
	}
	return
}

type TaskName string
type TaskMap map[TaskName]string

func taskMap(taskList freshbooks.TaskList) (ret TaskMap) {
	ret = make(TaskMap)
	for _, task := range taskList.Tasks {
		ret[TaskName(task.Name)] = task.ID
	}
	return
}

func Create(entries []TimeEntry, projects ProjectMap, tasks TaskMap) error {
	for _, entry := range entries {
		request := struct {
			XMLName   xml.Name             `xml:"request"`
			Method    string               `xml:"method,attr"`
			TimeEntry freshbooks.TimeEntry `xml:"time_entry"`
		}{
			Method: "time_entry.create",
			TimeEntry: freshbooks.TimeEntry{
				ProjectId: projects[ProjectName(entry.Project)],
				TaskId:    tasks[TaskName(entry.Task)],
				Hours:     entry.Hours,
				Notes:     entry.Notes,
				Date:      freshbooks.Date{entry.Date.Time},
			},
		}
		// out, err := xml.MarshalIndent(request, " ", "  ")
		// fmt.Printf("%v\n%s\n", err, out)
		response, err := freshbooks.Do(request)
		if err != nil {
			return err
		}
		fmt.Printf("%+v\n%+s\n", err, response)
	}
	return nil
}

func main() {
	projectList, err := freshbooks.ListProjects()
	if err != nil {
		fmt.Errorf("error: %v\n", err)
	}
	taskList, err := freshbooks.ListTasks()
	if err != nil {
		fmt.Errorf("error: %v\n", err)
	}
	entries, err := LoadEntries()
	if err != nil {
		fmt.Errorf("error: %v\n", err)
	}
	if err := Create(entries, projectMap(projectList), taskMap(taskList)); err != nil {
		fmt.Errorf("error: %v\n", err)
	}
}
