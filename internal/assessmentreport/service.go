package assessmentreport

import "context"

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Report(ctx context.Context) (Report, error) {
	var report Report
	var err error
	if report.Summary, err = s.repository.Summary(ctx); err != nil {
		return report, err
	}
	counts, err := s.repository.Distribution(ctx)
	if err != nil {
		return report, err
	}
	labels := []string{"Sesuai Standar", "Mendekati Standar", "Kurang", "Kurang Sekali", "Sangat Kurang Sekali"}
	colors := []string{"emerald", "orange", "amber", "rose", "slate"}
	offset := 0.0
	for i, label := range labels {
		percentage := 0.0
		if report.Summary.AssessmentCount > 0 {
			percentage = float64(counts[label]) * 100 / float64(report.Summary.AssessmentCount)
		}
		report.Distribution = append(report.Distribution, Distribution{Label: label, Count: counts[label], Percentage: percentage, Offset: -offset, Color: colors[i]})
		offset += percentage
	}
	if report.Aspects, err = s.repository.Aspects(ctx); err != nil {
		return report, err
	}
	if report.Stores, err = s.repository.Stores(ctx); err != nil {
		return report, err
	}
	if report.Departments, err = s.repository.Departments(ctx); err != nil {
		return report, err
	}
	if report.Recent, err = s.repository.Recent(ctx); err != nil {
		return report, err
	}
	return report, nil
}
