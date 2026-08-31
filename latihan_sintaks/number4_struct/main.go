package main

import "fmt"

type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

func (s Student) GetInfo() string {
	return fmt.Sprintf("[%d] %s - Nilai: %.1f - Status: %v", s.ID, s.Name, s.Grade, s.IsActive)
}

func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

func (s *Student) Activate() {
	s.IsActive = true
}

func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	mhs := Student{ID: 101, Name: "Diaul", Grade: 80.0, IsActive: false}
	fmt.Println(mhs.GetInfo())

	mhs.Activate()
	fmt.Println(mhs.GetInfo())

	mhs.UpdateGrade(94.0)
	fmt.Println(mhs.GetInfo())

	mhs.Deactivate()
	fmt.Println(mhs.GetInfo())
}