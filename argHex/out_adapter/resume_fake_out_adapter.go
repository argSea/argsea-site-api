package out_adapter

import (
	"fmt"
	"time"

	"github.com/argSea/portfolio_blog_api/argHex/domain"
	"github.com/argSea/portfolio_blog_api/argHex/out_port"
)

type resumeFakeOutAdapter struct {
}

func NewResumeFakeOutAdapter() out_port.ResumeRepo {
	return resumeFakeOutAdapter{}
}

func (u resumeFakeOutAdapter) GetByUserID(id string) (domain.Resumes, int64, error) {
	resume := domain.Resume{}
	college := domain.CollegeExperience{}
	exp := domain.Experience{}
	courses := domain.Course{}
	skills := domain.SkillSection{}

	college.CumulativeGPA = 4.0
	college.Degree = "Masters of Software Engineering"
	college.EndDate.Year = 2016
	college.EndDate.Month = time.August
	college.StartDate.Year = 2014
	college.StartDate.Month = time.August
	college.Highlights = append(college.Highlights, domain.Highlight{Description: "Some highlight", Visible: true})
	college.Major = "Software Engineering"
	college.MajorGPA = 4.0
	college.School.Name = "Penn State"

	exp.StartDate.Year = 2016
	exp.StartDate.Month = time.May
	exp.Work.Name = "Pittsburgh Post-Gazette"
	exp.Title = "Systems Architect"
	exp.Highlights = append(exp.Highlights, domain.Highlight{Description: "Did something", Visible: true})

	courses.Institution.Name = "Udemy"
	courses.Name = "Blender BS"

	skills.Name = "Programming Languages"
	skills.Skills = append(skills.Skills, domain.Skill{Name: "C++"})

	resume.Id = "not"
	resume.UserID = "cabbage"
	resume.Education = append(resume.Education, college)
	resume.Experiences = append(resume.Experiences, exp)
	resume.About = "I'm me"
	resume.ExtraCourses = append(resume.ExtraCourses, courses)
	resume.SkillSections = append(resume.SkillSections, skills)

	fmt.Println(resume)

	resumes := []domain.Resume{}
	resumes = append(resumes, resume)

	return resumes, 1, nil
}

func (res resumeFakeOutAdapter) Get(id string) domain.Resume {
	return domain.Resume{}
}

func (res resumeFakeOutAdapter) Set(resume domain.Resume) error {
	return nil
}

func (res resumeFakeOutAdapter) Add(resume domain.Resume) (string, error) {
	return "", nil
}

func (res resumeFakeOutAdapter) Remove(resume domain.Resume) error {
	return nil
}
