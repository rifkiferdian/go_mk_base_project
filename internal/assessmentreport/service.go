package assessmentreport

import "context"

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Report(ctx context.Context) (Report, error) {
	return s.ReportWithFilter(ctx, Filter{})
}

func (s *Service) ReportWithFilter(ctx context.Context, filter Filter) (Report, error) {
	var report Report
	var err error
	if report.Summary, err = s.repository.Summary(ctx, filter); err != nil {
		return report, err
	}
	counts, err := s.repository.Distribution(ctx, filter)
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
	if report.Aspects, err = s.repository.Aspects(ctx, filter); err != nil {
		return report, err
	}
	if report.Stores, err = s.repository.Stores(ctx, filter); err != nil {
		return report, err
	}
	if report.Departments, err = s.repository.Departments(ctx, filter); err != nil {
		return report, err
	}
	if report.Periods, err = s.repository.Periods(ctx, filter); err != nil {
		return report, err
	}
	if report.Recent, err = s.repository.Recent(ctx, filter); err != nil {
		return report, err
	}
	return report, nil
}

func (s *Service) FilterOptions(ctx context.Context) ([]DepartmentOption, []RubricOption, error) {
	return s.repository.FilterOptions(ctx)
}
